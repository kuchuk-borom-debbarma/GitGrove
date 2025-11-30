# GitGrove Technical Design & Internals

## Overview
GitGrove is a tool for managing monorepos by creating a **virtualized, flattened view** of nested repositories. It allows developers to work on a specific sub-project (like `backend/service-a`) as if it were the root of the repository, while maintaining a single unified Git history.

## Core Architecture

### 1. The Two Worlds
GitGrove separates **System Metadata** from **User Content**.

*   **System Branch (`gitgroove/system`)**:
    *   Stores the "truth" about the monorepo structure.
    *   Contains a hidden directory `.gg/` with metadata about registered repositories and their relationships.
    *   This branch is never checked out directly by the user for editing.
    *   **Git Plumbing**: Uses `git worktree add --detach` to safely manipulate this metadata in the background without touching the user's working directory.

*   **Repo Branches (`gitgroove/repos/<name>/branches/<branch>`)**:
    *   These are the branches the user actually works on.
    *   **Flattened View**: When you check out one of these branches, the content of the specific repository (e.g., `backend/service-a`) appears at the **root** of your working directory.
    *   **Isolation**: You only see the files relevant to that repository.

### 2. The Identity Marker (`.gitgroverepo`)
Every registered repository contains a marker file named `.gitgroverepo` at its root.
*   **Content**: The name of the repository (e.g., `service-a`).
*   **Purpose**:
    *   **Context Awareness**: Tells GitGrove "I am currently inside `service-a`".
    *   **Safety**: Prevents committing to the wrong repository. If you try to commit to `service-a` but the marker says `backend`, GitGrove blocks it.

---

## Command Internals

Here is a deep dive into what happens under the hood for each command.

### `init`
Initializes GitGrove in an existing Git repository.
*   **Action**: Creates the `gitgroove/system` branch.
*   **Git Commands**:
    *   `git hash-object -w --stdin`: Creates an empty tree object.
    *   `git commit-tree`: Creates an initial commit for the system branch.
    *   `git update-ref`: Creates `refs/heads/gitgroove/system` pointing to that commit.

### `register`
Registers a subdirectory as a GitGrove repository.
*   **Action**:
    1.  Updates metadata in `gitgroove/system`.
    2.  Creates an initial "orphan" branch for the repo.
*   **Git Commands**:
    *   `git worktree add --detach`: Creates a temp workspace to edit `.gg/repos`.
    *   `git add` & `git commit`: Saves the new repo metadata (path, name).
    *   `git update-ref`: Updates `gitgroove/system`.
    *   `git rev-parse HEAD:<path>`: Finds the tree hash of the existing subdirectory.
    *   `git commit-tree`: Creates a new commit where that subdirectory's tree is the **root** tree. This is the magic behind flattening.
    *   `git update-ref`: Creates `refs/heads/gitgroove/repos/<name>/branches/main`.

### `link`
Defines a parent-child relationship (purely metadata).
*   **Action**: Records "Child -> Parent" in `gitgroove/system`.
*   **Git Commands**:
    *   Similar to `register`, it uses a detached worktree to update `.gg/repos/<child>/parent` and `.gg/repos/<parent>/children/<child>`.

### `switch <repo> <branch>`
Switches the user's working directory to a specific repo branch.
*   **Action**: Checks out the flattened branch.
*   **Git Commands**:
    *   `git checkout gitgroove/repos/<repo>/branches/<branch>`: Standard git checkout.
    *   Since the branch was created with the subdirectory as its root, `git checkout` naturally replaces your working directory with just those files.

### `up` / `down <child>`
Navigates the hierarchy.
*   **Action**:
    1.  Reads `.gitgroverepo` to find current repo name.
    2.  Reads `gitgroove/system` to find parent/child name.
    3.  Calls `switch` to the target repo.

### `add` (Stage)
Stages files with safety checks.
*   **Action**: Filters files before passing them to Git.
*   **Git Commands**:
    *   `git status -z`: Lists changed files in a machine-readable format.
    *   **Logic**:
        *   Resolves symlinks to ensure paths match Git's canonical view.
        *   Checks if files are within the allowed scope (though in flattened view, everything at root is allowed).
        *   **Blocks** `.gg/` files (metadata protection).
    *   `git add <files>`: Finally stages the valid files.

### `commit`
Records a snapshot.
*   **Action**: Enforces identity and boundaries.
*   **Git Commands**:
    *   **Identity Check**: Reads `.gitgroverepo` from the working root. Verifies it matches the repo name implied by the current branch ref.
    *   `git diff --cached --name-only`: Lists staged files.
    *   **Scope Check**: Ensures no files belong to other known repositories (crucial in non-flattened views, less so in flattened).
    *   `git commit -m <msg>`: Performs the actual commit.

### `move`
Relocates a repository to a new path.
*   **Action**: Updates the path in `gitgroove/system`.
*   **Git Commands**:
    *   Updates `.gg/repos/<name>/path` in the system branch.
    *   **Note**: This does *not* move files in the user's history immediately. It only updates where GitGrove *looks* for them in future operations.

## Data Model (`.gg` folder)

The `gitgroove/system` branch contains a `.gg` folder with this structure:

```text
.gg/
└── repos/
    ├── backend/
    │   ├── path      (content: "backend")
    │   └── children/
    │       └── service-a (empty file)
    └── service-a/
        ├── path      (content: "backend/service-a")
        └── parent    (content: "backend")
```

This simple file-based database allows GitGrove to handle complex hierarchies using standard Git versioning, ensuring that your project structure is versioned right alongside your code.
