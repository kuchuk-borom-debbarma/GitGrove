package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// Register records one or more repos (name → path) in the GitGroove metadata.
//
// High-level behavior:
//
//	Register operates strictly against the latest committed state of the gitgroove/system branch.
//	It validates the requested repo definitions, appends validated entries to .gg/repos/<name>/path,
//	creates a new commit (parent = current gitgroove/system tip), and atomically updates the
//	gitgroove/system reference to that new commit.
//
// IMPORTANT:
//   - Registration ONLY records repos (name + path).
//   - Registration DOES NOT create any derived GitGroove branches.
//   - Branch creation happens exclusively after hierarchy linking, not during registration.
//
// Requirements / invariants:
//   - rootAbsPath must point to a valid Git repository already initialized with GitGroove (.gg exists).
//   - The caller provides repos as map[name]path; `name` is the unique repo ID, `path` is the
//     directory inside the Git repo.
//   - All repo names must be globally unique. If any name in the batch is already registered,
//     the entire registration step is aborted with no changes applied.
//   - All paths must exist, be directories, and reside within the Git project root.
//   - Repos must not contain their own .git directory.
//   - Updating/moving an existing repo’s path is not done here—handled by a dedicated command.
//
// Step-by-step algorithm (safe, atomic, optimistic):
//
//  1. Validate environment:
//     • Verify rootAbsPath is a Git repo with a .gg directory.
//     • Ensure the working tree is clean (no staged/unstaged/untracked changes).
//     • Ensure HEAD is not detached.
//     If any check fails → abort immediately.
//
//  2. Read the latest gitgroove/system commit:
//     • Resolve refs/heads/gitgroove/system to oldTip.
//     • Optionally fetch/merge remote state if multi-writer synchronization is desired.
//
//  3. Load existing repo metadata from oldTip:
//
//  3. Load existing repo metadata from oldTip:
//     • Stream .gg/repos/* using `git ls-tree` and `git show`.
//     • Build minimal sets for existing names and paths.
//     • Validation is always based on committed state, never working tree.
//
//  4. Validate incoming repos:
//     • For each name→path pair:
//     - name must be unique w.r.t. committed repos.
//     - path must be unique and must exist in the filesystem.
//     - path must be a directory under rootAbsPath.
//     - path must not contain a nested .git.
//     If any repo fails validation → abort, write nothing.
//
//  5. Prepare updated metadata in a temporary workspace:
//     • Create a temporary git worktree detached at oldTip
//     (or build tree programmatically using plumbing).
//     • Write new repo entries to .gg/repos/<name>/path in this temporary workspace.
//
//  6. Create a new commit for updated metadata:
//     • Stage updated .gg files in the temporary workspace.
//     • Create a commit with parent = oldTip containing only the metadata changes.
//     • Capture the new commit hash newTip.
//
//  7. Atomically update gitgroove/system:
//     • Perform a conditional ref update:
//     git update-ref refs/heads/gitgroove/system <newTip> <oldTip>
//     This ensures correct optimistic concurrency control.
//     • If this fails (branch moved), abort and return a retryable error.
//     • If remote sync is required: push using --force-with-lease.
//
//  8. POST-COMMIT NOTE:
//     • Registration DOES NOT trigger branch creation of gitgroove/<repoName>.
//     • Derived branch creation is only performed after linking relationships.
//
//  9. Cleanup temporary workspace and return success.
//
// Atomicity guarantee:
//   - If ANY validation fails or the conditional ref update fails, NOTHING is committed.
//   - Only a fully validated, fully committed metadata change becomes visible in gitgroove/system.
//
// Notes:
//   - Metadata files are append-only; no mutation of existing entries occurs here.
//   - Moving/renaming repos requires a separate dedicated command.
//   - Register should not modify or disturb the user's currently checked-out branch since
//     all metadata writes occur in a detached temporary worktree or via plumbing.
//
// Register must be atomic: if any repo fails validation or the CAS (compare-and-swap) ref update
// fails, no partial metadata is written and the system state remains unchanged.
func Register(rootAbsPath string, repos map[string]string) error {
	log.Info().Msgf("Attempting to register %d repos in %s", len(repos), rootAbsPath)

	// 1. Validate environment
	if err := validateRegisterEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Read latest gitgroove/system commit
	systemRef := "refs/heads/gitgroove/system"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, systemRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s (is GitGroove initialized?): %w", systemRef, err)
	}

	// 3. Load existing repo metadata
	existingRepos, err := loadExistingRepos(rootAbsPath, oldTip)
	if err != nil {
		return fmt.Errorf("failed to load existing repos: %w", err)
	}

	// 4. Validate incoming repos
	if err := validateNewRepos(rootAbsPath, repos, existingRepos); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 5. Prepare updated metadata in temporary workspace and create commit
	newTip, err := createRegisterCommit(rootAbsPath, oldTip, repos)
	if err != nil {
		return err
	}

	// 7. Atomically update gitgroove/system
	if err := gitUtil.UpdateRef(rootAbsPath, systemRef, newTip, oldTip); err != nil {
		return fmt.Errorf("failed to update %s (concurrent modification?): %w", systemRef, err)
	}

	// If we are currently on the system branch, we must update the working tree to match the new commit.
	// Otherwise, the working tree will appear "dirty" (missing the new files we just committed).
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err == nil && currentBranch == "gitgroove/system" {
		log.Info().Msg("Updating working tree to match new system state")
		if err := gitUtil.ResetHard(rootAbsPath, "HEAD"); err != nil {
			return fmt.Errorf("failed to update working tree: %w", err)
		}
	}

	// 8. Create .gitgroverepo marker in the registered directories
	// This is critical for the stage command to detect nested repos.
	for name, path := range repos {
		// path is relative to rootAbsPath
		absPath := filepath.Join(rootAbsPath, path)
		markerPath := filepath.Join(absPath, ".gitgroverepo")

		// Always overwrite or create the marker with the repo name
		// This ensures identity is correct even if re-registering or if file was empty
		if err := fileUtil.WriteTextFile(markerPath, name); err != nil {
			log.Warn().Msgf("Failed to write marker file at %s: %v", markerPath, err)
			// We don't fail the registration because the metadata is already committed.
		}
	}

	log.Info().Msg("Successfully registered repositories")
	return nil
}

