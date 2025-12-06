package registerrepo

import (
	"fmt"
	"os"
	"path/filepath"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

// Description returns a description of the register repo process.
func Description() string {
	return "Register Repo: Isolates a folder into a logical repository.\n" +
		"- Adds repository metadata to gg.json\n" +
		"- Creates an orphan branch (gg/<trunk>/<repoName>) with isolated history\n" +
		"- Moves files to root in the orphan branch"
}

// RegisterRepo registers a folder as a "repo" within the GitGrove monorepo.
//
// Concept: The Split
// When a folder is registered as a "repo," GG creates a parallel history for it.
//
// Workflow:
//  1. Orphan Creation: GG scans the history of the target folder.
//  2. Path Translation: It uses git subtree split to create a new orphan branch (e.g., gg/serviceA).
//  3. Root Projection: Inside this new branch, the files are moved to the Root Directory.
//     - Trunk View: ./backend/services/serviceA/main.go
//     - Orphan View: ./main.go
//
// Limitation: Nested Repositories
// Nested directories cannot be registered as repositories at this time.
// Rules for this are yet to be clearly defined.
func RegisterRepo(repos []model.GGRepo, ggRepoPath string) error {
	// Validate ggRepoPath (has .gg/gg.json and is git repo too)
	if err := gitUtil.IsGitRepository(ggRepoPath); err != nil {
		return err
	}

	// Check if Grove is initialized
	configPath := filepath.Join(ggRepoPath, ".gg", "gg.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("gitgrove is not initialized in %s", ggRepoPath)
	}

	// Load repos from gg.json (needed for validation)
	config, err := groveUtil.LoadConfig(ggRepoPath)
	if err != nil {
		return err
	}

	// Validate BEFORE doing any git operations
	if err := groveUtil.ValidateRepoRegistration(ggRepoPath, config, repos); err != nil {
		return err
	}

	// Get current branch
	currentBranch, err := gitUtil.CurrentBranch(ggRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// If all good, proceed creating the orphan branch
	for _, repo := range repos {
		branchName := fmt.Sprintf("gg/%s/%s", currentBranch, repo.Name)
		if err := gitUtil.SubtreeSplit(ggRepoPath, repo.Path, branchName); err != nil {
			return fmt.Errorf("failed to create subtree split for %s: %w", repo.Name, err)
		}
	}

	// ONLY if git operations succeed, update gg.json
	if err := groveUtil.AddReposToConfig(ggRepoPath, repos); err != nil {
		return fmt.Errorf("failed to update config after branch creation: %w", err)
	}

	return nil
}
