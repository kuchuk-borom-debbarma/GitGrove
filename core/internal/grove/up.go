package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/info"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// Up switches the working tree to the parent repository's branch.
func Up(rootAbsPath string) error {
	// Check if we are on system branch
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err == nil && currentBranch == "gitgroove/system" {
		// Already at System Root.
		fmt.Println("Already at System Root.")
		return nil
	}

	// 1. Identify current repo
	markerPath := filepath.Join(rootAbsPath, ".gitgroverepo")
	content, err := os.ReadFile(markerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not inside a registered repository (missing .gitgroverepo)")
		}
		return fmt.Errorf("failed to read .gitgroverepo: %w", err)
	}
	currentRepoName := strings.TrimSpace(string(content))

	// 2. Get repo info to find parent
	repoInfo, err := info.GetRepoInfo(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to load repo info: %w", err)
	}

	currentRepo, ok := repoInfo.Repos[currentRepoName]
	if !ok {
		return fmt.Errorf("current repo '%s' not found in metadata", currentRepoName)
	}

	parentName := currentRepo.Repo.Parent
	if parentName == "" {
		// If no parent, we are at a root repo.
		// "Up" from root means going to the System Root view.
		return SwitchToSystem(rootAbsPath)
	}

	// We need the parent repo object to get its name and default branch
	parentRepo, ok := repoInfo.Repos[parentName]
	if !ok {
		return fmt.Errorf("parent repo '%s' not found in metadata", parentName)
	}

	// 3. Switch to parent
	// We use the default branch for now (main)
	// TODO: Track last active branch for each repo?
	branch := "main" // Assuming 'main' is the default branch for now
	if err := CheckoutRepo(rootAbsPath, parentRepo.Repo.Name, branch, false, false); err != nil {
		return fmt.Errorf("failed to switch to parent repo '%s': %w", parentName, err)
	}
	return nil
}
