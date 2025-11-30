package grove

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
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

	// Expected format: gitgroove/repos/<repoName>/branches/<branchName>
	targetRepoName, err := ParseRepoFromBranch(currentBranch)
	if err != nil {
		return err
	}

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
	// git status --porcelain -u
	out, err := gitUtil.RunGit(rootAbsPath, "status", "--porcelain", "-u")
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(out) == "" {
		return []string{}, nil
	}

	var allChanged []string
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		// Format is "XY PATH"
		path := line[3:]
		// Handle quoted paths if any (git status quotes paths with spaces/special chars)
		if strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"") {
			path = path[1 : len(path)-1]
			// TODO: Unescape if needed, but for now assume simple quotes
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
	targetRepoAbsPath := filepath.Join(rootAbsPath, targetRepo.Path)
	var filesToStage []string

	for _, file := range files {
		// file is relative to rootAbsPath (from expandToChangedFiles)
		absFile := filepath.Join(rootAbsPath, file)

		// Check existence (should exist if it came from git status, but good to be safe)
		if !fileUtil.Exists(absFile) {
			// If deleted, git status shows it. git add should handle deleted files too.
			// But fileUtil.Exists returns false.
			// If we want to stage deletions, we should allow non-existent files if they are in git status.
			// For now, let's assume we proceed.
		}

		// Verify file is strictly inside the target repo
		rel, err := filepath.Rel(targetRepoAbsPath, absFile)
		if err != nil || strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") {
			fmt.Printf("Warning: Skipping '%s' - outside current repository scope (%s)\n", file, targetRepo.Name)
			continue
		}

		// Nested Repo Check
		if err := checkNestedRepo(targetRepoAbsPath, absFile); err != nil {
			fmt.Printf("Warning: Skipping '%s' - %v\n", file, err)
			continue
		}

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
