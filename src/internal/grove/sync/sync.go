package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// SyncOrphanFromTrunk merges changes from the trunk (for the specific repo path) into the current orphan branch.
// It uses the sticky trunk context or explicit trunk name to determine the source.
func SyncOrphanFromTrunk(rootPath, currentBranch, trunkBranch, repoName string) error {
	// 0. Resolve Current Branch (if missing)
	if currentBranch == "" {
		cb, err := gitUtil.CurrentBranch(rootPath)
		if err != nil {
			return fmt.Errorf("failed to determine current branch: %w", err)
		}
		currentBranch = cb
	}

	// 1. Identify Source Trunk
	targetTrunk := trunkBranch
	if targetTrunk == "" {
		stickyTrunk, err := groveUtil.GetContextTrunk(rootPath)
		if err != nil || stickyTrunk == "" {
			return fmt.Errorf("unknown trunk branch. Please checkout repo again from TUI to set context")
		}
		targetTrunk = stickyTrunk
	}

	// 2. Load Config from Trunk to find Repo Path
	config, err := groveUtil.LoadConfigFromGitRef(rootPath, targetTrunk)
	if err != nil {
		return fmt.Errorf("failed to load config from trunk '%s': %w", targetTrunk, err)
	}

	repoConfig, exists := config.Repositories[repoName]
	if !exists {
		return fmt.Errorf("repository '%s' not found in trunk configuration", repoName)
	}

	repoRelPath := repoConfig.Path

	// Pre-flight checks to provide better errors
	// 1. Verify trunk exists
	if _, err := gitUtil.RepoRoot(); err == nil { // quick check if git repo
		cmd := exec.Command("git", "rev-parse", "--verify", targetTrunk)
		cmd.Dir = rootPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("trunk branch '%s' does not exist locally. Try 'git fetch origin %s:%s'", targetTrunk, targetTrunk, targetTrunk)
		}
	}

	// 2. Verify path exists in trunk and is a directory
	// git ls-tree -d <trunk> <path>
	// We use execCommand to capture output for debugging
	lsCmd := exec.Command("git", "ls-tree", targetTrunk, repoRelPath)
	lsCmd.Dir = rootPath
	lsOut, err := lsCmd.CombinedOutput()

	// Log the ls-tree output to helps us see what git sees
	if err != nil || len(lsOut) == 0 {
		return fmt.Errorf("path '%s' not found in trunk '%s' (git ls-tree output: %s). Is it committed?", repoRelPath, targetTrunk, string(lsOut))
	}

	// Check if it is a directory (tree)
	// Output format: <mode> tree <hash> <path>
	lsOutStr := string(lsOut)
	_ = os.WriteFile(filepath.Join(rootPath, "gitgrove_debug_ls.log"), lsOut, 0644) // Check log

	if len(lsOutStr) < 6 || lsOutStr[7:11] != "tree" { // simplified check, usually "040000 tree ..."
		// It might be a blob?
		return fmt.Errorf("path '%s' in trunk '%s' is not a directory. Git subtree requires a directory. (git check: %s)", repoRelPath, targetTrunk, lsOutStr)
	}

	// 3. Create a temporary branch with the latest subtree state from trunk
	tempSyncBranch := fmt.Sprintf("gg-sync/%s/%d", repoName, time.Now().Unix())

	// Strategy: Switch to trunk to perform split, then switch back.
	// This avoids "fatal: prefix does not exist" errors when running from orphan branch.

	// A. Checkout Trunk
	if err := gitUtil.Checkout(rootPath, targetTrunk); err != nil {
		return fmt.Errorf("failed to checkout trunk '%s' for sync operation: %w", targetTrunk, err)
	}

	// Defer return to orphan branch in case of early failure
	// We need to capture the fact that we switched.
	// Better to just ensure we try to switch back.

	// B. Split Subtree (now we are in trunk)
	// We use "HEAD" explicitly since we have checked out the trunk.
	if err := gitUtil.SubtreeSplitFrom(rootPath, repoRelPath, "HEAD", tempSyncBranch); err != nil {
		// Try to return to orphan even if failed
		_ = gitUtil.Checkout(rootPath, currentBranch)

		// Log full error
		fullError := fmt.Sprintf("Command: git subtree split -P %s -b %s %s\nError: %v", repoRelPath, tempSyncBranch, targetTrunk, err)
		_ = os.WriteFile(filepath.Join(rootPath, "gitgrove_error.log"), []byte(fullError), 0644)
		return fmt.Errorf("failed to split subtree while in trunk. Check gitgrove_error.log")
	}

	// C. Return to Orphan Branch
	if err := gitUtil.Checkout(rootPath, currentBranch); err != nil {
		return fmt.Errorf("failed to return to orphan branch '%s' after split: %w", currentBranch, err)
	}

	// Clean immediately to remove artifacts from trunk visitation (e.g. vekku-server folder)
	if err := gitUtil.Clean(rootPath); err != nil {
		return fmt.Errorf("post-checkout clean failed: %w", err)
	}

	// Defer cleanup of temp branch
	defer func() {
		gitUtil.DeleteBranch(rootPath, tempSyncBranch, true)
	}()

	// 4. Merge the temporary branch into current orphan branch
	if err := gitUtil.Merge(rootPath, tempSyncBranch); err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	// 5. Clean untracked files again
	// Switching to trunk might have left untracked files that shouldn't be here in orphan mode.
	if err := gitUtil.Clean(rootPath); err != nil {
		return fmt.Errorf("post-sync clean failed: %w", err)
	}

	return nil
}
