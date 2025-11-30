package grove

import (
	"fmt"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// CheckoutRepo switches the user's working tree to a specific branch of a nested repository.
//
// It uses the flat branch naming structure:
// gitgroove/repos/<repoName>/branches/<branchName>
func CheckoutRepo(rootAbsPath, repoName, branchName string, keepEmptyDirs bool) error {
	// 1. Validate environment
	if err := validateSwitchEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Checkout gitgroove/system to load authoritative metadata
	log.Info().Msg("Checking out gitgroove/system to load metadata")
	if err := gitUtil.Checkout(rootAbsPath, "gitgroove/system"); err != nil {
		return fmt.Errorf("failed to checkout gitgroove/system: %w", err)
	}

	// 3. Load existing repos from the system branch
	repos, err := loadExistingRepos(rootAbsPath, "HEAD")
	if err != nil {
		return fmt.Errorf("failed to load repos from system branch: %w", err)
	}

	// 4. Find target repo
	_, ok := repos[repoName]
	if !ok {
		return fmt.Errorf("repo '%s' not found in metadata", repoName)
	}

	// 5. Construct branch ref path
	// New simplified branch naming: refs/heads/gitgroove/repos/<repoName>/branches/<branchName>
	fullRefPath := RepoBranchRef(repoName, branchName)

	// 6. Validate branch exists
	if !gitUtil.RefExists(rootAbsPath, fullRefPath) {
		return fmt.Errorf("target branch '%s' does not exist", fullRefPath)
	}

	// 7. Checkout the branch
	// We use the short name (without refs/heads/) to attach to the branch
	shortBranchName := RepoBranchShortFromRef(fullRefPath)
	log.Info().Msgf("Switching to %s", shortBranchName)
	if err := gitUtil.Checkout(rootAbsPath, shortBranchName); err != nil {
		return fmt.Errorf("failed to checkout target branch: %w", err)
	}

	// 8. Clean up empty directories if requested
	if !keepEmptyDirs {
		log.Info().Msg("Cleaning up empty directories...")
		if err := file.CleanEmptyDirsRecursively(rootAbsPath); err != nil {
			// Log warning but don't fail the checkout
			log.Warn().Err(err).Msg("Failed to clean up empty directories")
		}
	}

	log.Info().Msgf("Successfully switched to repo '%s' on branch '%s'", repoName, branchName)
	return nil
}
