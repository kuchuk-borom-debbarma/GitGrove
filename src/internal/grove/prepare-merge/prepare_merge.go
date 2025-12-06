package preparemerge

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// Description returns a description of the prepare merge process.
func Description() string {
	return "Prepare Merge: Prepares work for integration into the Trunk.\n" +
		"- Switches to trunk\n" +
		"- Creates a temporary merge-prep branch\n" +
		"- Merges changes from the orphan branch (restoring directory structure)\n" +
		"- Excludes .gg/trunk artifact"
}

// PrepareMerge handles the logic for preparing a merge from an orphan branch to the trunk.
func PrepareMerge(ggRepoPath string, repoNameArg string) error {
	ggRepoPath = filepath.Clean(ggRepoPath)
	// 1. Context Detection
	currentBranch, err := gitUtil.CurrentBranch(ggRepoPath)
	if err != nil {
		return err
	}

	var targetRepoName string
	var trunkBranch string = "main" // Default fallback
	initialBranch := currentBranch

	if strings.HasPrefix(currentBranch, "gg/") {
		// Orphan Branch Context: gg/<trunk>/<repoName>
		parts := strings.Split(currentBranch, "/")
		if len(parts) >= 3 {
			// Expected format: gg/<trunk>/<repoName>
			// Assuming trunk doesn't contain slashes for now, or we take parts[1]
			// If we want to support trunk with slashes, we need to know where repoName starts.
			// Currently register_repo uses gg/<trunk>/<repoName>.
			// Let's assume repoName is the last part.
			targetRepoName = parts[len(parts)-1]
			trunkBranch = strings.Join(parts[1:len(parts)-1], "/")
		} else {
			// Fallback for old format or unexpected: gg/<repoName> (implied trunk=main)
			targetRepoName = strings.TrimPrefix(currentBranch, "gg/")
			trunkBranch = "main"
		}

		fmt.Printf("Detected orphan branch context for repo: %s (trunk: %s)\n", targetRepoName, trunkBranch)

		fmt.Printf("Switching to trunk '%s'...\n", trunkBranch)
		if err := gitUtil.Checkout(ggRepoPath, trunkBranch); err != nil {
			return fmt.Errorf("failed to checkout trunk '%s': %w", trunkBranch, err)
		}
	} else {
		// Trunk Context (or other)
		if repoNameArg == "" {
			return fmt.Errorf("repository name is required when not in an orphan branch")
		}
		targetRepoName = repoNameArg
		trunkBranch = currentBranch // We assume we are on trunk
	}

	// 2. Validation (Now that we are potentially on Trunk)
	// Check if config exists
	configPath := filepath.Join(ggRepoPath, ".gg", "gg.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If we switched to main and still no config, something is wrong.
		configPath = filepath.Join(ggRepoPath, ".gg", "gg.json")
		// Double check if path is correct or if we should rely on LoadConfig error?
		// os.Stat check is good for specific error message.
		return fmt.Errorf("gitgrove is not initialized in %s", ggRepoPath)
	}

	config, err := groveUtil.LoadConfig(ggRepoPath)
	if err != nil {
		return err
	}

	repoConfig, exists := config.Repositories[targetRepoName]
	if !exists {
		return fmt.Errorf("repository '%s' not registered in gitgrove", targetRepoName)
	}

	// 3. Branch Preparation
	// 3. Branch Preparation
	orphanBranchName := fmt.Sprintf("gg/%s/%s", trunkBranch, targetRepoName)
	timestamp := time.Now().Format("20060102-150405")
	prepareBranchName := fmt.Sprintf("gg/merge-prep/%s/%s", targetRepoName, timestamp)

	fmt.Printf("Creating prepare-merge branch: %s\n", prepareBranchName)
	// We are currently on Trunk (main)
	if err := gitUtil.CreateBranch(ggRepoPath, prepareBranchName); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", prepareBranchName, err)
	}

	// 4. Merge
	fmt.Printf("Merging changes from %s...\n", orphanBranchName)
	if err := gitUtil.SubtreeMerge(ggRepoPath, repoConfig.Path, orphanBranchName); err != nil {
		return fmt.Errorf("failed to merge orphan branch %s: %w", orphanBranchName, err)
	}

	// 4.1. Exclude .gg/trunk if present
	trunkFilePath := filepath.Join(ggRepoPath, ".gg", "trunk")
	if _, err := os.Stat(trunkFilePath); err == nil {
		fmt.Println("Removing .gg/trunk from merge result...")
		if err := os.Remove(trunkFilePath); err != nil {
			return fmt.Errorf("failed to remove .gg/trunk: %w", err)
		}
		// We need to stage this removal so it's part of the next commit or if we are amending?
		// Subtree merge usually commits.
		// If subtree merge committed, we need to create a new commit or amend?
		// Typically prepare-merge creates a branch with the merge.
		// If we remove the file now, it's a modification on top of the merge.
		// We should commit it? Or just leave it as Unstaged?
		// Best is to amend the merge commit to exclude this file.
		// Or just create a new commit "Exclude .gg/trunk".
		// Or `git rm .gg/trunk` and commit.

		cmd := exec.Command("git", "rm", ".gg/trunk")
		cmd.Dir = ggRepoPath
		if err := cmd.Run(); err != nil {
			// If git rm fails (maybe not tracked?), try just removing it.
			// But if it was merged, it is tracked.
		}

		if err := gitUtil.Commit(ggRepoPath, []string{".gg/trunk"}, "Exclude .gg/trunk from merge"); err != nil {
			// If commiting failed, maybe nothing to commit?
		}
	}

	// 5. Success
	fmt.Println("\nSuccess! Prepare-merge branch created.")
	fmt.Printf("Branch: %s\n", prepareBranchName)
	fmt.Println("Review the changes and submit a Pull Request to merge into main.")

	// If we started on orphan, we are now on the new branch. This is the desired behavior ("prepare for merge").
	if initialBranch != "main" && !strings.HasPrefix(initialBranch, "prepare-merge") {
		// If user was on orphan, they are now on prepare-merge branch.
		// If user was on main, they are now on prepare-merge branch.
		// Logic holds consistent.
	}

	return nil
}
