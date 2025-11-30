package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// CheckoutRepo switches the user's working tree to a specific branch of a nested repository.
//
// It uses the flat branch naming structure:
// gitgroove/repos/<repoName>/branches/<branchName>
func CheckoutRepo(rootAbsPath, repoName, branchName string, keepEmptyDirs, flat bool) error {
	// 1. Validate environment
	if err := validateSwitchEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Checkout gitgroove/internal to load authoritative metadata
	log.Info().Msg("Checking out gitgroove/internal to load metadata")
	if err := gitUtil.Checkout(rootAbsPath, InternalBranchName); err != nil {
		return fmt.Errorf("failed to checkout gitgroove/internal: %w", err)
	}

	// 3. Load existing repos from the system branch
	repos, err := loadExistingRepos(rootAbsPath, "HEAD")
	if err != nil {
		return fmt.Errorf("failed to load repos from internal branch: %w", err)
	}

	// 4. Find target repo
	targetRepo, ok := repos[repoName]
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

	// 8. Configure sparse-checkout to hide nested repos
	if err := configureSparseCheckout(rootAbsPath, targetRepo.Path, repos, flat); err != nil {
		log.Warn().Err(err).Msg("Failed to configure sparse-checkout")
	}

	// 9. Clean up empty directories if requested
	if !keepEmptyDirs {
		log.Info().Msg("Cleaning up empty directories...")
		if err := file.CleanEmptyDirsRecursively(rootAbsPath); err != nil {
			// Log warning but don't fail the checkout
			log.Warn().Err(err).Msg("Failed to clean up empty directories")
		}
	}

	// 10. Ensure .gitgroverepo marker exists
	if err := ensureRepoMarker(rootAbsPath, repoName); err != nil {
		log.Warn().Err(err).Msg("Failed to ensure .gitgroverepo marker")
	}

	log.Info().Msgf("Successfully switched to repo '%s' on branch '%s'", repoName, branchName)
	return nil
}

func configureSparseCheckout(rootAbsPath, currentRepoPath string, repos map[string]model.Repo, flat bool) error {
	// If flat is true, we want to see everything (disable sparse-checkout)
	if flat {
		return gitUtil.DisableSparseCheckout(rootAbsPath)
	}

	// Find children repos
	var children []string
	for _, r := range repos {
		if r.Path == currentRepoPath {
			continue
		}
		// Check if r.Path is inside currentRepoPath
		rel, err := filepath.Rel(currentRepoPath, r.Path)
		if err == nil && !strings.HasPrefix(rel, "..") && rel != "." {
			// It's a child
			// Sparse-checkout patterns are relative to the root of the working tree.
			// Since we are checked out to the repo's branch, the working tree root IS the repo root.
			// So 'rel' is the correct path to exclude.
			// We need to ensure it ends with / to match directory
			if !strings.HasSuffix(rel, "/") {
				rel += "/"
			}
			children = append(children, rel)
		}
	}

	if len(children) == 0 {
		// No children to hide, disable sparse-checkout
		return gitUtil.DisableSparseCheckout(rootAbsPath)
	}

	// We have children to hide
	// Pattern:
	// /*
	// !child1/
	// !child2/
	patterns := []string{"/*"}
	for _, child := range children {
		patterns = append(patterns, "!"+child)
	}

	log.Info().Msgf("Hiding %d nested repositories...", len(children))
	return gitUtil.SetSparseCheckout(rootAbsPath, patterns)
}

func ensureRepoMarker(rootAbsPath, repoName string) error {
	markerPath := filepath.Join(rootAbsPath, ".gitgroverepo")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		log.Info().Msg("Restoring missing .gitgroverepo marker")
		return os.WriteFile(markerPath, []byte(repoName), 0644)
	}
	return nil
}
