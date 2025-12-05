package gitUtil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// GetStagedFiles returns a list of files that are currently staged for commit.
func GetStagedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %s: %w", string(output), err)
	}

	files := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			files = append(files, strings.TrimSpace(line))
		}
	}
	return files, nil
}

// SubtreeMerge merges the given branch using git subtree merge.
func SubtreeMerge(repoPath string, prefix string, branchName string) error {
	// Use 'git merge -s subtree' directly to support --allow-unrelated-histories
	// Use -Xtheirs to resolve "add/add" conflicts caused by unrelated histories (assuming main hasn't changed registered paths as per design)
	cmd := exec.Command("git", "merge", "-s", "subtree", "--allow-unrelated-histories", "-Xsubtree="+prefix, "-Xtheirs", branchName, "-m", "Merge orphan branch "+branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge -s subtree failed: %s: %w", string(output), err)
	}
	return nil
}

// CurrentBranch returns the name of the current git branch.
func CurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git current-branch failed: %s: %w", string(output), err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Checkout switches to the specified branch.
func Checkout(repoPath string, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %s: %w", string(output), err)
	}
	return nil
}

// CreateBranch creates a new branch (off the current HEAD) and switches to it.
func CreateBranch(repoPath string, branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git create-branch failed: %s: %w", string(output), err)
	}
	return nil
}
