package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/info"
)

// Down switches the working tree to a child repository's branch.
func Down(rootAbsPath, childName string) error {
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

	// 2. Get repo info to validate child
	repoInfo, err := info.GetRepoInfo(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to load repo info: %w", err)
	}

	// Verify child exists and is actually a child of current repo
	childRepo, ok := repoInfo.Repos[childName]
	if !ok {
		return fmt.Errorf("child repo '%s' not found", childName)
	}

	if childRepo.Repo.Parent != currentRepoName {
		return fmt.Errorf("repo '%s' is not a child of '%s'", childName, currentRepoName)
	}

	// 3. Switch to child
	branchName := "main"
	return CheckoutRepo(rootAbsPath, childName, branchName)
}
