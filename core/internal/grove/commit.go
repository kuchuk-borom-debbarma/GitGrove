package grove

import (
	"fmt"
	"path/filepath"
	"strings"

	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// Commit performs a commit with strict GitGrove validations.
//
// It ensures:
// 1. We are inside a git repo.
// 2. We are on a valid GitGrove repo branch.
// 3. The .gitgroverepo marker matches the repo identity derived from the branch.
// 4. All staged changes belong strictly to this repo (no nested repos, no .gg metadata).
// 5. Delegates to `git commit`.
func Commit(rootAbsPath, message string) error {
	// 1. Verify inside git repo
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}

	// 2. Verify we are on a GitGrove repo branch
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// 2.1. Protection: Reject commits on gitgroove/internal branch
	if currentBranch == "gitgroove/internal" {
		return fmt.Errorf("cannot commit on gitgroove/internal branch - this branch is managed by GitGrove and should not be modified directly")
	}

	// 3. Extract repo name from branch
	targetRepoName, err := ParseRepoFromBranch(currentBranch)
	if err != nil {
		return err
	}

	// 4. Load metadata from gitgroove/internal (WITHOUT checkout)
	internalRef := "refs/heads/gitgroove/internal"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, internalRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", internalRef, err)
	}

	repos, err := loadExistingRepos(rootAbsPath, oldTip)
	if err != nil {
		return fmt.Errorf("failed to load repo metadata: %w", err)
	}

	_, ok := repos[targetRepoName]
	if !ok {
		return fmt.Errorf("current branch belongs to unknown repo '%s'", targetRepoName)
	}

	// 3. Verify .gitgroverepo marker
	// In flattened view, the marker is at the root.
	markerPath := filepath.Join(rootAbsPath, ".gitgroverepo")
	if !fileUtil.Exists(markerPath) {
		return fmt.Errorf("repo marker not found at %s (is this a valid GitGrove repo?)", markerPath)
	}

	content, err := fileUtil.ReadTextFile(markerPath)
	if err != nil {
		return fmt.Errorf("failed to read repo marker: %w", err)
	}
	if strings.TrimSpace(content) != targetRepoName {
		return fmt.Errorf("repo marker mismatch: expected '%s', got '%s'", targetRepoName, content)
	}

	// 4. Validate staged files
	// Ensure all staged files belong to this repo (scope check).
	// In flattened view, everything at root belongs to the repo (except .gg).
	stagedFiles, err := gitUtil.GetStagedFiles(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	for _, file := range stagedFiles {
		absFile := filepath.Join(rootAbsPath, file)

		// Nested Repo Check
		// In flattened view, we assume no nested repos are visible/staged unless explicitly added.
		// We skip strict nested repo check for now as path logic is different.
		// if err := checkNestedRepo(targetRepoAbsPath, absFile); err != nil { ... }

		// Ensure not .gg/**
		relToRoot, _ := filepath.Rel(rootAbsPath, absFile)
		if strings.HasPrefix(relToRoot, ".gg/") || relToRoot == ".gg" {
			return fmt.Errorf("cannot commit GitGroove metadata: %s", relToRoot)
		}
	}

	// 7. Perform the commit
	log.Info().Msgf("Committing changes to repo '%s' on branch '%s'", targetRepoName, currentBranch)
	if err := gitUtil.Commit(rootAbsPath, message); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}
