package core

import (
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/doctor"
)

// Init initializes GitGroove on the current Git repository.
// It delegates to the internal grove package.
func Init(absolutePath string) error {
	return grove.Init(absolutePath)
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

// Doctor returns the current health and status of the GitGroove project.
func Doctor(rootAbsPath string) (string, error) {
	d, err := doctor.GetDoctor(rootAbsPath)
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

// Stage adds file contents to the staging area with GitGrove-specific validations.
func Stage(rootAbsPath string, files []string) error {
	return grove.Stage(rootAbsPath, files)
}

// Commit performs a commit with strict GitGrove validations.
func Commit(rootAbsPath, message string) error {
	return grove.Commit(rootAbsPath, message)
}

// Move relocates a registered repository to a new path within the project.
func Move(rootAbsPath, repoName, newRelPath string) error {
	return grove.Move(rootAbsPath, repoName, newRelPath)
}
