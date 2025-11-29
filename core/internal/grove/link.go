package grove

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
	"github.com/rs/zerolog/log"
)

/*
Link defines hierarchy relationships (childName → parentName) and updates GitGroove metadata.
It operates atomically against the latest committed state of gitgroove/system.

High-Level Responsibility:
  - Connects registered repos into a parent→child tree.
  - Stores relationships in .gg/repos/<repo>/parent and .gg/repos/<parent>/children/<child>.
  - Rebuilds derived repo branches (gitgroove/repos/<repo>/branches/main) based on ancestry.

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

	// 2. Read latest gitgroove/system commit
	systemRef := "refs/heads/gitgroove/system"
	oldTip, err := gitUtil.ResolveRef(rootAbsPath, systemRef)
	if err != nil {
		return fmt.Errorf("failed to resolve %s (is GitGroove initialized?): %w", systemRef, err)
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

	// 7. Atomically update gitgroove/system
	if err := gitUtil.UpdateRef(rootAbsPath, systemRef, newTip, oldTip); err != nil {
		return fmt.Errorf("failed to update %s (concurrent modification?): %w", systemRef, err)
	}

	// If we are currently on the system branch, we must update the working tree to match the new commit.
	currentBranch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err == nil && currentBranch == "gitgroove/system" {
		log.Info().Msg("Updating working tree to match new system state")
		if err := gitUtil.ResetHard(rootAbsPath, "HEAD"); err != nil {
			return fmt.Errorf("failed to update working tree: %w", err)
		}
	}

	// 8. Build derived branches.
	if err := rebuildDerivedBranches(rootAbsPath, newTip); err != nil {
		return err
	}

	log.Info().Msg("Successfully linked repositories")
	return nil
}

func validateLinkEnvironment(rootAbsPath string) error {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
	}
	return nil
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
	}

	// 6. Commit updated metadata
	if err := gitUtil.StagePath(tempDir, ".gg/repos"); err != nil {
		return "", fmt.Errorf("failed to stage .gg/repos: %w", err)
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

func rebuildDerivedBranches(rootAbsPath, newTip string) error {
	// Reload repos from the new system state to get the complete picture including new links
	allRepos, err := loadExistingRepos(rootAbsPath, newTip)
	if err != nil {
		return fmt.Errorf("failed to reload repos from new tip: %w", err)
	}

	// For now, we use the project HEAD as the content for all branches.
	// TODO: In the future, construct an isolated repo tree containing only the ancestry paths.
	projectHead, err := gitUtil.GetHeadCommit(rootAbsPath)
	if err != nil {
		return fmt.Errorf("failed to get project HEAD: %w", err)
	}

	for _, repo := range allRepos {
		// Build ancestry: child -> parent -> ... -> root
		var ancestry []model.Repo
		current := repo
		for {
			ancestry = append(ancestry, current)
			if current.Parent == "" {
				break
			}
			parent, ok := allRepos[current.Parent]
			if !ok {
				// This implies a broken link in the committed metadata, which shouldn't happen if validation works.
				log.Warn().Msgf("Repo %s has missing parent %s, skipping branch build", current.Name, current.Parent)
				break
			}
			current = parent
		}

		// Construct branch name
		// gitgroove/repos/<root>/children/<child1>/children/<child2>/branches/main
		var pathSegments []string

		// Start with root
		rootRepo := ancestry[len(ancestry)-1]
		pathSegments = append(pathSegments, rootRepo.Name)

		// Append children recursively
		for i := len(ancestry) - 2; i >= 0; i-- {
			pathSegments = append(pathSegments, "children", ancestry[i].Name)
		}

		branchName := fmt.Sprintf("refs/heads/gitgroove/repos/%s/branches/%s", strings.Join(pathSegments, "/"), model.DefaultRepoBranch)

		// Update the branch ref
		if err := gitUtil.SetRef(rootAbsPath, branchName, projectHead); err != nil {
			return fmt.Errorf("failed to update branch ref %s: %w", branchName, err)
		}
	}
	return nil
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
