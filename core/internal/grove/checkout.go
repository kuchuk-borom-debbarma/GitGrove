package grove

import (
	"fmt"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// CheckoutRepo switches the user's working tree to a specific branch of a nested repository.
//
// It supports the branch path structure:
// gitgroove/repos/<a>/children/<b>/branches/<branchName>
func CheckoutRepo(rootAbsPath, repoName, branchName string) error {
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
	targetRepo, ok := repos[repoName]
	if !ok {
		return fmt.Errorf("repo '%s' not found in metadata", repoName)
	}

	// 5. Build ancestry chain: child -> parent -> ... -> root
	var ancestry []model.Repo
	current := targetRepo
	for {
		ancestry = append(ancestry, current)
		if current.Parent == "" {
			break
		}
		parent, ok := repos[current.Parent]
		if !ok {
			return fmt.Errorf("broken ancestry: parent '%s' of '%s' missing", current.Parent, current.Name)
		}
		current = parent
	}

	// 6. Construct branch ref path
	// Reverse ancestry to get root -> ... -> child
	// Format: gitgroove/repos/<root>/children/<child1>/children/<child2>/branches/<branchName>
	var pathSegments []string

	// Start with root
	rootRepo := ancestry[len(ancestry)-1]
	pathSegments = append(pathSegments, rootRepo.Name)

	// Append children recursively
	for i := len(ancestry) - 2; i >= 0; i-- {
		pathSegments = append(pathSegments, "children", ancestry[i].Name)
	}

	fullRefPath := fmt.Sprintf("refs/heads/gitgroove/repos/%s/branches/%s", strings.Join(pathSegments, "/"), branchName)

	// 7. Validate branch exists
	if !gitUtil.RefExists(rootAbsPath, fullRefPath) {
		return fmt.Errorf("target branch '%s' does not exist", fullRefPath)
	}

	// 8. Checkout the branch
	// We use the short name (without refs/heads/) to attach to the branch
	shortBranchName := strings.TrimPrefix(fullRefPath, "refs/heads/")
	log.Info().Msgf("Switching to %s", shortBranchName)
	if err := gitUtil.Checkout(rootAbsPath, shortBranchName); err != nil {
		return fmt.Errorf("failed to checkout target branch: %w", err)
	}

	log.Info().Msgf("Successfully switched to repo '%s' on branch '%s'", repoName, branchName)
	return nil
}
