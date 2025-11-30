package core

import (
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/info"
)

// Init initializes GitGroove on the current Git repository.
// It delegates to the internal grove package.
func Init(absolutePath string) error {
	return grove.Init(absolutePath)
}

// IsInitialized checks if GitGrove is initialized in the given path.
func IsInitialized(rootAbsPath string) (bool, error) {
	return grove.IsInitialized(rootAbsPath)
}

// Register records one or more repos (name → path) in the GitGroove metadata.
func Register(rootAbsPath string, repos map[string]string) error {
	return grove.Register(rootAbsPath, repos)
}

// Link defines one or more repo hierarchy relationships (childName → parentName).
func Link(rootAbsPath string, relationships map[string]string) error {
	return grove.Link(rootAbsPath, relationships)
}

// Switch moves the user's working tree to the GitGroove branch associated with the specified repo.
func Switch(rootAbsPath, repoName, branch string) error {
	return grove.Switch(rootAbsPath, repoName, branch)
}

// Info returns the current health and status of the GitGroove project.
func Info(rootAbsPath string) (string, error) {
	d, err := info.GetInfo(rootAbsPath)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}

// CreateRepoBranch creates a new branch for a specific nested repository.
func CreateRepoBranch(rootAbsPath, repoName, branchName string) error {
	return grove.CreateRepoBranch(rootAbsPath, repoName, branchName)
}

// CheckoutRepo switches the user's working tree to a specific branch of a nested repository.
func CheckoutRepo(rootAbsPath, repoName, branchName string) error {
	return grove.CheckoutRepo(rootAbsPath, repoName, branchName)
}

// Add adds file contents to the staging area with GitGrove-specific validations.
func Add(rootAbsPath string, files []string) error {
	return grove.Add(rootAbsPath, files)
}

// Commit performs a commit with strict GitGrove validations.
func Commit(rootAbsPath, message string) error {
	return grove.Commit(rootAbsPath, message)
}

// Move relocates a registered repository to a new path within the project.
func Move(rootAbsPath, repoName, newRelPath string) error {
	return grove.Move(rootAbsPath, repoName, newRelPath)
}

// Repo represents a registered repository in the public API.
type Repo struct {
	Name   string
	Path   string
	Parent string
}

// GetRepositories returns a list of all registered repositories.
func GetRepositories(rootAbsPath string) ([]Repo, error) {
	repoInfo, err := info.GetRepoInfo(rootAbsPath)
	if err != nil {
		return nil, err
	}

	var repos []Repo
	for _, state := range repoInfo.Repos {
		repos = append(repos, Repo{
			Name:   state.Repo.Name,
			Path:   state.Repo.Path,
			Parent: state.Repo.Parent,
		})
	}
	return repos, nil
}
