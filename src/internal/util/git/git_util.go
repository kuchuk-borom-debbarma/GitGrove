package gitUtil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// IsGitRepository checks if the given path is a valid git repository.
func IsGitRepository(path string) error {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: %s", path)
		}
		return fmt.Errorf("error checking git repository: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf(".git is not a directory in %s", path)
	}
	return nil
}

// Commit stages the given files and commits them with the provided message.
func Commit(repoPath string, files []string, message string) error {
	// Stage files
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s: %w", string(output), err)
	}

	// Commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s: %w", string(output), err)
	}

	return nil
}

// SubtreeSplit creates a new branch using git subtree split.
func SubtreeSplit(repoPath string, prefix string, branchName string) error {
	cmd := exec.Command("git", "subtree", "split", "--prefix="+prefix, "-b", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git subtree split failed: %s: %w", string(output), err)
	}
	return nil
}
