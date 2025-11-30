package grove

import (
	"fmt"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// validateCleanGitRepo validates that the given path is a clean git repository.
// This is a common validation used across multiple GitGrove operations.
func validateCleanGitRepo(rootAbsPath string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}
	return nil
}

// validateRepoExists checks if a repository exists in the given map.
func validateRepoExists(repos map[string]model.Repo, name string) error {
	if _, ok := repos[name]; !ok {
		return fmt.Errorf("repository '%s' not found", name)
	}
	return nil
}

// validateNoCycles checks that the given relationships don't create cycles.
// This is extracted from link.go for reusability.
func validateNoCycles(relationships map[string]string, existing map[string]model.Repo) error {
	// Build full parent map (existing + new)
	parentMap := make(map[string]string)
	for name, repo := range existing {
		if repo.Parent != "" {
			parentMap[name] = repo.Parent
		}
	}
	for child, parent := range relationships {
		parentMap[child] = parent
	}

	// Check each new relationship for cycles
	for child := range relationships {
		visited := make(map[string]bool)
		current := child

		for current != "" {
			if visited[current] {
				return fmt.Errorf("cycle detected: repository '%s' is part of a circular parent chain", child)
			}
			visited[current] = true
			current = parentMap[current]
		}
	}

	return nil
}
