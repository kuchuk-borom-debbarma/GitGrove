package grove

import (
	"fmt"
	"strings"

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
	repo, ok := existingRepos[repoName]
	if !ok {
		return fmt.Errorf("repo '%s' not found", repoName)
	}

	// 5. Construct branch ref path
	// New simplified branch naming: refs/heads/gitgroove/repos/<repoName>/branches/<branchName>
	fullRefPath := RepoBranchRef(repoName, branchName)

	// 6. Get HEAD commit
	headCommit, err := gitUtil.GetHeadCommit(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get project HEAD: %w", err)
	}

	// 7. Create the flattened branch commit
	// Instead of pointing directly to HEAD, we want to create a commit that has
	// the REPO's subtree as its root tree.

	// Get the tree hash of the repo's path within the current HEAD
	var repoTreeHash string
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err == nil && strings.Contains(currentBranch, fmt.Sprintf("gitgroove/repos/%s/branches/", repoName)) {
		// We are already on a branch of this repo, so the HEAD tree IS the repo tree (flattened)
		repoTreeHash, err = gitUtil.GetCommitTree(rootAbsPath, headCommit)
		if err != nil {
			return fmt.Errorf("failed to get tree from commit %s: %w", headCommit, err)
		}
	} else {
		// We are likely on the root or another repo, so we need to find the subtree
		repoTreeHash, err = gitUtil.GetSubtreeHash(rootAbsPath, headCommit, repo.Path)
		if err != nil {
			return fmt.Errorf("failed to get subtree hash for repo %s at %s: %w", repoName, repo.Path, err)
		}
	}

	// Create a new commit with this tree, using HEAD as parent (to keep history linkage)
	// Note: This creates a "synthetic" commit history for the branch.
	newCommitHash, err := gitUtil.CommitTree(rootAbsPath, repoTreeHash, fmt.Sprintf("Branch %s for repo %s", branchName, repoName), headCommit)
	if err != nil {
		return fmt.Errorf("failed to create commit for branch: %w", err)
	}

	// 8. Create the branch ref
	log.Info().Msgf("Creating branch ref: %s -> %s", fullRefPath, newCommitHash)
	if err := gitUtil.SetRef(rootAbsPath, fullRefPath, newCommitHash); err != nil {
		return fmt.Errorf("failed to create branch ref: %w", err)
	}

	return nil
}
