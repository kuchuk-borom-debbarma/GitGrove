package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// PreCommit enforces atomic commits in the GitGrove monorepo.
func PreCommit() error {
	// 1. Check for .gg/gg.json
	// Ensure we are at the root
	root, err := gitUtil.RepoRoot()
	if err != nil {
		// Fallback specific to pre-commit: if we can't find root, we can't enforce global rules reliably?
		// Or we assume CWD is ok?
		// Better to fail open or try CWD?
		// Let's try CWD as fallback.
		root, _ = os.Getwd()
	}

	config, err := groveUtil.LoadConfig(root)
	if err != nil {
		// If config load fails because file doesn't exist, we assume we are not in a context that needs enforcement
		// This covers orphan branches and non-grove repos
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			return nil
		}
		// Double check existence to be sure
		if _, statErr := os.Stat(filepath.Join(root, ".gg", "gg.json")); os.IsNotExist(statErr) {
			return nil
		}
		return err
	}

	// 2. Get staged files
	stagedFiles, err := gitUtil.GetStagedFiles(root)
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(stagedFiles) == 0 {
		return nil
	}

	// 3. Enforce Atomic Commit
	affectedRepos := make(map[string]bool)
	affectedRoot := false

	for _, file := range stagedFiles {
		matched := false
		for _, repo := range config.Repositories {
			// Check if file is inside repo.Path
			// We assume repo.Path is relative to root
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

	// Rules:
	// 1. Cannot touch > 1 registered repo
	if len(affectedRepos) > 1 {
		repos := []string{}
		for r := range affectedRepos {
			repos = append(repos, r)
		}
		return fmt.Errorf("atomic commit violation: commit touches multiple registered repositories: %v", repos)
	}

	// 2. Cannot touch Repo + Root
	if len(affectedRepos) == 1 && affectedRoot {
		var repoName string
		for r := range affectedRepos {
			repoName = r
		}
		return fmt.Errorf("atomic commit violation: commit mixes files from repository '%s' and root files", repoName)
	}

	return nil
}
