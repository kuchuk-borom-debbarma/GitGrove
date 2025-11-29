package grove

import (
	"fmt"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// CreateRepoBranch creates a new branch for a specific nested repository.
//
// The branch path structure is:
// gitgroove/repos/<a>/children/<b>/branches/<branchName>
//
// This function:
// 1. Validates the environment.
// 2. Loads existing repos from gitgroove/system.
// 3. Constructs the ancestry chain for the target repo.
// 4. Creates the branch ref pointing to the current project HEAD.
func CreateRepoBranch(rootAbsPath, repoName, branchName string) error {
	log.Info().Msgf("Attempting to create branch '%s' for repo '%s'", branchName, repoName)

	// 1. Validate environment
	// Reuse validateLinkEnvironment as it checks for clean state and git repo
	if err := validateLinkEnvironment(rootAbsPath); err != nil {
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

	// 4. Find target repo
	targetRepo, ok := existingRepos[repoName]
	if !ok {
		return fmt.Errorf("repo '%s' not found", repoName)
	}

	// 5. Build ancestry chain: child -> parent -> ... -> root
	var ancestry []model.Repo
	current := targetRepo
	for {
		ancestry = append(ancestry, current)
		if current.Parent == "" {
			break
		}
		parent, ok := existingRepos[current.Parent]
		if !ok {
			return fmt.Errorf("repo '%s' has missing parent '%s'", current.Name, current.Parent)
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

	// 7. Get HEAD commit
	headCommit, err := gitUtil.GetHeadCommit(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get project HEAD: %w", err)
	}

	// 8. Create the branch ref
	log.Info().Msgf("Creating branch ref: %s -> %s", fullRefPath, headCommit)
	if err := gitUtil.SetRef(rootAbsPath, fullRefPath, headCommit); err != nil {
		return fmt.Errorf("failed to create branch ref: %w", err)
	}

	return nil
}
