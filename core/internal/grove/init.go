package grove

import (
	"fmt"
	"path/filepath"

	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// Init initializes GitGroove on the current Git repository.
func Init(absolutePath string) error {
	log.Debug().Msg("Attempting to initialize GitGroove in path " + absolutePath)

	// Normalize path
	absolutePath = fileUtil.NormalizePath(absolutePath)
	ggPath := filepath.Join(absolutePath, ".gg")

	// 1. Ensure .gg does NOT exist
	if err := fileUtil.EnsureNotExists(ggPath); err != nil {
		return fmt.Errorf("GitGroove already initialized: %w", err)
	}

	// 2. Ensure repo exists OR init one
	if git.IsInsideGitRepo(absolutePath) {
		exists, err := git.HasBranch(absolutePath, "gitgroove/system")
		if err != nil {
			return fmt.Errorf("failed checking existing system branch: %w", err)
		}
		if exists {
			return fmt.Errorf("'gitgroove/system' already exists — GitGroove may already be initialized")
		}
	} else {
		if err := git.GitInit(absolutePath); err != nil {
			return fmt.Errorf("failed to run git init: %w", err)
		}
		log.Info().Msg("Initialized empty git repository")
	}

	// 3. Repo MUST be clean
	if err := git.VerifyCleanState(absolutePath); err != nil {
		return fmt.Errorf("git repository not clean: %w", err)
	}

	fmt.Println("Initializing GitGroove in", absolutePath)

	// 4. Create .gg directory
	if err := fileUtil.CreateDir(ggPath); err != nil {
		return fmt.Errorf("failed creating .gg: %w", err)
	}
	log.Info().Msg("Created .gg workspace directory")

	// 5. Create grove.json (empty)
	groveFile := filepath.Join(ggPath, "grove.json")
	if err := fileUtil.WriteJSONFile(groveFile, map[string]any{}); err != nil {
		return fmt.Errorf("failed to create grove.json: %w", err)
	}
	log.Info().Msg("Created grove.json")

	// 6. Create / checkout system branch
	if err := git.CreateAndCheckoutBranch(absolutePath, "gitgroove/system"); err != nil {
		return fmt.Errorf("failed creating gitgrove/system: %w", err)
	}
	log.Info().Msg("Checked out gitgrove/system")

	// 7. Stage .gg
	if err := git.StagePath(absolutePath, ".gg"); err != nil {
		return fmt.Errorf("failed staging .gg directory: %w", err)
	}

	// 8. Commit
	if err := git.Commit(absolutePath, "Initialize GitGroove system branch"); err != nil {
		return fmt.Errorf("failed committing GitGroove metadata: %w", err)
	}
	log.Info().Msg("Committed .gg metadata")

	return nil
}
