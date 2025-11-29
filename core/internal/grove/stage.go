package grove

import (
	"fmt"
	"path/filepath"
	"strings"

	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// Stage adds file contents to the staging area, with GitGrove-specific validations.
//
// It ensures that:
// 1. We are inside a valid Git repository.
// 2. We are on a valid GitGrove repo branch (gitgroove/repos/...).
// 3. The files exist and are within the SPECIFIC repository bound to the current branch.
// 4. The files do NOT belong to a nested GitGrove repository.
func Stage(rootAbsPath string, files []string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}

	// 1. Get current branch and validate it's a GitGrove repo branch
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Expected format: gitgroove/repos/<hierarchy>/branches/<branchName>
	// Hierarchy is like: root/children/child
	// We need to parse this robustly.
	prefix := "gitgroove/repos/"
	if !strings.HasPrefix(currentBranch, prefix) {
		return fmt.Errorf("not on a valid GitGrove repository branch (current: %s)", currentBranch)
	}

	// Remove prefix
	trimmed := strings.TrimPrefix(currentBranch, prefix)

	// Find the hierarchy part.
	// The hierarchy ends where "/branches/" begins.
	// Since repo names cannot contain "/", the hierarchy is a sequence of names separated by "/children/".
	// However, a repo name COULD be "branches" or "children" (if validNameRegex allows).
	// But the structure is fixed: name -> children -> name -> children -> ...
	// So we can walk the path segments.

	// Split by "/"
	segments := strings.Split(trimmed, "/")

	// We expect: [repo1, "children", repo2, "children", ..., "branches", branchName...]
	// We need to find the "branches" segment that terminates the hierarchy.
	// But "branches" could be a repo name.
	// Let's parse strictly based on the alternating pattern.

	var repoNames []string
	var branchNameIdx int = -1

	// State machine: 0=ExpectRepoName, 1=ExpectChildrenOrBranches
	state := 0
	for i, seg := range segments {
		if state == 0 {
			// Expect repo name
			repoNames = append(repoNames, seg)
			state = 1
		} else {
			// Expect "children" or "branches"
			if seg == "children" {
				state = 0 // Next is repo name
			} else if seg == "branches" {
				branchNameIdx = i + 1
				break
			} else {
				// Invalid structure
				return fmt.Errorf("invalid branch format at segment '%s': expected 'children' or 'branches'", seg)
			}
		}
	}

	if branchNameIdx == -1 || branchNameIdx >= len(segments) {
		return fmt.Errorf("invalid GitGrove branch format: missing branch name")
	}

	// The target repo is the last one in the hierarchy
	if len(repoNames) == 0 {
		return fmt.Errorf("invalid GitGrove branch format: no repo found")
	}
	targetRepoName := repoNames[len(repoNames)-1]

	// 2. Load repo metadata to get the path
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

	// 3. Validate files
	targetRepoAbsPath := filepath.Join(rootAbsPath, targetRepo.Path)
	var filesToStage []string

	for _, file := range files {
		// Normalize and resolve absolute path
		cleanFile := fileUtil.NormalizePath(file)
		absFile := cleanFile
		if !filepath.IsAbs(cleanFile) {
			absFile = filepath.Join(rootAbsPath, cleanFile)
		}

		// Check existence
		if !fileUtil.Exists(absFile) {
			return fmt.Errorf("pathspec '%s' did not match any files", file)
		}

		// Verify file is strictly inside the target repo
		rel, err := filepath.Rel(targetRepoAbsPath, absFile)
		if err != nil || strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") {
			return fmt.Errorf("path '%s' is outside the current repository scope (%s)", file, targetRepo.Name)
		}

		// Nested Repo Check
		if err := checkNestedRepo(targetRepoAbsPath, absFile); err != nil {
			return err
		}

		// Collect relative path for batch staging
		relToRoot, _ := filepath.Rel(rootAbsPath, absFile)

		// Forbid staging .gg/ files
		if strings.HasPrefix(relToRoot, ".gg/") || relToRoot == ".gg" {
			return fmt.Errorf("cannot stage GitGroove metadata: %s", relToRoot)
		}

		filesToStage = append(filesToStage, relToRoot)
	}

	// 4. Batch Stage
	if len(filesToStage) > 0 {
		// We can pass multiple files to git add
		// gitUtil.StagePath currently takes one file. We need to use runGit directly or update StagePath.
		// Let's use runGit directly here for efficiency.
		args := append([]string{"add", "-f", "--"}, filesToStage...)
		if _, err := gitUtil.RunGit(rootAbsPath, args...); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}
	}

	return nil
}

func checkNestedRepo(rootAbsPath, fileAbsPath string) error {
	// Start checking from the file's directory
	dir := filepath.Dir(fileAbsPath)

	// Normalize paths for comparison
	root := filepath.Clean(rootAbsPath)
	current := filepath.Clean(dir)

	// Traverse up until we reach the root
	for {
		if current == root {
			break
		}

		// Check for .gitgroverepo marker
		markerPath := filepath.Join(current, ".gitgroverepo")
		if fileUtil.Exists(markerPath) {
			rel, _ := filepath.Rel(root, current)
			return fmt.Errorf("path '%s' belongs to nested repo '%s'", filepath.Base(fileAbsPath), rel)
		}

		// Move up
		parent := filepath.Dir(current)
		if parent == current {
			// Should not happen if we are inside root, but safety break
			break
		}
		current = parent
	}

	return nil
}
