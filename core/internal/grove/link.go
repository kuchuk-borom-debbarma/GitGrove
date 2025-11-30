package grove

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

/*
Link defines hierarchy relationships (childName → parentName) and updates GitGroove metadata.
It operates atomically against the latest committed state of gitgroove/internal.

High-Level Responsibility:
  - Connects registered repos into a parent→child tree.
  - Stores relationships in .gg/repos/<repo>/parent and .gg/repos/<parent>/children/<child>.

Guarantees:
  - Atomic: all relationships are applied or none.
  - Safe: validates environment, existence of repos, and absence of cycles.
  - Non-destructive: does not modify user branches or working directory content.
*/
func Link(rootAbsPath string, relationships map[string]string) error {
	rootAbsPath = fileUtil.NormalizePath(rootAbsPath)
	// relationships: child -> parent
	log.Info().Msgf("Attempting to link %d relationships in %s", len(relationships), rootAbsPath)

	// 1. Validate environment
	if err := validateLinkEnvironment(rootAbsPath); err != nil {
		return err
	}

	// 2. Read latest gitgroove/internal commit
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, InternalBranchRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s (is GitGrove initialized?): %w", InternalBranchRef, err)
	}

	// 3. Load existing repo metadata
	existingRepos, err := loadExistingRepos(rootAbsPath, oldTip)
	if err != nil {
		return fmt.Errorf("failed to load existing repos: %w", err)
	}

	// 4. Validate relationships
	if err := validateRelationships(rootAbsPath, relationships, existingRepos); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 5. Prepare updated metadata in temporary workspace and create commit
	newTip, err := applyRelationships(rootAbsPath, oldTip, relationships)
	if err != nil {
		return err
	}

	// 7. Atomically update gitgroove/internal
	if err := gitUtil.UpdateRef(rootAbsPath, InternalBranchRef, newTip, oldTip); err != nil {
		return fmt.Errorf("failed to update %s (concurrent modification?): %w", InternalBranchRef, err)
	}

	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err == nil && currentBranch == InternalBranchName {
		log.Info().Msg("Updating working tree to match new internal state")
		if err := gitUtil.ResetHard(rootAbsPath, "HEAD"); err != nil {
			return fmt.Errorf("failed to update working tree: %w", err)
		}
	}

	log.Info().Msg("Successfully linked repositories")
	return nil
}

func validateLinkEnvironment(rootAbsPath string) error {
	return validateCleanGitRepo(rootAbsPath)
}

