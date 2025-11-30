package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/info"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// Ls lists the child repositories of the current repository.
func Ls(rootAbsPath string) ([]string, error) {
	// Check if we are on system branch
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err == nil && currentBranch == "gitgroove/system" {
		// We are at System Root. List all root repositories.
		repoInfo, err := info.GetRepoInfo(rootAbsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load repo info: %w", err)
		}
		var roots []string
		for name, state := range repoInfo.Repos {
			if state.Repo.Parent == "" {
				roots = append(roots, name)
			}
		}
		sort.Strings(roots)
		return roots, nil
	}

	// Read current repo from .gitgroverepo marker
	markerPath := filepath.Join(rootAbsPath, ".gitgroverepo")
	content, err := os.ReadFile(markerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not inside a registered repository (missing .gitgroverepo)")
		}
		return nil, fmt.Errorf("failed to read .gitgroverepo: %w", err)
	}
	currentRepoName := strings.TrimSpace(string(content))

	// Load repository metadata
	repoInfo, err := info.GetRepoInfo(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load repo info: %w", err)
	}

	// Find all children of the current repo
	var children []string
	for name, state := range repoInfo.Repos {
		if state.Repo.Parent == currentRepoName {
			children = append(children, name)
		}
	}

	sort.Strings(children)
	return children, nil
}
