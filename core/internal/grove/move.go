package grove

import (
	"fmt"
	"os"
	"path/filepath"

	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

/*
Move relocates a registered repository to a new path within the project.

It operates atomically on the gitgroove/internal branch to update metadata,
and physically moves the directory on disk.

High-Level Responsibility:
  - Validates environment and inputs.
  - Checks out gitgroove/internal to load authoritative metadata.
  - Verifies repo existence and new path availability.
  - Moves the directory on disk.
  - Updates .gg/repos/<name>/path.
  - Commits the change to gitgroove/internal.

Guarantees:
  - Atomic metadata update.
  - Preserves repo identity (name and branch remain unchanged).
  - Does NOT rebuild branches (Option B architecture).
*/
func Move(rootAbsPath, repoName, newRelPath string) error {
	log.Info().Msgf("Attempting to move repo '%s' to '%s'", repoName, newRelPath)

	// 1. Validate environment
	if err := validateMoveEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Resolve gitgroove/internal ref
	internalRef := "refs/heads/gitgroove/internal"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, internalRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", internalRef, err)
	}

	// 3. Load existing repos from system ref
	repos, err := loadExistingRepos(rootAbsPath, oldTip)
	if err != nil {
		return fmt.Errorf("failed to load repos: %w", err)
	}

	// 4. Validate move operation
	repo, ok := repos[repoName]
	if !ok {
		return fmt.Errorf("repo '%s' not found", repoName)
	}

	// Normalize new path
	newRelPath = fileUtil.NormalizePath(newRelPath)
	if newRelPath == "" || newRelPath == "." || newRelPath == "/" {
		return fmt.Errorf("invalid destination path '%s'", newRelPath)
	}

	// Check if new path is already taken by another repo
	for _, r := range repos {
		if r.Path == newRelPath {
			return fmt.Errorf("path '%s' is already used by repo '%s'", newRelPath, r.Name)
		}
	}

	// Check if new path exists on disk (unless it's the same as old path, which is a no-op)
	if repo.Path == newRelPath {
		log.Info().Msg("Source and destination paths are identical. Nothing to do.")
		return nil
	}

	oldAbsPath := filepath.Join(rootAbsPath, repo.Path)
	newAbsPath := filepath.Join(rootAbsPath, newRelPath)

	if _, err := os.Stat(newAbsPath); err == nil {
		return fmt.Errorf("destination path '%s' already exists", newRelPath)
	}

	// 5. Perform physical move
	// Ensure parent dir of new path exists
	parentDir := filepath.Dir(newAbsPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory '%s': %w", parentDir, err)
	}

	log.Info().Msgf("Moving directory from '%s' to '%s'", repo.Path, newRelPath)
	if err := os.Rename(oldAbsPath, newAbsPath); err != nil {
		return fmt.Errorf("failed to move directory: %w", err)
	}

	// 6. Update metadata using temp worktree
	newTip, err := updateRepoPathInSystem(rootAbsPath, oldTip, repoName, newRelPath)
	if err != nil {
		// Attempt rollback of physical move
		_ = os.Rename(newAbsPath, oldAbsPath)
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	// 7. Atomically update gitgroove/internal
	if err := gitUtil.UpdateRef(rootAbsPath, internalRef, newTip, oldTip); err != nil {
		// Attempt rollback of physical move
		_ = os.Rename(newAbsPath, oldAbsPath)
		return fmt.Errorf("failed to update %s (concurrent modification?): %w", internalRef, err)
	}

	log.Info().Msg("Successfully moved repository")
	return nil
}

func updateRepoPathInSystem(rootAbsPath, oldTip, repoName, newPath string) (string, error) {
	tempDir, err := os.MkdirTemp("", "gitgroove-move-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := gitUtil.WorktreeAddDetached(rootAbsPath, tempDir, oldTip); err != nil {
		return "", fmt.Errorf("failed to create temporary worktree: %w", err)
	}
	defer gitUtil.WorktreeRemove(rootAbsPath, tempDir)

	// Update path file
	pathFile := filepath.Join(tempDir, ".gg", "repos", repoName, "path")
	if err := os.MkdirAll(filepath.Dir(pathFile), 0755); err != nil {
		return "", fmt.Errorf("failed to create repo metadata dir: %w", err)
	}
	if err := os.WriteFile(pathFile, []byte(newPath), 0644); err != nil {
		return "", fmt.Errorf("failed to write new path: %w", err)
	}

	// Commit
	if err := gitUtil.StagePath(tempDir, ".gg/repos"); err != nil {
		return "", fmt.Errorf("failed to stage metadata: %w", err)
	}
	if err := gitUtil.Commit(tempDir, fmt.Sprintf("Move repo '%s' to '%s'", repoName, newPath)); err != nil {
		return "", fmt.Errorf("failed to commit move: %w", err)
	}

	return gitUtil.GetHeadCommit(tempDir)
}

func validateMoveEnvironment(rootAbsPath string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}
	return nil
}
