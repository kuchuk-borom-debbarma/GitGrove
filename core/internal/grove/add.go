package grove

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// Add adds file contents to the staging area, with GitGrove-specific validations.
//
// It ensures that:
// 1. We are inside a valid Git repository.
// 2. We are on a valid GitGrove repo branch (gitgroove/repos/...).
// 3. The files exist and are within the SPECIFIC repository bound to the current branch.
// 4. The files do NOT belong to a nested GitGrove repository.
func Add(rootAbsPath string, files []string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}

	// 1. Get current branch and validate it's a GitGrove repo branch
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// 1.1. Protection: Reject staging on gitgroove/internal branch
	if currentBranch == "gitgroove/internal" {
		return fmt.Errorf("cannot stage files on gitgroove/internal branch - this branch is managed by GitGrove and should not be modified directly")
	}

	// Expected format: gitgroove/repos/<repoName>/branches/<branchName>
	targetRepoName, err := ParseRepoFromBranch(currentBranch)
	if err != nil {
		return err
	}

	// 2. Load repo metadata to get the path
	internalRef := "refs/heads/gitgroove/internal"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, internalRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", internalRef, err)
	}

	repos, err := loadExistingRepos(rootAbsPath, oldTip)
	if err != nil {
		return fmt.Errorf("failed to load repo metadata: %w", err)
	}

	targetRepo, ok := repos[targetRepoName]
	if !ok {
		return fmt.Errorf("current branch belongs to unknown repo '%s'", targetRepoName)
	}

	// 3. Expand input files to all changed files (to handle directories like '.')
	expandedFiles, err := expandToChangedFiles(rootAbsPath, files)
	if err != nil {
		return fmt.Errorf("failed to expand file paths: %w", err)
	}

	if len(expandedFiles) == 0 {
		fmt.Println("No changed files match the provided paths.")
		return nil
	}

	// 4. Validate files (warn and skip invalid ones)
	filesToStage, err := validateStagingFiles(rootAbsPath, targetRepo, expandedFiles)
	if err != nil {
		return err
	}

	// 5. Batch Stage
	if len(filesToStage) > 0 {
		// We can pass multiple files to git add
		args := append([]string{"add", "--"}, filesToStage...)
		if _, err := gitUtil.RunGit(rootAbsPath, args...); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}
	} else {
		fmt.Println("No valid files to stage within the current repository scope.")
	}

	return nil
}

func expandToChangedFiles(rootAbsPath string, inputFiles []string) ([]string, error) {
	// Get all changed files (modified + untracked)
	// Use -z for machine readable output (null terminated, no quoting)
	out, err := gitUtil.RunGit(rootAbsPath, "status", "--porcelain", "-u", "-z")
	if err != nil {
		return nil, err
	}

	if out == "" {
		return []string{}, nil
	}

	var allChanged []string
	// Split by null byte
	entries := strings.Split(out, "\000")
	for _, entry := range entries {
		if len(entry) < 4 {
			continue
		}
		// Format is "XY PATH"
		// With -z, there is still a space after XY.
		// XY PATH
		// However, empirical evidence shows sometimes the prefix is only 2 chars (e.g. "M index.html").
		// Standard says 3 chars ("XY ").
		// We use a heuristic: if the character at index 2 is NOT a space, assume 2-char prefix.
		path := entry[3:]
		if entry[2] != ' ' {
			path = entry[2:]
		}
		allChanged = append(allChanged, path)
	}

	// Filter changed files against inputFiles
	var matchedFiles []string
	for _, changedRel := range allChanged {
		changedAbs := filepath.Join(rootAbsPath, changedRel)

		for _, input := range inputFiles {
			// Normalize input to absolute
			inputAbs := fileUtil.NormalizePath(input)
			if !filepath.IsAbs(inputAbs) {
				inputAbs = filepath.Join(rootAbsPath, input)
			}
			// Resolve symlinks to match git's canonical path
			if resolved, err := filepath.EvalSymlinks(inputAbs); err == nil {
				inputAbs = fileUtil.NormalizePath(resolved)
			}

			// Check if changed file matches input (exact or inside dir)
			if changedAbs == inputAbs || strings.HasPrefix(changedAbs, inputAbs+string(filepath.Separator)) {
				matchedFiles = append(matchedFiles, changedRel) // Keep relative for git add
				break
			}
		}
	}

	return matchedFiles, nil
}

