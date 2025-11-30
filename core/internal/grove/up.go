package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/info"
)

// Up switches the working tree to the parent repository's branch.
func Up(rootAbsPath string) error {
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
		return fmt.Errorf("repo '%s' has no parent (it is a root)", currentRepoName)
	}

	// 3. Switch to parent
	// We use the default branch for now (main)
	// TODO: Track last active branch for each repo?
	branchName := "main"
	return CheckoutRepo(rootAbsPath, parentName, branchName)
}
