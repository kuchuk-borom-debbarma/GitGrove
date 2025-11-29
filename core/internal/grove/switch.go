package grove

import (
	"fmt"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

/*
Switch moves the user's working tree to the GitGroove branch associated with the
specified repo and (optionally) branch name.

===========================================================
=                 HIGH-LEVEL RESPONSIBILITY               =
===========================================================

Switch:
  - First ensures the working tree is clean.
  - Then checks out the latest committed gitgroove/system branch to load
    authoritative metadata (.gg/repos).
  - Resolves the full ancestry chain of the requested repo.
  - Constructs the GitGroove repo branch ref:
    refs/heads/gitgroove/repos/<root/.../child>/branches/<branch>
  - Checks out the resolved branch.

Switch DOES NOT:
  - Modify metadata.
  - Create or rebuild derived branches.
  - Infer missing branches.
  - Touch project files except through standard git checkout.

===========================================================
=               MANDATORY PRE-STEP (CRITICAL)             =
===========================================================

Switch MUST always begin by checking out the latest gitgroove/system commit.

Reason:
  - All metadata lives ONLY on gitgroove/system.
  - Other working branches may not contain .gg/ at all.
  - Derived branch structures depend on committed hierarchy, not working state.
  - Guarantees deterministic behavior and prevents stale metadata usage.

Flow:
 1. Ensure working tree is clean.
 2. `git checkout gitgroove/system`
 3. Reload metadata from .gg/repos
 4. Continue Switch logic.

===========================================================
=                   BEHAVIORAL CONTRACT                    =
===========================================================

Given: Switch(repoName, branchName)

1. Preconditions:
  - Working tree must be clean.
  - Repo must exist in committed metadata.

2. Checkout system branch:
  - `git checkout gitgroove/system`
  - Ensures .gg/ is in correct, authoritative state.

3. Determine branch name:
  - Use "main" if branchName is empty.

4. Resolve ancestry chain:

  - Follow repo.parent → parent → ... → root.

  - Build path:
    <root>/<...>/<child>

    5. Construct fully-qualified GitGroove branch ref:
    refs/heads/gitgroove/repos/<path>/branches/<branchName>

6. Validate branch exists.

7. Checkout the branch:
  - Restore working tree to that isolated commit graph.

===========================================================
=                         GUARANTEES                       =
===========================================================

• Switch always uses fresh metadata (never stale .gg).
• No metadata is modified.
• No derived branches are created or rebuilt.
• Switch either fully succeeds or leaves the user on gitgroove/system.
• Repo ancestry resolution is always based on committed system state.
*/
func Switch(rootAbsPath, repoName, branch string) error {
	// 1. Validate environment
	if err := validateSwitchEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Checkout gitgroove/system to load authoritative metadata
	// This is CRITICAL: we must be on the system branch to read the correct .gg/ state.
	// We use "force" checkout? No, we already validated clean state.
	// Standard checkout is fine.
	log.Info().Msg("Checking out gitgroove/system to load metadata")
	if err := gitUtil.Checkout(rootAbsPath, "gitgroove/system"); err != nil {
		return fmt.Errorf("failed to checkout gitgroove/system: %w", err)
	}

	// 3. Load existing repos from the system branch (which is now HEAD)
	// We can use "HEAD" since we just checked it out.
	repos, err := loadExistingRepos(rootAbsPath, "HEAD")
	if err != nil {
		return fmt.Errorf("failed to load repos from system branch: %w", err)
	}

	// 4. Resolve ancestry chain
	targetRepo, ok := repos[repoName]
	if !ok {
		return fmt.Errorf("repo '%s' not found in metadata", repoName)
	}

	// Build ancestry: child -> parent -> ... -> root
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

	// Reverse to get root -> ... -> child
	var pathSegments []string
	for i := len(ancestry) - 1; i >= 0; i-- {
		pathSegments = append(pathSegments, ancestry[i].Name)
	}

	// 5. Construct fully-qualified GitGroove branch ref
	// Default to "main" if branch is empty
	targetBranchName := branch
	if targetBranchName == "" {
		targetBranchName = model.DefaultRepoBranch
	}

	// refs/heads/gitgroove/repos/<path>/branches/<branchName>
	fullBranchRef := fmt.Sprintf("refs/heads/gitgroove/repos/%s/branches/%s", strings.Join(pathSegments, "/"), targetBranchName)

	// 6. Validate branch exists
	if !gitUtil.RefExists(rootAbsPath, fullBranchRef) {
		return fmt.Errorf("target branch '%s' does not exist", fullBranchRef)
	}

	// 7. Checkout the branch
	// We must use the short name to ensure we attach to the branch (avoid detached HEAD).
	shortBranchName := strings.TrimPrefix(fullBranchRef, "refs/heads/")
	log.Info().Msgf("Switching to %s", shortBranchName)
	if err := gitUtil.Checkout(rootAbsPath, shortBranchName); err != nil {
		return fmt.Errorf("failed to checkout target branch: %w", err)
	}

	log.Info().Msgf("Successfully switched to repo '%s' on branch '%s'", repoName, targetBranchName)
	return nil
}

func validateSwitchEnvironment(rootAbsPath string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}
	return nil
}
