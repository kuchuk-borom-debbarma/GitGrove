package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

// Register records one or more repos (name → path) in the GitGroove metadata.
//
// It operates atomically against the latest committed state of the gitgroove/internal branch.
// Validated entries are appended to .gg/repos/<name>/path, a new commit is created,
// and the gitgroove/internal reference is updated.
//
// Guarantees:
//   - Atomic: either all repos are registered or none.
//   - Safe: validates environment (clean state, valid git repo) and inputs (unique names/paths).
//   - Non-destructive: does not modify user branches or working directory content (except .gitgroverepo marker).
func Register(rootAbsPath string, repos map[string]string) error {
	log.Info().Msgf("Attempting to register %d repos in %s", len(repos), rootAbsPath)

	// 1. Validate environment
	if err := validateRegisterEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Read latest gitgroove/internal commit
	internalRef := "refs/heads/gitgroove/internal"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, internalRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s (is GitGrove initialized?): %w", internalRef, err)
	}

	// 3. Load existing repo metadata
	existingRepos, err := loadExistingRepos(rootAbsPath, oldTip)
	if err != nil {
		return fmt.Errorf("failed to load existing repos: %w", err)
	}

	// 4. Validate incoming repos
	if err := validateNewRepos(rootAbsPath, repos, existingRepos); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 5. Prepare updated metadata in temporary workspace and create commit
	newTip, err := createRegisterCommit(rootAbsPath, oldTip, repos)
	if err != nil {
		return err
	}

	// 7. Atomically update gitgroove/internal
	if err := gitUtil.UpdateRef(rootAbsPath, internalRef, newTip, oldTip); err != nil {
		return fmt.Errorf("failed to update %s (concurrent modification?): %w", internalRef, err)
	}

	// If we are currently on the internal branch, we must update the working tree to match the new commit.
	// Otherwise, the working tree will appear "dirty" (missing the new files we just committed).
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err == nil && currentBranch == "gitgroove/internal" {
		log.Info().Msg("Updating working tree to match new internal state")
		if err := gitUtil.ResetHard(rootAbsPath, "HEAD"); err != nil {
			return fmt.Errorf("failed to update working tree: %w", err)
		}
	}

	// 8. Create orphan branches for each registered repo
	for name, path := range repos {
		// Branch: refs/heads/gitgroove/repos/<name>/branches/main
		// The orphan branch will include the .gitgroverepo marker from the internal branch commit
		branchRef := RepoBranchRef(name, model.DefaultRepoBranch)
		if !gitUtil.RefExists(rootAbsPath, branchRef) {
			log.Info().Msgf("Creating orphan branch %s", branchRef)

			// Create a tree containing the existing content of the repo folder (flattened)
			// PLUS the marker file.

			// 1. Get HEAD commit to find existing content
			headCommit, err := gitUtil.GetHeadCommit(rootAbsPath)
			var baseTreeHash string
			if err == nil {
				// 2. Get subtree hash for the repo path
				// If path is not in HEAD (e.g. new untracked folder), this might fail or return error.
				// We treat error as "empty tree".
				subtree, err := gitUtil.GetSubtreeHash(rootAbsPath, headCommit, path)
				if err == nil {
					baseTreeHash = subtree
				} else {
					log.Debug().Msgf("Subtree for %s not found in HEAD (new repo?): %v", path, err)
				}
			}

			// 3. Add .gitgroverepo marker to this tree
			treeHash, err := gitUtil.AddFileToTree(rootAbsPath, baseTreeHash, ".gitgroverepo", name)
			if err != nil {
				log.Warn().Msgf("Failed to create tree for orphan branch: %v", err)
				continue
			}

			commitHash, err := gitUtil.CommitTree(rootAbsPath, treeHash, "Initial repo structure")
			if err != nil {
				log.Warn().Msgf("Failed to create orphan branch %s: %v", branchRef, err)
			} else {
				if err := gitUtil.SetRef(rootAbsPath, branchRef, commitHash); err != nil {
					log.Warn().Msgf("Failed to set ref %s: %v", branchRef, err)
				}
			}
		}
	}

	log.Info().Msg("Successfully registered repositories")
	return nil
}

func validateRegisterEnvironment(rootAbsPath string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}
	return nil
}

