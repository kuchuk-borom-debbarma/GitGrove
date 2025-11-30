package grove

import (
	"fmt"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// SwitchToSystem switches the working tree to the gitgroove/internal branch.
// This provides the "System Root" view where all root repositories are visible.
func SwitchToSystem(rootAbsPath string) error {
	log.Info().Msg("Switching to System Root view...")

	// 1. Validate environment
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}

	// 2. Checkout system branch
	if err := gitUtil.Checkout(rootAbsPath, "gitgroove/internal"); err != nil {
		return fmt.Errorf("failed to checkout system branch: %w", err)
	}

	log.Info().Msg("Successfully switched to System Root")

	// 3. Ensure clean state (User Request: ALWAYS CLEAN)
	if err := gitUtil.ResetHard(rootAbsPath, "HEAD"); err != nil {
		return fmt.Errorf("failed to reset hard: %w", err)
	}
	if err := gitUtil.CleanFD(rootAbsPath); err != nil {
		return fmt.Errorf("failed to clean -fd: %w", err)
	}

	return nil
}
