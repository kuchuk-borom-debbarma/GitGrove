package grove

import (
	"fmt"
	"path/filepath"

	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// Init initializes GitGroove on the current Git repository.
//
// Requirements:
// - Must be inside an already initialized Git repo
// - Repo must be 100% clean (no staged/unstaged/untracked)
// - .gg must not already exist
func Init(absolutePath string) error {
	log.Debug().Msg("Attempting to initialize GitGroove in path " + absolutePath)

	// Normalize
	absolutePath = fileUtil.NormalizePath(absolutePath)
	ggPath := filepath.Join(absolutePath, ".gg")

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

	// Ensure system branch does NOT already exist
	exists, err := git.HasBranch(absolutePath, "gitgroove/system")
	if err != nil {
		return fmt.Errorf("failed checking system branch: %w", err)
	}
	if exists {
		return fmt.Errorf("gitgroove/system branch already exists — GitGroove may already be initialized")
	}

	fmt.Println("Initializing GitGroove in", absolutePath)

	// Create .gg
	if err := fileUtil.CreateDir(ggPath); err != nil {
		return fmt.Errorf("failed to create .gg: %w", err)
	}
	log.Info().Msg("Created .gg workspace directory")

	// Create grove.json
	groveFile := filepath.Join(ggPath, "grove.json")
	if err := fileUtil.WriteJSONFile(groveFile, map[string]any{}); err != nil {
		return fmt.Errorf("failed to create grove.json: %w", err)
	}
	log.Info().Msg("Created grove.json")

	// Create and checkout system branch
	if err := git.CreateAndCheckoutBranch(absolutePath, "gitgroove/system"); err != nil {
		return fmt.Errorf("failed creating system branch: %w", err)
	}

	// Stage .gg
	if err := git.StagePath(absolutePath, ".gg"); err != nil {
		return fmt.Errorf("failed staging .gg: %w", err)
	}

	// Commit
	if err := git.Commit(absolutePath, "Initialize GitGroove system branch"); err != nil {
		return fmt.Errorf("failed committing metadata: %w", err)
	}

	log.Info().Msg("GitGroove initialized successfully")

	return nil
}
