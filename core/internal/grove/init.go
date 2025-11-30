package grove

import (
	"fmt"
	"os"
	"path/filepath"

	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// Init initializes GitGroove on the current Git repository.
//
// High-level behavior:
//
//	Init bootstraps the GitGroove metadata structure within an existing Git repository.
//	It creates the hidden .gg directory, initializes the .gg/repos structure, and establishes
//	the detached gitgroove/internal branch to track metadata history.
//
// Requirements / invariants:
//   - absolutePath must point to a valid, existing Git repository.
//   - The repository working tree must be 100% clean (no staged, unstaged, or untracked changes)
//     to ensure safe branch creation and switching.
//   - The .gg directory must not already exist (idempotency check).
//   - The gitgroove/internal branch must not already exist.
//
// Step-by-step algorithm:
//
//  1. Validate environment:
//     • Normalize the path.
//     • Verify it is a Git repository.
//     • Verify the working tree is clean.
//     • Verify .gg does not exist.
//     • Verify gitgroove/internal branch does not exist.
//     If any check fails → abort immediately.
//
//  2. Create metadata directory structure:
//     • Create .gg directory.
//     • Create .gg/repos directory.
//     • Create .gg/repos/.gitkeep to ensure git tracks the directory even if empty.
//
//  3. Initialize internal branch:
//     • Create and checkout a new orphan-like branch 'gitgroove/internal'.
//     (Note: In this implementation, it branches off the current HEAD, effectively making
//     history shared until this point, or it might be intended as an orphan.
//     The current implementation uses `checkout -b`, which branches from current HEAD.)
//
//  4. Commit initial state:
//     • Stage the .gg directory.
//     • Commit with message "Initialize GitGroove internal branch".
//
//  5. Result:
//     • The user is left on the gitgroove/internal branch (based on current implementation).
//     • .gg exists and is tracked.
//
// Atomicity:
//   - This operation is not fully atomic in the ACID sense (filesystem changes happen before git ops).
//   - However, checks are performed upfront to minimize failure risk.
//   - If it fails midway, the user might be left with a partial .gg directory or a new branch.
func Init(absolutePath string) error {
	log.Debug().Msg("Attempting to initialize GitGroove in path " + absolutePath)

	// Normalize
	absolutePath = fileUtil.NormalizePath(absolutePath)
	ggPath := filepath.Join(absolutePath, ".gg")

	if err := validateInitEnvironment(absolutePath, ggPath); err != nil {
		return err
	}

	fmt.Println("Initializing GitGroove in", absolutePath)

	if err := createMetadataDirectories(ggPath); err != nil {
		return err
	}

	if err := initializeInternalBranch(absolutePath); err != nil {
		return err
	}

	log.Info().Msg("GitGroove initialized successfully")

	return nil
}

// IsInitialized checks if GitGrove is initialized in the given path.
// It returns true if the .gg directory exists.
func IsInitialized(absolutePath string) (bool, error) {
	// Normalize
	absolutePath = fileUtil.NormalizePath(absolutePath)
	ggPath := filepath.Join(absolutePath, ".gg")

	_, err := os.Stat(ggPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check for .gg directory: %w", err)
}

func validateInitEnvironment(absolutePath, ggPath string) error {
	// MUST be an existing git repo
	if !git.IsInsideGitRepo(absolutePath) {
		return fmt.Errorf("GitGroove cannot initialize: not a valid Git repository")
	}

	// MUST be clean
	if err := git.VerifyCleanState(absolutePath); err != nil {
		return fmt.Errorf("GitGroove cannot initialize: %w", err)
	}

	// .gg must NOT exist
	if err := fileUtil.EnsureNotExists(ggPath); err != nil {
		return fmt.Errorf("GitGroove already initialized: %w", err)
	}

	// Ensure internal branch does NOT already exist
	exists, err := git.HasBranch(absolutePath, "gitgroove/internal")
	if err != nil {
		return fmt.Errorf("failed checking internal branch: %w", err)
	}
	if exists {
		return fmt.Errorf("gitgroove/internal branch already exists — GitGroove may already be initialized")
	}
	return nil
}

func createMetadataDirectories(ggPath string) error {
	// Create .gg
	if err := fileUtil.CreateDir(ggPath); err != nil {
		return fmt.Errorf("failed to create .gg: %w", err)
	}
	log.Info().Msg("Created .gg workspace directory")

	// Create .gg/repos
	reposPath := filepath.Join(ggPath, "repos")
	if err := fileUtil.CreateDir(reposPath); err != nil {
		return fmt.Errorf("failed to create .gg/repos: %w", err)
	}
	log.Info().Msg("Created .gg/repos directory")

	// Create .gitkeep to ensure repos dir is tracked
	gitKeepFile := filepath.Join(reposPath, ".gitkeep")
	if err := fileUtil.WriteTextFile(gitKeepFile, ""); err != nil {
		return fmt.Errorf("failed to create .gitkeep: %w", err)
	}
	return nil
}

func initializeInternalBranch(absolutePath string) error {
	// Create and checkout internal branch
	if err := git.CreateAndCheckoutBranch(absolutePath, "gitgroove/internal"); err != nil {
		return fmt.Errorf("failed creating internal branch: %w", err)
	}

	// Stage .gg
	if err := git.StagePath(absolutePath, ".gg"); err != nil {
		return fmt.Errorf("failed staging .gg: %w", err)
	}

	// Commit
	if err := git.Commit(absolutePath, "Initialize GitGroove internal branch"); err != nil {
		return fmt.Errorf("failed committing metadata: %w", err)
	}
	return nil
}