// canonicalizePath normalizes and converts a path to be relative to rootAbsPath.
// This ensures consistent path handling throughout the registration process.
func canonicalizePath(rootAbsPath, path string) string {
	cleanPath := fileUtil.NormalizePath(path)
	if filepath.IsAbs(cleanPath) {
		rel, _ := filepath.Rel(rootAbsPath, cleanPath)
		cleanPath = fileUtil.NormalizePath(rel)
	}
	return cleanPath
}

func createRegisterCommit(rootAbsPath, oldTip string, repos map[string]string) (string, error) {
	// 5. Prepare updated metadata in temporary workspace
	tempDir, err := os.MkdirTemp("", "gitgroove-register-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // cleanup

	// Create detached worktree at oldTip
	if err := gitUtil.WorktreeAddDetached(rootAbsPath, tempDir, oldTip); err != nil {
		return "", fmt.Errorf("failed to create temporary worktree: %w", err)
	}
	defer gitUtil.WorktreeRemove(rootAbsPath, tempDir) // cleanup worktree

	// Write new repos to .gg/repos/<name>/path AND write marker files
	for name, path := range repos {
		repoDir := filepath.Join(tempDir, ".gg", "repos", name)
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create dir for repo %s: %w", name, err)
		}

		pathFile := filepath.Join(repoDir, "path")
		// Canonicalize path before writing
		cleanPath := canonicalizePath(rootAbsPath, path)

		if err := os.WriteFile(pathFile, []byte(cleanPath), 0644); err != nil {
			return "", fmt.Errorf("failed to write path for repo %s: %w", name, err)
		}

		// Write marker file in the temp worktree so it gets committed
		// This ensures that when we are on gitgroove/internal, the marker is tracked.
		// Note: repoDir is .gg/repos/<name>, but the marker should be in the actual repo path?
		// WAIT. The marker file goes into the ACTUAL repo path, not .gg/repos.
		// The `createRegisterCommit` function is working in a temp worktree of the ROOT repo.
		// The `repos` map contains paths relative to root.
		// So we need to write to `tempDir/<path>/.gitgroverepo`.

		// Let's re-read the plan and the code.
		// The code currently stages ".gg/repos".
		// I need to write the marker to `tempDir/<path>/.gitgroverepo` and ALSO stage it.
	}

	// Write marker files to the actual repo locations in the temp worktree
	for name, path := range repos {
		// path is relative to root
		// We need to write to tempDir/path/.gitgroverepo
		cleanPath := canonicalizePath(rootAbsPath, path)

		fullPath := filepath.Join(tempDir, cleanPath)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create repo dir in temp worktree %s: %w", fullPath, err)
		}

		markerPath := filepath.Join(fullPath, ".gitgroverepo")
		if err := os.WriteFile(markerPath, []byte(name), 0644); err != nil {
			return "", fmt.Errorf("failed to write marker in temp worktree %s: %w", markerPath, err)
		}

		// Create directory stub in the root of the internal branch (tempDir)
		// This makes the repo visible when "ls" is run in the internal branch.
		// We create <repoName>/.gitkeep
		stubDir := filepath.Join(tempDir, name)
		if err := os.MkdirAll(stubDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create stub dir %s: %w", stubDir, err)
		}
		stubKeep := filepath.Join(stubDir, ".gitkeep")
		if err := fileUtil.CreateEmptyFile(stubKeep); err != nil {
			return "", fmt.Errorf("failed to create stub .gitkeep %s: %w", stubKeep, err)
		}
	}

	// 6. Create new commit
	// Stage everything in .gg/repos
	if err := gitUtil.StagePath(tempDir, ".gg/repos"); err != nil {
		return "", fmt.Errorf("failed to stage .gg/repos: %w", err)
	}

	// Stage marker files and stub directories
	for name, path := range repos {
		// Canonicalize path
		cleanPath := canonicalizePath(rootAbsPath, path)

		markerRelPath := filepath.Join(cleanPath, ".gitgroverepo")
		if err := gitUtil.StagePath(tempDir, markerRelPath); err != nil {
			return "", fmt.Errorf("failed to stage marker file %s: %w", markerRelPath, err)
		}

		// Stage stub directory
		// We just need to stage the .gitkeep file inside it
		stubKeepRel := filepath.Join(name, ".gitkeep")
		if err := gitUtil.StagePath(tempDir, stubKeepRel); err != nil {
			return "", fmt.Errorf("failed to stage stub file %s: %w", stubKeepRel, err)
		}
	}

	if err := gitUtil.Commit(tempDir, fmt.Sprintf("Register %d new repositories", len(repos))); err != nil {
		return "", fmt.Errorf("failed to commit metadata changes: %w", err)
	}
	newTip, err := gitUtil.GetHeadCommit(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to get new commit hash: %w", err)
	}
	return newTip, nil
}

