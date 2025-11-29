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
	_, err := runGit(repoPath, "add", "-f", relativePath)
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

// SetRef updates a ref to a new value unconditionally.
//
// It runs `git update-ref <ref> <newHash>`.
func SetRef(repoPath, ref, newHash string) error {
	_, err := runGit(repoPath, "update-ref", ref, newHash)
	return err
}

func GetHeadCommit(repoPath string) (string, error) {
	return runGit(repoPath, "rev-parse", "HEAD")
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

func GetCurrentBranch(repoPath string) (string, error) {
	return runGit(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}

func Checkout(repoPath, branch string) error {
	_, err := runGit(repoPath, "checkout", branch)
	return err
}

func CheckoutPath(repoPath, ref, path string) error {
	_, err := runGit(repoPath, "checkout", ref, "--", path)
	return err
}

func UnstagePath(repoPath, path string) error {
	_, err := runGit(repoPath, "restore", "--staged", path)
	return err
}

func ResetHard(repoPath, ref string) error {
	_, err := runGit(repoPath, "reset", "--hard", ref)
	return err
}

// RefExists checks if a reference exists.
func RefExists(repoPath, ref string) bool {
	_, err := runGit(repoPath, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

// Init initializes a new git repository.
func Init(repoPath string) error {
	_, err := runGit(repoPath, "init")
	return err
}

// CreateBranch creates a new branch pointing to the specified start point.
func CreateBranch(repoPath, branch, startPoint string) error {
	_, err := runGit(repoPath, "branch", branch, startPoint)
	return err
}

// RunGit runs an arbitrary git command.
func RunGit(repoPath string, args ...string) (string, error) {
	return runGit(repoPath, args...)
}

// CommitTree creates a commit from a tree object.
func CommitTree(repoPath, treeHash, message string, parents ...string) (string, error) {
	args := []string{"commit-tree", treeHash, "-m", message}
	for _, p := range parents {
		args = append(args, "-p", p)
	}
	return runGit(repoPath, args...)
}

// GetEmptyTreeHash returns the hash of an empty tree.
func GetEmptyTreeHash(repoPath string) (string, error) {
	// The empty tree hash is a constant in git: 4b825dc642cb6eb9a060e54bf8d69288fbee4904
	return "4b825dc642cb6eb9a060e54bf8d69288fbee4904", nil
}
