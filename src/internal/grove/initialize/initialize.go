package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// Description returns a description of the initialization process.
func Description() string {
	return "Initialize: Establishes the current branch as the Trunk.\n" +
		"- Creates .gg/gg.json registry\n" +
		"- Installs pre-commit and prepare-commit-msg hooks\n" +
		"- Commits the configuration to the current branch"
}

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
//     as the root of the GitGrove system.
func Initialize(path string, atomicCommit bool) error {
	path = filepath.Clean(path)
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
	if err := groveUtil.CreateGroveConfig(path, atomicCommit); err != nil {
		return err
	}

	// Install hooks
	if err := installHooks(path); err != nil {
		return err
	}

	//Commit this configuration to the current branch
	// Use CommitNoVerify to prevent hook failure during initialization if the global binary is mismatched
	if err := gitUtil.CommitNoVerify(path, []string{".gg/gg.json"}, "Initialize GitGrove"); err != nil {
		return err
	}

	return nil
}

func installHooks(path string) error {
	// Pre-commit
	preCommitHookPath := filepath.Join(path, ".git", "hooks", "pre-commit")
	preCommitContent := `#!/bin/sh
# GitGrove Pre-commit Hook
# This hook ensures atomic commits across the GitGrove monorepo.

# Check if git-grove is in PATH
if ! command -v git-grove >/dev/null 2>&1; then
    echo "Warning: git-grove not found in PATH. Skipping atomic commit enforcement."
    exit 0
fi

# Execute git-grove hook
# We capture output to check for errors, but also allow stdout to pass through if needed
OUTPUT=$(git-grove hook pre-commit 2>&1)
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
    echo "GitGrove Pre-commit Hook Failed:"
    echo "$OUTPUT"
    echo ""
    echo "Tip: Ensure you have the latest version of git-grove installed and in your PATH."
    exit $EXIT_CODE
fi
`
	if err := os.WriteFile(preCommitHookPath, []byte(preCommitContent), 0755); err != nil {
		return fmt.Errorf("failed to create pre-commit hook: %w", err)
	}

	// Prepare-commit-msg
	prepareMsgHookPath := filepath.Join(path, ".git", "hooks", "prepare-commit-msg")
	prepareMsgContent := `#!/bin/sh
# GitGrove Prepare-commit-msg Hook

if command -v git-grove >/dev/null 2>&1; then
    git-grove hook prepare-commit-msg "$1" "$2" "$3"
fi
`
	if err := os.WriteFile(prepareMsgHookPath, []byte(prepareMsgContent), 0755); err != nil {
		return fmt.Errorf("failed to create prepare-commit-msg hook: %w", err)
	}

	return nil
}
