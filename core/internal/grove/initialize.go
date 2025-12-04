package grove

import (
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/grove"
)

// Initialize establishes the "Trunk" for the GitGrove monorepo.
//
// Concept: The Trunk
// GitGrove (GG) operates on the principle of a unified integration trunk ("The Trunk")
// while maintaining strict history isolation for sub-projects. Initialization is the
// first step where the current branch is designated as this source of truth.
//
// Workflow:
//  1. Initialize the GG environment.
//  2. Create a `gg.json` metadata file. This file acts as the registry for all
//     logical repositories and their paths within the monorepo.
//  3. Commit this configuration to the current branch, formally establishing it
//     as the root of the GitGrove system.
func Initialize(path string) error {
	//Validations
	//Validate if its a valid git repository
	if err := gitUtil.IsGitRepository(path); err != nil {
		return err
	}
	//Validate that there is no existing .gg/gg.json
	if err := groveUtil.IsGroveInitialized(path); err != nil {
		return err
	}

	//create .gg/gg.json file
	if err := groveUtil.CreateGroveConfig(path); err != nil {
		return err
	}
	//Commit this configuration to the current branch
	if err := gitUtil.Commit(path, []string{".gg/gg.json"}, "Initialize GitGrove"); err != nil {
		return err
	}

	return nil
}
