package git

import (
	"bytes"
	"errors"
	"io"
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

// IsInsideGitRepo checks if the given path is inside a Git working tree.
//
// It runs `git rev-parse --is-inside-work-tree`.
// Returns true only if the command succeeds and outputs "true".
func IsInsideGitRepo(path string) bool {
	out, err := runGit(path, "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// IsDetachedHEAD checks if HEAD is detached (not pointing to a branch ref).
//
// It runs `git rev-parse --symbolic-full-name HEAD`.
// If HEAD is detached, the output is "HEAD". If attached, it returns the full ref (e.g., "refs/heads/main").
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

// VerifyCleanState ensures the repository is in a clean state suitable for critical operations.
//
// It performs a comprehensive check:
//  1. Is HEAD detached? (We generally require being on a branch for safety, though some ops might work detached).
//  2. Are there staged changes? (git diff --cached --quiet)
//  3. Are there unstaged changes? (git diff --quiet)
//  4. Are there untracked files? (git ls-files --others --exclude-standard)
//
// Returns nil if clean, or an error detailing all found issues.
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

// HasBranch checks if a local branch exists.
//
// It uses `git rev-parse --verify --quiet <branch>`.
// Returns true if the branch ref exists, false otherwise.
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

func ResolveRef(repoPath, ref string) (string, error) {
	return runGit(repoPath, "rev-parse", ref)
}

func ShowFile(repoPath, ref, filePath string) (string, error) {
	return runGit(repoPath, "show", ref+":"+filePath)
}

// WorktreeAddDetached creates a temporary linked worktree detached at a specific commit.
//
// This is used for safe metadata manipulation without disturbing the user's primary working tree.
// It runs `git worktree add --detach <worktreePath> <ref>`.
func WorktreeAddDetached(repoPath, worktreePath, ref string) error {
	_, err := runGit(repoPath, "worktree", "add", "--detach", worktreePath, ref)
	return err
}

func WorktreeRemove(repoPath, worktreePath string) error {
	_, err := runGit(repoPath, "worktree", "remove", "--force", worktreePath)
	return err
}

// UpdateRef updates a ref to a new value, but ONLY if it currently matches oldHash.
//
// This implements Optimistic Concurrency Control (CAS - Compare And Swap).
// It runs `git update-ref <ref> <newHash> <oldHash>`.
// If the ref has changed since oldHash was read, this command fails, preventing race conditions.
func UpdateRef(repoPath, ref, newHash, oldHash string) error {
	_, err := runGit(repoPath, "update-ref", ref, newHash, oldHash)
	return err
}

func GetHeadCommit(repoPath string) (string, error) {
	return runGit(repoPath, "rev-parse", "HEAD")
}

func StreamFile(repoPath, ref, filePath string) (io.ReadCloser, error) {
	cmd := exec.Command("git", "show", ref+":"+filePath)
	cmd.Dir = repoPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// We return a ReadCloser that waits for command completion on Close
	return &cmdReadCloser{pipe: stdout, cmd: cmd}, nil
}

type cmdReadCloser struct {
	pipe io.ReadCloser
	cmd  *exec.Cmd
}

func (c *cmdReadCloser) Read(p []byte) (n int, err error) {
	return c.pipe.Read(p)
}

func (c *cmdReadCloser) Close() error {
	// Close pipe first
	c.pipe.Close()
	// Wait for command to finish
	return c.cmd.Wait()
}

func ListTree(repoPath, ref, path string) ([]string, error) {
	out, err := runGit(repoPath, "ls-tree", "--name-only", ref+":"+path)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	return strings.Split(out, "\n"), nil
}