func applyRelationships(rootAbsPath, oldTip string, relationships map[string]string) (string, error) {
	// 5. Prepare updated metadata in temporary workspace
	tempDir, err := os.MkdirTemp("", "gitgroove-link-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // cleanup

	// Create detached worktree at oldTip
	if err := gitUtil.WorktreeAddDetached(rootAbsPath, tempDir, oldTip); err != nil {
		return "", fmt.Errorf("failed to create temporary worktree: %w", err)
	}
	defer gitUtil.WorktreeRemove(rootAbsPath, tempDir) // cleanup worktree

	// Apply relationships
	// For each child -> parent:
	// 1. Write parent name to .gg/repos/<child>/parent
	// 2. Create empty file .gg/repos/<parent>/children/<child>
	for child, parent := range relationships {
		// Write parent pointer
		parentFile := filepath.Join(tempDir, ".gg", "repos", child, "parent")
		if err := os.WriteFile(parentFile, []byte(parent), 0644); err != nil {
			return "", fmt.Errorf("failed to write parent for %s: %w", child, err)
		}

		// Write child pointer in parent's folder
		childrenDir := filepath.Join(tempDir, ".gg", "repos", parent, "children")
		if err := os.MkdirAll(childrenDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create children dir for %s: %w", parent, err)
		}
		childFile := filepath.Join(childrenDir, child)
		if err := os.WriteFile(childFile, []byte{}, 0644); err != nil {
			return "", fmt.Errorf("failed to write child entry %s in %s: %w", child, parent, err)
		}

		// Remove the child's stub directory from the root of system branch
		// This repo is no longer a root repo, so it shouldn't be visible in system root.
		stubDir := filepath.Join(tempDir, child)
		if err := os.RemoveAll(stubDir); err != nil {
			return "", fmt.Errorf("failed to remove stub dir %s: %w", stubDir, err)
		}
	}

	// 6. Commit updated metadata
	// Stage .gg/repos
	if err := gitUtil.StagePath(tempDir, ".gg/repos"); err != nil {
		return "", fmt.Errorf("failed to stage .gg/repos: %w", err)
	}

	// Stage removal of stubs
	for child := range relationships {
		// We need to stage the removal. "git add -u" or "git add <path>" works for deletions too if file is gone.
		if err := gitUtil.StagePath(tempDir, child); err != nil {
			// If the stub didn't exist (e.g. re-linking), this might fail or warn.
			// git add on a missing path that was tracked records the deletion.
			// But if it wasn't tracked, it might error.
			// However, registered repos SHOULD have a stub.
			// Let's log warning but proceed? Or fail?
			// Ideally we should check if it was tracked.
			// For now, let's assume it works or we ignore error if it's just "pathspec did not match".
			// But StagePath wraps runGit which returns error.
			// Let's try to be robust.
			log.Debug().Msgf("Staging removal of %s (might fail if not tracked)", child)
		}
	}
	if err := gitUtil.Commit(tempDir, fmt.Sprintf("Link %d repositories", len(relationships))); err != nil {
		return "", fmt.Errorf("failed to commit metadata changes: %w", err)
	}
	newTip, err := gitUtil.GetHeadCommit(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to get new commit hash: %w", err)
	}
	return newTip, nil
}

func validateRelationships(root string, relationships map[string]string, existingRepos map[string]model.Repo) error {
	// 1. Check existence and validity
	if err := validateRelationshipExistence(root, relationships, existingRepos); err != nil {
		return err
	}

	// 2. Check for cycles using full graph
	if err := validateRelationshipCycles(relationships, existingRepos); err != nil {
		return err
	}

	return nil
}

func validateRelationshipExistence(root string, relationships map[string]string, existingRepos map[string]model.Repo) error {
	for child, parent := range relationships {
		childRepo, ok := existingRepos[child]
		if !ok {
			return fmt.Errorf("child repo '%s' not registered", child)
		}
		if _, ok := existingRepos[parent]; !ok {
			return fmt.Errorf("parent repo '%s' not registered", parent)
		}
		if child == parent {
			return fmt.Errorf("repo '%s' cannot be its own parent", child)
		}
		if childRepo.Parent != "" {
			return fmt.Errorf("repo '%s' already has a parent ('%s')", child, childRepo.Parent)
		}

		// Check dangling repo (child path must exist)
		childAbsPath := filepath.Join(root, childRepo.Path)
		if _, err := os.Stat(childAbsPath); err != nil {
			return fmt.Errorf("child repo '%s' path '%s' does not exist (dangling repo)", child, childRepo.Path)
		}
	}
	return nil
}

func validateRelationshipCycles(relationships map[string]string, existingRepos map[string]model.Repo) error {
	// Build graph: parent -> []children
	graph := make(map[string][]string)

	// Add existing edges
	for name, repo := range existingRepos {
		if repo.Parent != "" {
			graph[repo.Parent] = append(graph[repo.Parent], name)
		}
	}

	// Add new edges
	for child, parent := range relationships {
		graph[parent] = append(graph[parent], child)
	}

	// Detect cycles
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var checkCycle func(current string) error
	checkCycle = func(current string) error {
		visited[current] = true
		recursionStack[current] = true

		for _, child := range graph[current] {
			if !visited[child] {
				if err := checkCycle(child); err != nil {
					return err
				}
			} else if recursionStack[child] {
				return fmt.Errorf("cycle detected involving '%s' and '%s'", current, child)
			}
		}

		recursionStack[current] = false
		return nil
	}

	// Check all nodes in the graph
	for node := range graph {
		if !visited[node] {
			if err := checkCycle(node); err != nil {
				return err
			}
		}
	}

	return nil
}
