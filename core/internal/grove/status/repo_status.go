package status

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

type RepoState struct {
	Repo       model.Repo
	PathExists bool
}

type RepoStatus struct {
	Repos map[string]RepoState
}

func GetRepoStatus(rootAbsPath string) (*RepoStatus, error) {
	systemRef := "refs/heads/gitgroove/system"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, systemRef)
	if err != nil {
		// If system branch doesn't exist, we assume no repos registered or not init
		return &RepoStatus{Repos: map[string]RepoState{}}, nil
	}

	repos, err := loadRepos(rootAbsPath, oldTip)
	if err != nil {
		return nil, err
	}

	repoStates := make(map[string]RepoState)
	for name, repo := range repos {
		absPath := filepath.Join(rootAbsPath, repo.Path)
		exists := false
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			exists = true
		}
		repoStates[name] = RepoState{
			Repo:       repo,
			PathExists: exists,
		}
	}

	return &RepoStatus{Repos: repoStates}, nil
}

// loadRepos is a helper to load repo metadata from a specific commit.
func loadRepos(root, ref string) (map[string]model.Repo, error) {
	entries, err := gitUtil.ListTree(root, ref, ".gg/repos")
	if err != nil {
		// Likely empty or doesn't exist
		return map[string]model.Repo{}, nil
	}

	repos := make(map[string]model.Repo)
	for _, name := range entries {
		if name == ".gitkeep" {
			continue
		}

		pathFile := fmt.Sprintf(".gg/repos/%s/path", name)
		content, err := gitUtil.ShowFile(root, ref, pathFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read path for repo %s: %w", name, err)
		}
		repoPath := strings.TrimSpace(content)

		parentFile := fmt.Sprintf(".gg/repos/%s/parent", name)
		parentContent, err := gitUtil.ShowFile(root, ref, parentFile)
		parent := ""
		if err == nil {
			parent = strings.TrimSpace(parentContent)
		}

		repos[name] = model.Repo{
			Name:   name,
			Path:   repoPath,
			Parent: parent,
		}
	}
	return repos, nil
}
