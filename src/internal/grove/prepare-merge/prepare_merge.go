package preparemerge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// PrepareMerge handles the logic for preparing a merge from an orphan branch to the trunk.
func PrepareMerge(ggRepoPath string, repoNameArg string) error {
	// 1. Context Detection
	currentBranch, err := gitUtil.CurrentBranch(ggRepoPath)
	if err != nil {
		return err
	}

	var targetRepoName string
	initialBranch := currentBranch

	if strings.HasPrefix(currentBranch, "gg/") {
		// Orphan Branch Context
		targetRepoName = strings.TrimPrefix(currentBranch, "gg/")
		fmt.Printf("Detected orphan branch context for repo: %s\n", targetRepoName)

		// Switch to Trunk (Assuming 'main' for now)
		// TODO: Ideally, we should detect the trunk branch name from config or context if possible.
		// For now, we assume 'main' as per design docs.
		trunkBranch := "main"
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
	orphanBranchName := fmt.Sprintf("gg/%s", targetRepoName)
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
