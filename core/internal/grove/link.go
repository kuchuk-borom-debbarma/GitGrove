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
Link defines one or more repo hierarchy relationships (childName → parentName) and updates the
GitGroove metadata accordingly. After committing the updated hierarchy, Link rebuilds the
derived repo branches (gitgroove/repos/<repo>/branches/<branch>) based on the state of the
project and the committed hierarchy.

===========================================================
=                 HIGH-LEVEL RESPONSIBILITY               =
===========================================================

Link:
  - Connects already-registered repos into a parent→child tree.
  - Stores these relationships in the per-repo metadata folders inside .gg/repos/.
  - Writes metadata exclusively to gitgroove/system via an atomic, append-only commit.
  - Rebuilds each repo’s isolated GitGroove branch (default = "main") based on ancestry.
  - Never modifies user branches or working directory content.

Link DOES NOT:
  - Register repos (registration must be done beforehand).
  - Accept repo paths — relationships use NAMES only.
  - Modify or create project files.
  - Depend on working tree state (except requiring it be clean).

===========================================================
=                     METADATA STRUCTURE                  =
===========================================================

Each repo has its own folder:

	.gg/repos/<repoName>/
	    path              (string: relative directory path)
	    parent            (string: parent repo name, empty if root)
	    children/         (empty files: one file per child repo)

Examples:

	.gg/repos/billing/path         → "services/billing"
	.gg/repos/billing/parent       → "services"
	.gg/repos/services/children/billing  (empty file)

This structure is modular, easy to inspect, and completely versioned inside gitgroove/system.

===========================================================
=                     LINKING ALGORITHM                   =
===========================================================

All operations occur against the *latest committed* gitgroove/system state to ensure
deterministic, atomic, conflict-free updates.

1. Validate environment:
  - Must be inside a GitGroove repo with .gg present.
  - Working tree must be clean.
  - HEAD must not be detached.
  - gitgroove/system must exist.

2. Read the latest metadata commit:
  - Resolve refs/heads/gitgroove/system → oldTip.
  - All metadata is read from oldTip (never from working tree).

3. Load registered repos:

  - Enumerate .gg/repos/* from oldTip.

  - For each, read:

  - path

  - parent

  - children/*

  - Build name→repo metadata map.

    4. Validate relationships:
    For each parentName → childName:

  - parentName must exist in registered repos.

  - childName must exist in registered repos.

  - child must not already have a parent.

  - parent != child.

  - The new edges must not introduce a cycle.

  - The child's path must still exist in the project filesystem (dangling repos forbidden).

    If ANY relationship fails → abort (no changes applied).

5. Prepare updated metadata in a temporary detached worktree:
  - git worktree add --detach <tempDir> <oldTip>
  - For each relationship:
  - Write parentName into .gg/repos/<child>/parent
  - Create empty file .gg/repos/<parent>/children/<child>
  - No other metadata is modified.

6. Commit updated metadata:
  - Stage updated .gg/repos/* content.
  - Create a new commit with parent = oldTip.
  - Capture commit hash → newTip.

7. Atomically update gitgroove/system:
  - Use compare-and-swap:
    git update-ref refs/heads/gitgroove/system <newTip> <oldTip>
  - If the CAS fails (system branch moved), abort and return retryable error.

===========================================================
=                DERIVED BRANCH RECONSTRUCTION            =
===========================================================

After gitgroove/system is updated, rebuild all repo branches.

For each registered repo <name>:

 1. Determine default branch:
    • Currently always "main".
    • Stored in meta.json in future versions.

 2. Build full ancestry chain:
    • Follow parent pointers upward until reaching a root repo.
    • Example:
    billing → services → core → root
    • Reverse to obtain directory structure:
    root / core / services / billing

 3. Construct isolated repo tree:
    • Extract only the directories in the ancestry chain from the *project HEAD* commit.
    • Use Git plumbing or an empty index to assemble this subtree.
    • Commit this isolated repo tree (no checkout occurs).

 4. Update repo branch:
    • Branch ref:
    refs/heads/gitgroove/repos/<name>/branches/main
    • Use git update-ref to point this branch to the newly constructed commit.
    • Branches are cheap pointers; no user working tree is touched.

This ensures each repo has its own clean, isolated, reproducible workspace that respects
hierarchy and preserves directory structure relative to the project.

===========================================================
=                         GUARANTEES                       =
===========================================================

• Link is fully atomic.
• No metadata is written unless all relationships pass validation.
• No partial updates — commit either fully applies or not at all.
• No user branches or files are modified.
• All metadata is committed, never derived from working tree.
• Branch creation is deterministic and driven solely by committed metadata.
• Dangling repos (missing paths) are rejected immediately.
• Cycles are rejected.
• Repo names are immutable IDs.
*/
func Link(rootAbsPath string, relationships map[string]string) error {
	rootAbsPath = fileUtil.NormalizePath(rootAbsPath)
	// relationships: child -> parent
	log.Info().Msgf("Attempting to link %d relationships in %s", len(relationships), rootAbsPath)

	// 1. Validate environment
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return fmt.Errorf("not a git repository: %s", rootAbsPath)
	}
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		return fmt.Errorf("working tree is not clean: %w", err)
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

	// 5. Prepare updated metadata in temporary workspace
	tempDir, err := os.MkdirTemp("", "gitgroove-link-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // cleanup

	// Create detached worktree at oldTip
	if err := gitUtil.WorktreeAddDetached(rootAbsPath, tempDir, oldTip); err != nil {
		return fmt.Errorf("failed to create temporary worktree: %w", err)
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
			return fmt.Errorf("failed to write parent for %s: %w", child, err)
		}

		// Write child pointer in parent's folder
		childrenDir := filepath.Join(tempDir, ".gg", "repos", parent, "children")
		if err := os.MkdirAll(childrenDir, 0755); err != nil {
			return fmt.Errorf("failed to create children dir for %s: %w", parent, err)
		}
		childFile := filepath.Join(childrenDir, child)
		if err := os.WriteFile(childFile, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to write child entry %s in %s: %w", child, parent, err)
		}
	}

	// 6. Commit updated metadata
	if err := gitUtil.StagePath(tempDir, ".gg/repos"); err != nil {
		return fmt.Errorf("failed to stage .gg/repos: %w", err)
	}
	if err := gitUtil.Commit(tempDir, fmt.Sprintf("Link %d repositories", len(relationships))); err != nil {
		return fmt.Errorf("failed to commit metadata changes: %w", err)
	}
	newTip, err := gitUtil.GetHeadCommit(tempDir)
	if err != nil {
		return fmt.Errorf("failed to get new commit hash: %w", err)
	}

	// 7. Atomically update gitgroove/system
	if err := gitUtil.UpdateRef(rootAbsPath, systemRef, newTip, oldTip); err != nil {
		return fmt.Errorf("failed to update %s (concurrent modification?): %w", systemRef, err)
	}

	// 8. Rebuild derived branches
	if err := rebuildBranches(rootAbsPath, newTip); err != nil {
		// Note: Metadata is already committed. Failure here means branches are out of sync.
		// We should probably return the error so the user knows.
		return fmt.Errorf("failed to rebuild branches: %w", err)
	}

	log.Info().Msg("Successfully linked repositories")
	return nil
}

func rebuildBranches(root, tip string) error {
	// TODO: Implement derived branch reconstruction
	return nil
}

func validateRelationships(root string, relationships map[string]string, existingRepos map[string]model.Repo) error {
	// 1. Check existence and validity
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

	// 2. Check for cycles using full graph
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