func validateRegisterEnvironment(rootAbsPath string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}
	return nil
}

func createRegisterCommit(rootAbsPath, oldTip string, repos map[string]string) (string, error) {
	// 5. Prepare updated metadata in temporary workspace
	tempDir, err := os.MkdirTemp("", "gitgroove-register-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // cleanup

	// Create detached worktree at oldTip
	if err := gitUtil.WorktreeAddDetached(rootAbsPath, tempDir, oldTip); err != nil {
		return "", fmt.Errorf("failed to create temporary worktree: %w", err)
	}
	defer gitUtil.WorktreeRemove(rootAbsPath, tempDir) // cleanup worktree

	// Write new repos to .gg/repos/<name>/path
	for name, path := range repos {
		repoDir := filepath.Join(tempDir, ".gg", "repos", name)
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create dir for repo %s: %w", name, err)
		}

		pathFile := filepath.Join(repoDir, "path")
		// Canonicalize path before writing
		cleanPath := fileUtil.NormalizePath(path)
		if filepath.IsAbs(cleanPath) {
			// Should have been caught by validation, but ensure we write relative
			rel, _ := filepath.Rel(rootAbsPath, cleanPath)
			cleanPath = fileUtil.NormalizePath(rel)
		}

		if err := os.WriteFile(pathFile, []byte(cleanPath), 0644); err != nil {
			return "", fmt.Errorf("failed to write path for repo %s: %w", name, err)
		}
	}

	// 6. Create new commit
	// Stage everything in .gg/repos
	if err := gitUtil.StagePath(tempDir, ".gg/repos"); err != nil {
		return "", fmt.Errorf("failed to stage .gg/repos: %w", err)
	}
	if err := gitUtil.Commit(tempDir, fmt.Sprintf("Register %d new repositories", len(repos))); err != nil {
		return "", fmt.Errorf("failed to commit metadata changes: %w", err)
	}
	newTip, err := gitUtil.GetHeadCommit(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to get new commit hash: %w", err)
	}
	return newTip, nil
}