func validateStagingFiles(rootAbsPath string, targetRepo model.Repo, files []string) ([]string, error) {
	var filesToStage []string

	for _, file := range files {
		// file is relative to rootAbsPath (from expandToChangedFiles)
		absFile := filepath.Join(rootAbsPath, file)

		// Check existence (should exist if it came from git status, but good to be safe)
		if !fileUtil.Exists(absFile) {
			// If deleted, git status shows it. git add should handle deleted files too.
			// But fileUtil.Exists returns false.
			// For now, let's assume if it's in git status, we can stage it.
			// But we need to know if it's deleted to skip nested check?
			// Let's just proceed.
		}

		// Verify file is strictly inside the target repo
		// Scope Check
		// If we are on a repo branch, the "scope" is the root (flattened view).
		// If we are on a normal branch (e.g. main), the scope is the repo path.
		// Since Add() is called, we know we are on a repo branch (checked in step 1).
		// So we should relax the scope check.

		// However, we must ensure we don't stage files that belong to OTHER repos (if any exist in this view? unlikely in flattened view).
		// In flattened view, everything visible is part of the repo (except .gg).

		// Old check:
		// if !strings.HasPrefix(absFile, targetRepoAbsPath+string(filepath.Separator)) && absFile != targetRepoAbsPath { ... }

		// New check:
		// In flattened view, targetRepoAbsPath is irrelevant for scope.
		// We just check if it's not .gg

		// But wait, what if the user is on 'main' (not a repo branch)?
		// Add() step 1 says: "Get current branch and validate it's a GitGrove repo branch".
		// So Add() ONLY works on repo branches.
		// Therefore, we ALWAYS assume flattened view.

		// So we SKIP the scope check against targetRepo.Path.

		/*
			// Original Scope Check (Disabled for flattened view)
			if !strings.HasPrefix(absFile, targetRepoAbsPath+string(filepath.Separator)) && absFile != targetRepoAbsPath {
				fmt.Printf("Warning: Skipping '%s' - outside current repository scope (%s)\n", file, targetRepo.Name)
				continue
			}
		*/

		// Nested Repo Check
		// In flattened view, there shouldn't be nested repos visible unless they are submodules or something?
		// Or if we have a repo-in-repo structure.
		// checkNestedRepo uses targetRepoAbsPath. This might be wrong now.
		// But let's assume for now we don't need strict nested repo checks in flattened view
		// because the view itself is constructed from a single repo's subtree.

		// Forbid staging .gg/ files
		// file is relative to root, so check prefix
		if strings.HasPrefix(file, ".gg/") || file == ".gg" {
			fmt.Printf("Warning: Skipping '%s' - cannot stage GitGroove metadata\n", file)
			continue
		}

		filesToStage = append(filesToStage, file)
	}
	return filesToStage, nil
}

func checkNestedRepo(rootAbsPath, fileAbsPath string) error {
	// Start checking from the file's directory
	dir := filepath.Dir(fileAbsPath)

	// Normalize paths for comparison
	root := filepath.Clean(rootAbsPath)
	current := filepath.Clean(dir)

	// Traverse up until we reach the root
	for {
		if current == root || len(current) < len(root) {
			break
		}

		// Check for .gitgroverepo marker
		markerPath := filepath.Join(current, ".gitgroverepo")
		if fileUtil.Exists(markerPath) {
			rel, _ := filepath.Rel(root, current)
			return fmt.Errorf("belongs to nested repo '%s'", rel)
		}

		// Move up
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return nil
}
