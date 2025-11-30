package grove

import (
	"fmt"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// validateCleanGitRepo validates that the given path is a clean git repository.
// This is a common validation used across multiple GitGrove operations.
func validateCleanGitRepo(rootAbsPath string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}
	return nil
}
