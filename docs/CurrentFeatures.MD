# GitGrove: Current Features

This document outlines the currently implemented features of GitGrove (GG) as of the latest build.

## 1. Initialization (The Trunk)

Establishes the current branch as the source of truth (The Trunk) for the monorepo.

*   **Command**: `gg init` (via TUI or CLI)
*   **Functionality**:
    *   **Validation**: Checks if the current directory is a valid Git repository and ensures GitGrove is not already initialized.
    *   **Configuration**: Creates a `.gg` directory and a `gg.json` metadata file to store repository paths and relationships.
    *   **Hook Installation**: Installs a Git `pre-commit` hook to enforce atomic commits.
    *   **Commit**: Automatically commits the initial configuration (`.gg/gg.json`) to the current branch.

## 2. Registration (The Split)

Registers a folder as a logical "repository" within the monorepo, creating an isolated history for it.

*   **Command**: `gg register <name> <path>`
*   **Functionality**:
    *   **Validation**:
        *   Ensures GitGrove is initialized.
        *   Checks for name and path conflicts with existing repositories.
        *   **Nested Repository Prevention**: Strictly prevents registering a repository inside another registered repository (and vice versa).
    *   **Configuration Update**: Updates `gg.json` with the new repository's metadata.
    *   **Orphan Branch Creation**: Uses `git subtree split` to create a new orphan branch (e.g., `gg/<name>`) containing only the history of the specified folder.
    *   **Path Translation**: In the orphan branch, files are moved to the root directory, simulating a standalone repository.

## 3. Atomic Commit Enforcement (The Hook)

Ensures the integrity of the monorepo history by preventing "mixed" commits that could entangle independent projects.

*   **Mechanism**: Git Pre-commit Hook
*   **Logic**:
    *   **Context Detection**: Checks for the presence of `.gg/gg.json`. If missing (e.g., in an orphan branch), checks are skipped.
    *   **Rule 1: Isolation**: Prevents a single commit from modifying files in more than one registered repository simultaneously.
    *   **Rule 2: Separation**: Prevents a single commit from mixing files from a registered repository with root-level files (files outside any registered repo).
    *   **Error Message**: Blocks the commit and provides a clear error message explaining the violation (e.g., "Atomic Commit Violation").

## 4. Terminal User Interface (TUI)

A text-based interface for interacting with GitGrove.

*   **Current State**:
    *   **Initialization Flow**: Detects if the current directory is a GitGrove repository. If not, offers an interactive "Init GitGrove" option.
    *   **Dashboard**: Displays basic repository information ("Welcome to GitGrove!") if the repository is already initialized.
    *   **Navigation**: Basic keyboard navigation (Up/Down, Enter, Quit).

---

## 5. Context Aware Commit Message (Auto-Tagging) [Implemented]

Automatically tags commit messages to maintain a clean history without manual effort.

*   **Mechanism**: `prepare-commit-msg` Hook
*   **Logic**:
    *   **Trigger**: Fires when a commit is created.
    *   **Condition**: If all staged files belong to a *single* registered repository (e.g., `serviceA`).
    *   **Action**: Automatically prepends `[serviceA]` to the commit message.
    *   **Configuration**: Can be enabled/disabled via `repo_aware_context_message` flag in `gg.json`.

---

## Features In Progress / Planned

The following features are designed but not yet fully implemented:

## 6. Integration (`gg sync`) [Implemented]

Merges changes from Orphan branches (isolated history) back into the Trunk (monorepo history).

*   **Command**: `gg sync [repo-name]`
*   **Default Behavior**:
    *   **Squash**: Compresses all orphan commits into a single commit for a clean trunk history.
    *   **Review Mode (No Auto-Commit)**: Stages the changes but **does not commit** them, allowing user review.
*   **Flags**:
    *   `--no-squash`: Preserves the full commit history of the orphan branch.
    *   `--commit`: Automatically creating the merge commit (bypassing the review stage).
*   **Logic**:
    *   Auto-detects if run from an orphan branch and switches to trunk automatically.
    *   Uses `git subtree merge` (or `git merge -Xsubtree`) to correctly map files.
    *   Appends `[RepoName]` to the merge commit message.

---

## Features In Progress / Planned

The following features are designed but not yet fully implemented:

*   **Development Loop (`gg checkout`)**: Commands to easily switch between the Trunk and Orphan branches are not yet implemented.
*   **TUI Enhancements**:
    *   "Register Repo" option (Logic exists in `register-repo` module, needs TUI wiring).
    *   Dashboard for viewing registered repositories.