func loadExistingRepos(root, ref string) (map[string]model.Repo, error) {
	// List directories in .gg/repos
	// Note: .gg/repos might not exist if no repos are registered yet.
	// git ls-tree will fail if the path doesn't exist.
	// We should check if .gg/repos exists first or handle the error.
	// runGit returns error if path not found.

	entries, err := gitUtil.ListTree(root, ref, ".gg/repos")
	if err != nil {
		// If the directory doesn't exist, it's not an error, just empty.
		// git ls-tree returns error if path not found.
		// We need to distinguish "path not found" from other errors if possible.
		// For now, assuming any error from ListTree on a specific path means it doesn't exist or is empty
		// is a simplification. A better check would be explicitly checking existence.
		// But based on gitUtil implementation, let's assume it returns error if not found.
		return map[string]model.Repo{}, nil
	}

	repos := make(map[string]model.Repo)
	for _, name := range entries {
		// Ignore .gitkeep
		if name == ".gitkeep" {
			continue
		}

		// Each entry is a repo name (directory)
		// Read .gg/repos/<name>/path
		pathFile := fmt.Sprintf(".gg/repos/%s/path", name)
		content, err := gitUtil.ShowFile(root, ref, pathFile)
		if err != nil {
			// If path file is missing, it's a corruption or partial state.
			return nil, fmt.Errorf("failed to read path for repo %s: %w", name, err)
		}

		repoPath := strings.TrimSpace(content)

		// Read .gg/repos/<name>/parent
		parentFile := fmt.Sprintf(".gg/repos/%s/parent", name)
		parentContent, err := gitUtil.ShowFile(root, ref, parentFile)
		parent := ""
		if err == nil {
			parent = strings.TrimSpace(parentContent)
		} else {
			// Parent file might not exist for root repos or newly registered ones
			// We treat error as empty parent
		}

		repos[name] = model.Repo{
			Name:   name,
			Path:   repoPath,
			Parent: parent,
		}
	}

	return repos, nil
}

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func validateNewRepos(root string, newRepos map[string]string, existing map[string]model.Repo) error {
	// 1. Check for name collisions and validity
	for name := range newRepos {
		if err := validateRepoName(name, existing); err != nil {
			return err
		}
	}

	// 2. Check for path collisions and validity
	existingPaths := make(map[string]bool)
	for _, r := range existing {
		existingPaths[r.Path] = true
	}

	// Track paths in the current batch to detect duplicates within the batch
	seenPaths := make(map[string]string)

	for name, relPath := range newRepos {
		if err := validateRepoPath(root, name, relPath, existingPaths, seenPaths); err != nil {
			return err
		}
	}

	return nil
}

func validateRepoName(name string, existing map[string]model.Repo) error {
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("invalid repo name '%s': must match %s", name, validNameRegex.String())
	}
	if _, ok := existing[name]; ok {
		return fmt.Errorf("repo name '%s' already registered", name)
	}
	return nil
}

func validateRepoPath(root, name, relPath string, existingPaths map[string]bool, seenPaths map[string]string) error {
	// Path uniqueness
	cleanPath := fileUtil.NormalizePath(relPath)

	if existingPaths[cleanPath] {
		return fmt.Errorf("path '%s' already registered", relPath)
	}

	if otherName, ok := seenPaths[cleanPath]; ok {
		return fmt.Errorf("duplicate path '%s' used by '%s' and '%s'", relPath, otherName, name)
	}
	seenPaths[cleanPath] = name

	// Existence and containment
	absPath := filepath.Join(root, cleanPath)

	// Verify path is inside root
	rel, err := filepath.Rel(root, absPath)
	if err != nil || strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") {
		return fmt.Errorf("path '%s' escapes project root", relPath)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path '%s' does not exist", relPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", relPath)
	}

	// No nested .git
	// Exception: if absPath is the root itself, .git is expected.
	if absPath != root && fileUtil.Exists(filepath.Join(absPath, ".git")) {
		return fmt.Errorf("repo '%s' contains .git directory (nested git repos not allowed)", relPath)
	}

	return nil
}
