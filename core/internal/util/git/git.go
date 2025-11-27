package git

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

// Internal helper to run git
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func IsInsideGitRepo(path string) bool {
	out, err := runGit(path, "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

func IsDetachedHEAD(path string) bool {
	out, err := runGit(path, "rev-parse", "--symbolic-full-name", "HEAD")
	if err != nil {
		return false
	}
	return out == "HEAD"
}

func HasStagedChanges(path string) bool {
	_, err := runGit(path, "diff", "--cached", "--quiet")
	return err != nil
}

func HasUnstagedChanges(path string) bool {
	_, err := runGit(path, "diff", "--quiet")
	return err != nil
}

func HasUntrackedFiles(path string) bool {
	out, err := runGit(path, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

func VerifyCleanState(path string) error {
	var issues []string

	if IsDetachedHEAD(path) {
		issues = append(issues, "HEAD is detached")
	}
	if HasStagedChanges(path) {
		issues = append(issues, "staged changes exist")
	}
	if HasUnstagedChanges(path) {
		issues = append(issues, "unstaged changes exist")
	}
	if HasUntrackedFiles(path) {
		issues = append(issues, "untracked files exist")
	}

	if len(issues) == 0 {
		return nil
	}

	return errors.New("repository is not clean: " + strings.Join(issues, "; "))
}

func HasBranch(path, branch string) (bool, error) {
	_, err := runGit(path, "rev-parse", "--verify", "--quiet", branch)

	if err == nil {
		return true, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}

func CreateAndCheckoutBranch(path, branch string) error {
	_, err := runGit(path, "checkout", "-b", branch)
	return err
}

func StagePath(repoPath, relativePath string) error {
	_, err := runGit(repoPath, "add", relativePath)
	return err
}

func Commit(repoPath, message string) error {
	_, err := runGit(repoPath, "commit", "-m", message)
	return err
}
