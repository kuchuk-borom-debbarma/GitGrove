package grove

/*
Link defines one or more repo hierarchy relationships (parentName → childName) and updates the
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
    • parentName must exist in registered repos.
    • childName must exist in registered repos.
    • child must not already have a parent.
    • parent != child.
    • The new edges must not introduce a cycle.
    • The child's path must still exist in the project filesystem (dangling repos forbidden).

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
func Link(relationships map[string]string) {

}
