package grove

import (
	"fmt"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// Push pushes the specified repositories to their configured remotes.
//
// It performs the following steps:
// 1. Validates the environment (clean state).
// 2. Stores the current branch to restore it later.
// 3. Loads metadata from gitgroove/internal.
// 4. Resolves the target repositories.
// 5. For each target:
//   - Switches to the repo's default branch (e.g., main).
//   - Pushes to the remote.
//   - Sets upstream if necessary.
//
// 6. Restores the original branch.
func Push(rootAbsPath string, targets []string) error {
	// 1. Validate environment
	if err := validateSwitchEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Store current branch
	originalBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	log.Info().Msgf("Saving current branch: %s", originalBranch)

	// Ensure we return to the original branch
	defer func() {
		log.Info().Msgf("Restoring original branch: %s", originalBranch)
		if err := gitUtil.Checkout(rootAbsPath, originalBranch); err != nil {
			log.Error().Err(err).Msg("Failed to restore original branch")
		}
	}()

	// 3. Load metadata
	// We need to be on gitgroove/internal to read authoritative metadata
	if err := gitUtil.Checkout(rootAbsPath, "gitgroove/internal"); err != nil {
		return fmt.Errorf("failed to checkout gitgroove/internal: %w", err)
	}

	repos, err := loadExistingRepos(rootAbsPath, "HEAD")
	if err != nil {
		return fmt.Errorf("failed to load repos: %w", err)
	}

	// 4. Resolve targets
	targetRepos := make(map[string]model.Repo)
	pushAll := false
	for _, t := range targets {
		if t == "*" {
			pushAll = true
			break
		}
	}

	if pushAll {
		for name, r := range repos {
			targetRepos[name] = r
		}
	} else {
		for _, t := range targets {
			if r, ok := repos[t]; ok {
				targetRepos[t] = r
			} else {
				return fmt.Errorf("repo '%s' not found", t)
			}
		}
	}

	if len(targetRepos) == 0 {
		return fmt.Errorf("no repositories selected to push")
	}

	// 5. Push each repo
	for name := range targetRepos {
		log.Info().Msgf("Processing repo: %s", name)

		// Determine branch to push. For MVP, we use the default branch.
		// Future improvement: allow pushing specific branches or current repo branch.
		branchToPush := model.DefaultRepoBranch

		// Construct the full GitGrove branch ref
		// refs/heads/gitgroove/repos/<repoName>/branches/<branchName>
		fullBranchRef := RepoBranchRef(name, branchToPush)
		shortBranchName := strings.TrimPrefix(fullBranchRef, "refs/heads/")

		// Check if branch exists locally
		if !gitUtil.RefExists(rootAbsPath, fullBranchRef) {
			log.Warn().Msgf("Branch %s does not exist for repo %s, skipping", branchToPush, name)
			continue
		}

		// Switch to the branch
		if err := gitUtil.Checkout(rootAbsPath, shortBranchName); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", shortBranchName, err)
		}

		// Push
		// We assume 'origin' is the remote.
		// We push the local branch (shortBranchName) to the remote branch (branchToPush).
		// e.g. git push origin refs/heads/gitgroove/.../main:refs/heads/main
		// Wait, user wants to push to remote.
		// If we just run `git push`, it pushes based on config.
		// If we want to set upstream, we need to be specific.

		// The local branch name is long: gitgroove/repos/backend/branches/main
		// The remote branch name should be: main

		remote := "origin"
		remoteBranch := branchToPush

		log.Info().Msgf("Pushing %s to %s/%s", shortBranchName, remote, remoteBranch)

		// Try pushing
		// git push origin <local>:<remote>
		refSpec := fmt.Sprintf("%s:refs/heads/%s", shortBranchName, remoteBranch)

		// We use RunGit directly to handle output and potential upstream issues
		// But wait, gitUtil.RunGit returns output and error.
		// We want to detect "upstream not set" or just force set upstream?
		// The user requirement: "if branch isnt up yet then sets upstream and pushes"

		// Let's try pushing with -u (set-upstream) always? No, that might be noisy or fail if already set?
		// Actually, `git push -u origin <refspec>` works fine even if already set.

		_, err := gitUtil.RunGit(rootAbsPath, "push", "-u", remote, refSpec)
		if err != nil {
			return fmt.Errorf("failed to push %s: %w", name, err)
		}

		log.Info().Msgf("Successfully pushed %s", name)
	}

	return nil
}
