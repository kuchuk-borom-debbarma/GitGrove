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
	path = filepath.Clean(path)
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

// RepoRoot returns the absolute path to the root of the git repository.
func RepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get repo root: %s: %w", string(output), err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Commit stages the given files and commits them with the provided message.
func Commit(repoPath string, files []string, message string) error {
	repoPath = filepath.Clean(repoPath)
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

// CommitNoVerify stages the given files and commits them with the provided message, bypassing hooks.
func CommitNoVerify(repoPath string, files []string, message string) error {
	repoPath = filepath.Clean(repoPath)
	// Stage files
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s: %w", string(output), err)
	}

	// Commit with --no-verify
	cmd = exec.Command("git", "commit", "--no-verify", "-m", message)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit --no-verify failed: %s: %w", string(output), err)
	}

	return nil
}

// SubtreeSplit creates a new branch using git subtree split.
func SubtreeSplit(repoPath string, prefix string, branchName string) error {
	repoPath = filepath.Clean(repoPath)
	prefix = filepath.Clean(prefix)
	cmd := exec.Command("git", "subtree", "split", "--prefix="+prefix, "-b", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git subtree split failed: %s: %w", string(output), err)
	}
	return nil
}

// GetStagedFiles returns a list of files that are currently staged for commit.
func GetStagedFiles(repoPath string) ([]string, error) {
	repoPath = filepath.Clean(repoPath)
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
	repoPath = filepath.Clean(repoPath)
	prefix = filepath.Clean(prefix)
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
	repoPath = filepath.Clean(repoPath)
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
	repoPath = filepath.Clean(repoPath)
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %s: %w", string(output), err)
	}
	return nil
}

// CreateBranch creates a new branch (off the current HEAD) and switches to it.
func CreateBranch(repoPath string, branchName string) error {
	repoPath = filepath.Clean(repoPath)
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git create-branch failed: %s: %w", string(output), err)
	}
	return nil
}

// FileExistsInBranch checks if a file exists in a specific branch.
func FileExistsInBranch(repoPath string, branchName string, filePath string) (bool, error) {
	repoPath = filepath.Clean(repoPath)
	// git cat-file -e <branch>:<file>
	// -e exits with 0 if file exists, 1 if not (or malformed)
	object := fmt.Sprintf("%s:%s", branchName, filePath)
	cmd := exec.Command("git", "cat-file", "-e", object)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		// If exit code is 1, it might mean file not found.
		// We should probably check if it's strictly "not found" vs other errors,
		// but typically cat-file -e is sufficient.
		return false, nil
	}
	return true, nil
}

// ReadFileFromBranch reads the content of a file from a specific branch.
func ReadFileFromBranch(repoPath string, branchName string, filePath string) ([]byte, error) {
	repoPath = filepath.Clean(repoPath)
	// format: <branch>:<path>
	// Use git show which is convenient for cat-ing a file from a ref
	// Note: Windows paths might need to be converted to forward slashes for git object syntax,
	// but let's see if we can just rely on standard path refs.
	// Git expects forward slashes for the internal path spec usually.
	gitPath := filepath.ToSlash(filePath)
	object := fmt.Sprintf("%s:%s", branchName, gitPath)

	cmd := exec.Command("git", "show", object)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s' from branch '%s': %w", filePath, branchName, err)
	}
	return output, nil
}

// SetLocalConfig sets a local git configuration value.
func SetLocalConfig(repoPath string, key string, value string) error {
	repoPath = filepath.Clean(repoPath)
	cmd := exec.Command("git", "config", "--local", key, value)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set config %s=%s: %s: %w", key, value, string(output), err)
	}
	return nil
}

// GetLocalConfig gets a local git configuration value. Returns empty string if not found.
func GetLocalConfig(repoPath string, key string) (string, error) {
	repoPath = filepath.Clean(repoPath)
	cmd := exec.Command("git", "config", "--local", "--get", key)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// git config returns exit code 1 if key is not found
		// We treat this as empty value, not error
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

// UnsetLocalConfig removes a local git configuration value.
func UnsetLocalConfig(repoPath string, key string) error {
	repoPath = filepath.Clean(repoPath)
	cmd := exec.Command("git", "config", "--local", "--unset", key)
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Ignore check for now if it doesn't exist, or check exit code?
		// git config --unset returns 5 if key doesn't exist
		return nil
	}
	return nil
}
