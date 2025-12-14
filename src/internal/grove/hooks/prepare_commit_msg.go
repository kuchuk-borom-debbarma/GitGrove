package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// PrepareCommitMsg modifies the commit message to prepend [RepoName] if strict context is established.
func PrepareCommitMsg(msgFile, source, sha string) error {
	// Determine Repository Root
	// Hooks are executed from the root of the repo (usually), or CWD depending on how git is invoked?
	// Actually, hooks are run with CWD set to the root of the working tree by git.
	// BUT, just to be safe and consistent esp. if called manually or in weird environments:
	root, err := gitUtil.RepoRoot()
	if err != nil {
		// Fallback to CWD if not in a git repo (unlikely for a hook)
		root, _ = os.Getwd()
	}

	// Only run on normal commits (skip rebase, merge, amend if preferred, but usually we want it on message creation)
	// source can be: message, template, merge, squash, commit (when -c/-C/-F is used).
	// We mainly want to catch when user types a new message or uses -m.
	// If source is empty, it's usually `git commit` (no -m) opening editor?
	// Actually git docs say:
	// - message (if -m or -F)
	// - template (if -t)
	// - merge (if merging)
	// - squash (if squashing)
	// - commit (if -c, -C, --amend)
	// If source is missing, it's standard commit.

	// If it's a merge, we probably shouldn't mess with it? Or maybe we should?
	// Let's stick to standard commits for now.

	config, err := groveUtil.LoadConfig(root)
	if err != nil {
		// Not a grove repo or error loading config.
		// If working in an orphan branch (gg/<trunk>/<repoName>), config might not exist on disk,
		// but should exist in the <trunk> branch.
		isMissing := os.IsNotExist(err) || strings.Contains(err.Error(), "no such file")

		loadedFromBranch := false
		if isMissing {
			// Check if we are in an orphan branch
			currentBranch, branchErr := gitUtil.CurrentBranch(root)
			if branchErr == nil && strings.HasPrefix(currentBranch, "gg/") {
				parts := strings.Split(currentBranch, "/")
				if len(parts) >= 3 {
					// gg/<trunk>/<repoName> -> we need <trunk> (might contain slashes)
					// Assumes loose structure: trunk is everything between gg/ and /<repoName>
					// Or we can try to find config in potential trunk candidates.
					// Simplest heuristic: <trunk> is everything in between.
					trunk := strings.Join(parts[1:len(parts)-1], "/")

					// Try to load from trunk
					branchConfig, branchConfigErr := groveUtil.LoadConfigFromGitRef(root, trunk)
					if branchConfigErr == nil {
						config = branchConfig
						loadedFromBranch = true
					}
				}
			}
		}

		if !loadedFromBranch {
			// If still not loaded, treat as genuine error or not initialized
			if isMissing {
				// Not initialized, just return
				return nil
			}
			// Double check existence to be sure
			if _, statErr := os.Stat(filepath.Join(root, ".gg", "gg.json")); os.IsNotExist(statErr) {
				return nil
			}
			return err
		}
	}

	// fmt.Printf("Debug: Config loaded. AtomicCommit: %v\n", config.AtomicCommit)

	if !config.RepoAwareContextMessage {
		return nil
	}

	// 0. Sticky Context Logic (Priority 0)
	stickyRepo, _ := groveUtil.GetContextRepo(root)
	if stickyRepo != "" {
		if _, exists := config.Repositories[stickyRepo]; exists {
			return prependRepoName(msgFile, stickyRepo)
		}
	}

	// 1. Orphan Branch Logic (Priority)
	currentBranch, err := gitUtil.CurrentBranch(root)
	if err == nil && strings.HasPrefix(currentBranch, "gg/") {
		// Parse repo name from branch: gg/<trunk>/<repoName>
		parts := strings.Split(currentBranch, "/")
		if len(parts) >= 3 {
			repoName := parts[len(parts)-1]
			// In an orphan branch, EVERYTHING belongs to this repo.
			return prependRepoName(msgFile, repoName)
		}
	}

	// 2. Trunk/Monorepo Logic (Scanning staged files)
	// Logic from pre_commit.go to find affected repos
	stagedFiles, err := gitUtil.GetStagedFiles(root)
	// fmt.Printf("Debug: Staged files: %v\n", stagedFiles)
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(stagedFiles) == 0 {
		return nil
	}

	affectedRepos := make(map[string]bool)
	affectedRoot := false

	for _, file := range stagedFiles {
		matched := false
		for _, repo := range config.Repositories {
			relPath, err := filepath.Rel(repo.Path, file)
			if err == nil && !strings.HasPrefix(relPath, "..") {
				affectedRepos[repo.Name] = true
				matched = true
				break
			}
		}
		if !matched {
			affectedRoot = true
		}
	}

	// Logic:
	// If 1 repo affected AND no root files affected -> Prepend [RepoName]
	if len(affectedRepos) == 1 && !affectedRoot {
		var repoName string
		for r := range affectedRepos {
			repoName = r
		}
		return prependRepoName(msgFile, repoName)
	}

	return nil
}

func prependRepoName(msgFile, repoName string) error {
	// Read existing message
	msgBytes, err := os.ReadFile(msgFile)
	if err != nil {
		return fmt.Errorf("failed to read commit message file: %w", err)
	}
	msg := string(msgBytes)

	prefix := fmt.Sprintf("[%s] ", repoName)
	if !strings.HasPrefix(msg, prefix) {
		newMsg := prefix + msg
		if err := os.WriteFile(msgFile, []byte(newMsg), 0644); err != nil {
			return fmt.Errorf("failed to write commit message file: %w", err)
		}
	}
	return nil
}
