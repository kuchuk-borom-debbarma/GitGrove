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

	// 3. Extract repo name from branch
	targetRepoName, err := ParseRepoFromBranch(currentBranch)
	if err != nil {
		return err
	}

	// 4. Load metadata from gitgroove/system (WITHOUT checkout)
	systemRef := "refs/heads/gitgroove/system"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, systemRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", systemRef, err)
	}

	repos, err := loadExistingRepos(rootAbsPath, oldTip)
	if err != nil {
		return fmt.Errorf("failed to load repo metadata: %w", err)
	}

	targetRepo, ok := repos[targetRepoName]
	if !ok {
		return fmt.Errorf("current branch belongs to unknown repo '%s'", targetRepoName)
	}

	targetRepoAbsPath := filepath.Join(rootAbsPath, targetRepo.Path)

	// 5. Validate .gitgroverepo identity
	markerPath := filepath.Join(targetRepoAbsPath, ".gitgroverepo")
	if !fileUtil.Exists(markerPath) {
		return fmt.Errorf("repo marker not found at %s (is this a valid GitGrove repo?)", markerPath)
	}

	markerContent, err := fileUtil.ReadTextFile(markerPath)
	if err != nil {
		return fmt.Errorf("failed to read repo marker: %w", err)
	}

	markerIdentity := strings.TrimSpace(markerContent)
	if markerIdentity != targetRepoName {
		return fmt.Errorf("repo identity mismatch: branch expects '%s', but marker says '%s'", targetRepoName, markerIdentity)
	}

	// 6. Validate staged changes belong to this repo
	// git diff --cached --name-only
	stagedFilesStr, err := gitUtil.RunGit(rootAbsPath, "diff", "--cached", "--name-only")
	if err != nil {
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	if stagedFilesStr == "" {
		return fmt.Errorf("nothing to commit (staged changes empty)")
	}

	stagedFiles := strings.Split(strings.TrimSpace(stagedFilesStr), "\n")
	for _, file := range stagedFiles {
		if file == "" {
			continue
		}

		absFile := filepath.Join(rootAbsPath, file)

		// Ensure file is inside repo root (targetRepoAbsPath)
		rel, err := filepath.Rel(targetRepoAbsPath, absFile)
		if err != nil || strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") {
			return fmt.Errorf("staged file '%s' is outside the current repository scope (%s)", file, targetRepoName)
		}

		// Ensure file is NOT inside nested repo
		if err := checkNestedRepo(targetRepoAbsPath, absFile); err != nil {
			return fmt.Errorf("staged file '%s' violates nested repo boundary: %w", file, err)
		}

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