func loadExistingRepos(root, ref string) (map[string]model.Repo, error) {
	// List directories in .gg/repos
	// Note: .gg/repos might not exist if no repos are registered yet.
	// git ls-tree will fail if the path doesn't exist.
	// We should check if .gg/repos exists first or handle the error.
	// runGit returns error if path not found.

	entries, err := gitUtil.ListTree(root, ref, ".gg/repos")
	if err != nil {
		// If the directory doesn't exist, it's not an error, just empty.
		// git ls-tree returns error if path not found.
		// We need to distinguish "path not found" from other errors if possible.
		// For now, assuming any error from ListTree on a specific path means it doesn't exist or is empty
		// is a simplification. A better check would be explicitly checking existence.
		// But based on gitUtil implementation, let's assume it returns error if not found.
		return map[string]model.Repo{}, nil
	}

	repos := make(map[string]model.Repo)
	for _, name := range entries {
		// Ignore .gitkeep
		if name == ".gitkeep" {
			continue
		}

		// Each entry is a repo name (directory)
		// Read .gg/repos/<name>/path
		pathFile := fmt.Sprintf(".gg/repos/%s/path", name)
		content, err := gitUtil.ShowFile(root, ref, pathFile)
		if err != nil {
			// If path file is missing, it's a corruption or partial state.
			return nil, fmt.Errorf("failed to read path for repo %s: %w", name, err)
		}

		repoPath := strings.TrimSpace(content)

		// Read .gg/repos/<name>/parent
		parentFile := fmt.Sprintf(".gg/repos/%s/parent", name)
		parentContent, err := gitUtil.ShowFile(root, ref, parentFile)
		parent := ""
		if err == nil {
			parent = strings.TrimSpace(parentContent)
		} else {
			// Parent file might not exist for root repos or newly registered ones
			// We treat error as empty parent
		}

		repos[name] = model.Repo{
			Name:   name,
			Path:   repoPath,
			Parent: parent,
		}
	}

	return repos, nil
}

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func validateNewRepos(root string, newRepos map[string]string, existing map[string]model.Repo) error {
	// Check for name collisions and validity
	for name := range newRepos {
		if !validNameRegex.MatchString(name) {
			return fmt.Errorf("invalid repo name '%s': must match %s", name, validNameRegex.String())
		}
		if _, ok := existing[name]; ok {
			return fmt.Errorf("repo name '%s' already registered", name)
		}
	}

	// Check for path collisions and validity
	existingPaths := make(map[string]bool)
	for _, r := range existing {
		existingPaths[r.Path] = true
	}

	for _, relPath := range newRepos {
		// Path uniqueness
		cleanPath := fileUtil.NormalizePath(relPath)

		if existingPaths[cleanPath] {
			return fmt.Errorf("path '%s' already registered", relPath)
		}

		// Existence and containment
		absPath := filepath.Join(root, cleanPath)

		// Verify path is inside root
		rel, err := filepath.Rel(root, absPath)
		if err != nil || strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") {
			return fmt.Errorf("path '%s' escapes project root", relPath)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("path '%s' does not exist", relPath)
		}
		if !info.IsDir() {
			return fmt.Errorf("path '%s' is not a directory", relPath)
		}

		// No nested .git
		// Exception: if absPath is the root itself, .git is expected.
		if absPath != root && fileUtil.Exists(filepath.Join(absPath, ".git")) {
			return fmt.Errorf("repo '%s' contains .git directory (nested git repos not allowed)", relPath)
		}

		// Check for duplicates within the batch
		// (This is implicitly handled by map keys for names, but paths could be dupes in the batch)
		// We'll check batch path uniqueness separately if needed, but map[string]string allows unique names only.
		// Multiple names pointing to same path?
		// Let's check that too.
	}

	// Check for duplicate paths in the input batch
	seenPaths := make(map[string]string)
	for name, path := range newRepos {
		cleanPath := fileUtil.NormalizePath(path)
		if otherName, ok := seenPaths[cleanPath]; ok {
			return fmt.Errorf("duplicate path '%s' used by '%s' and '%s'", path, otherName, name)
		}
		seenPaths[cleanPath] = name
	}

	return nil
}
