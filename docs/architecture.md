# GitGrove Architecture & Systems Guide

GitGrove is a tool for managing monorepos with the flexibility of polyrepos. It allows you to treat subdirectories as distinct logical repositories ("repos") with their own commit history and branching, while physically residing in a single Git monorepo.

This guide details the internal architecture, specifically the "Option B: Patch-Based Repo Branches" design.

## 1. Core Architecture: Patch-Based Repo Branches

The fundamental design choice of GitGrove is that **repo branches are semantic commit streams, not full project snapshots.**

### What does this mean?
In a standard Git repo, a commit represents the state of the *entire* project root.
In GitGrove:
- The **System Branch** (`gitgroove/system`) tracks metadata (what repos exist, where they are, how they link).
- **Repo Branches** (`gitgroove/repos/<name>/branches/<branch>`) track the history of *just that specific repository*.

Crucially, a repo branch commit **only contains the files belonging to that repo**. It does not contain files from other repos or the root, even if they exist in the working directory.

### Why this approach?
- **Performance**: We don't need to reconstruct massive trees for every commit.
- **Flexibility**: Moving a repo directory doesn't rewrite its entire history.
- **Simplicity**: Branch operations are lightweight.

---

## 2. The System Branch (`gitgroove/system`)

This is the "brain" of GitGrove. It is an orphan branch (unrelated to your code history) that stores metadata.

### Data Model
Metadata is stored as plain files in the `.gg` directory:

```
.gg/
  repos/
    backend/
      path      # Contains: "services/backend"
      parent    # Contains: "shared-lib" (optional)
    frontend/
      path      # Contains: "apps/frontend"
      parent    # Contains: "shared-lib"
```

### Operations
- **Atomic Updates**: All metadata changes use `git update-ref` with Compare-And-Swap (CAS) to ensure concurrency safety.
- **Temp Worktrees**: Operations like `Register` and `Move` create a temporary detached worktree to modify metadata, commit it, and then update the system branch pointer. This keeps the user's working tree clean.

---

## 3. Repository Identity & Markers

How does GitGrove know you are inside the "backend" repo?

### `.gitgroverepo` Marker
Every registered repository directory contains a `.gitgroverepo` file.
- **Content**: The name of the repo (e.g., `backend`).
- **Purpose**:
    1.  **Boundary Detection**: `gitgrove stage` uses this to prevent staging files from nested repos.
    2.  **Context Awareness**: `gitgrove commit` checks this to know which repo you are committing to.

> **Note**: These marker files are committed to the `gitgroove/system` branch (so they are tracked) and should also be committed to your user branches so they persist in the working tree.

---

## 4. Branch Naming Convention

GitGrove uses a canonical naming scheme for references:

- **Format**: `refs/heads/gitgroove/repos/<repoName>/branches/<branchName>`
- **Example**: `refs/heads/gitgroove/repos/backend/branches/feature/login`

This flat structure simplifies parsing and avoids complex ancestry logic in branch names.

---

## 5. Key Operations

### `Register`
Registers a new directory as a GitGrove repository.
1.  Validates the path is clean and unique.
2.  Creates an **orphan branch** for the repo (empty initial commit).
3.  Writes metadata (`path`) to `gitgroove/system`.
4.  Writes `.gitgroverepo` marker to the directory.

### `Link`
Defines a parent-child relationship (e.g., `frontend` depends on `shared`).
1.  Updates `.gg/repos/<child>/parent` in `gitgroove/system`.
2.  **Does NOT** change repo branches. Relationships are purely metadata-driven.

### `Switch`
Checks out a specific repo branch.
1.  Resolves the canonical ref: `gitgroove/repos/<repo>/branches/<branch>`.
2.  Performs a standard `git checkout`.
3.  The working tree updates to reflect that repo's state.

### `Stage`
A wrapper around `git add` with safety checks.
1.  Identifies the current repo from the branch name.
2.  Ensures you are only staging files *inside* that repo's directory.
3.  **Blocks** staging files from nested repos (detected via `.gitgroverepo`).

### `Commit`
A wrapper around `git commit`.
1.  Verifies the current branch is a valid GitGrove repo branch.
2.  Ensures staged files respect repo boundaries.
3.  Creates a standard Git commit on the repo branch.

### `Move`
Moves a repo to a new directory.
1.  Physically renames the directory on disk.
2.  Updates the `path` metadata in `gitgroove/system`.
3.  **Preserves History**: The repo branch is untouched. Since the branch stores content, and metadata stores location, moving is safe and easy.

---

## 6. Workflow Example

1.  **Init**: `gitgrove init` (creates `gitgroove/system`).
2.  **Register**: `gitgrove register --name backend --path ./backend`.
3.  **Work**:
    - `gitgrove switch backend main`
    - Edit `backend/main.go`
    - `gitgrove stage backend/main.go`
    - `gitgrove commit "Fix bug"`
4.  **Move**: `gitgrove move --repo backend --to ./services/backend`.

---

## 7. Developer Notes

- **Concurrency**: The system is designed to handle concurrent metadata updates via optimistic locking on the system branch ref.
- **Rollback**: Critical operations like `Move` attempt to rollback physical changes if metadata updates fail.
- **Extensibility**: The file-based metadata model (`.gg/repos/...`) is easily extensible for future features (e.g., storing descriptions, owners, or config).
