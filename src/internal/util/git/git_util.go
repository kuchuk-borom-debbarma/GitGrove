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

// GetCurrentBranch returns the name of the current branch.
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %s: %w", string(output), err)
	}
	return strings.TrimSpace(string(output)), nil
}

// SubtreeMerge merges the given branch into the current branch using git subtree merge.
func SubtreeMerge(repoPath string, prefix string, branchName string, squash bool, message string, commit bool) error {
	if !commit && !squash {
		// git merge -Xsubtree=<prefix> --no-commit --allow-unrelated-histories <branch>
		args := []string{"merge", "-Xsubtree=" + prefix, "--no-commit", "--allow-unrelated-histories", branchName}
		if message != "" {
			args = append(args, "-m", message)
		}
		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git merge -Xsubtree failed: %s: %w", string(output), err)
		}
		return nil
	}

	// Standard git subtree merge (handles squash and commit)
	args := []string{"subtree", "merge", "--prefix=" + prefix, branchName}
	if squash {
		args = append(args, "--squash")
	}

	if message != "" {
		args = append(args, "-m", message)
	} else {
		args = append(args, "-m", fmt.Sprintf("Merge %s into %s", branchName, prefix))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git subtree merge failed: %s: %w", string(output), err)
	}

	// If no-commit requested BUT we used subtree merge (which forced commit), we must undo the commit
	// This applies to the squash case.
	if !commit {
		// git reset --soft HEAD~1
		resetCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
		resetCmd.Dir = repoPath
		if output, err := resetCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to reset after subtree merge (simulating no-commit): %s: %w", string(output), err)
		}
	}

	return nil
}

// ReadFileFromBranch reads the content of a file from a specific branch.
func ReadFileFromBranch(repoPath string, branch string, filePath string) ([]byte, error) {
	// git show branch:path
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", branch, filePath))
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git show failed: %s: %w", string(output), err)
	}
	return output, nil
}
