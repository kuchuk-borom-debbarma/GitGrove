package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

// Sync handles merging changes from an orphan branch to the trunk.
func Sync(targetArg string, squash bool, commit bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	currentBranch, err := gitUtil.GetCurrentBranch(cwd)
	if err != nil {
		return err
	}

	// Case 1: We are on the Trunk (or at least valid GG root with config)
	if config, err := groveUtil.LoadConfig(cwd); err == nil {
		// We are likely on trunk. targetArg should be the repo name.
		repoName := targetArg
		if repoName == "" {
			return fmt.Errorf("usage: gg sync <repo-name> [--no-squash] [--no-commit] (when run from trunk)")
		}

		repo, exists := config.Repositories[repoName]
		if !exists {
			return fmt.Errorf("repository '%s' not found in gg.json", repoName)
		}

		fmt.Printf("Syncing %s (gg/%s) into %s... (Squash: %v, Commit: %v)\n", repoName, repoName, repo.Path, squash, commit)
		// Merge gg/<repoName> into current branch
		orphanBranch := fmt.Sprintf("gg/%s", repoName)
		message := fmt.Sprintf("[%s] Merge branch '%s' into '%s'", repoName, orphanBranch, repo.Path)

		if err := gitUtil.SubtreeMerge(cwd, repo.Path, orphanBranch, squash, message, commit); err != nil {
			return fmt.Errorf("subtree merge failed: %w", err)
		}
		if commit {
			fmt.Println("Sync successful.")
		} else {
			fmt.Println("Sync successful. Changes are staged (ready for review).")
		}
		return nil
	}

	// Case 2: We might be on an Orphan Branch
	if strings.HasPrefix(currentBranch, "gg/") {
		repoName := strings.TrimPrefix(currentBranch, "gg/")
		targetTrunk := targetArg
		if targetTrunk == "" {
			targetTrunk = "main" // Default to main
		}

		fmt.Printf("Detected orphan branch for '%s'. Switching to trunk '%s' to sync... (Squash: %v, Commit: %v)\n", repoName, targetTrunk, squash, commit)

		// Check if target trunk exists
		// Simple check by trying to read config from it? Or just checkout.
		// Let's read gg.json from target trunk to verify it knows about us.
		configFileContent, err := gitUtil.ReadFileFromBranch(cwd, targetTrunk, ".gg/gg.json")
		if err != nil {
			return fmt.Errorf("failed to read .gg/gg.json from '%s'. Is it a valid trunk?: %w", targetTrunk, err)
		}

		var config groveUtil.GGConfig
		if err := json.Unmarshal(configFileContent, &config); err != nil {
			return fmt.Errorf("failed to parse gg.json from '%s': %w", targetTrunk, err)
		}

		if config.Repositories == nil {
			config.Repositories = make(map[string]model.GGRepo)
		}

		repo, exists := config.Repositories[repoName]
		if !exists {
			return fmt.Errorf("repository '%s' is not registered in '%s'", repoName, targetTrunk)
		}

		// Checkout Trunk
		cmd := exec.Command("git", "checkout", targetTrunk)
		cmd.Dir = cwd
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout '%s': %w", targetTrunk, err)
		}

		fmt.Printf("Switched to '%s'. Merging...\n", targetTrunk)
		message := fmt.Sprintf("[%s] Merge branch '%s' into '%s'", repoName, currentBranch, repo.Path)

		if err := gitUtil.SubtreeMerge(cwd, repo.Path, currentBranch, squash, message, commit); err != nil {
			return fmt.Errorf("subtree merge failed: %w", err)
		}
		if commit {
			fmt.Println("Sync successful.")
		} else {
			fmt.Println("Sync successful. Changes are staged (ready for review).")
		}
		return nil
	}

	return fmt.Errorf("not in a valid GitGrove root (no gg.json) and not in a recognized orphan branch (gg/*)")
}
