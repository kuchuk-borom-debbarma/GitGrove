# GitGrove Architecture

GitGrove (GG) is a tool for managing a modular monorepo structure. It provides a layer of history isolation and strict workflow enforcement over a standard Git repository.

## Core Concepts

### 1. The Trunk (`main` / root)
The central integration branch. It contains the entire codebase. 
- **Role**: Source of Truth.
- **Workflow**: Developers merge into the trunk, but rarely commit directly to it for registered components.

### 2. The Split (Orphan Branches)
For every registered component (e.g., `backend/serviceA`), GitGrove maintains a parallel "orphan" branch (e.g., `gg/main/serviceA`).
- **Role**: Isolated development environment.
- **Mechanism**: `git subtree split`.
- **View**: Files from `backend/serviceA/*` are projected to the root `./*`.

### 3. The Guard (Hooks)
Automated checks to enforce the "Atomic Commit" and "Context Isolation" principles.

### 4. The Merge (Re-Integration)
Mechanism to bring changes from an isolated orphan branch back into the trunk without losing context or directory structure.
- **Role**: Safe integration.
- **Workflow**: `Orphan Branch` -> `Prepare Merge Branch` (Merge Prep) -> `Pull Request` -> `Trunk`.

---

## Internal Modules (`src/internal`)

### `grove/initialize`
Handles the setup of a GitGrove workspace.
- **Entry**: `Initialize(path string, atomicCommit bool)`
- **Key Actions**:
  1. Validates the directory is a git repo.
  2. Creates `.gg/gg.json` (Configuration).
  3. Installs git hooks (`pre-commit`, `prepare-commit-msg`).
  4. Commits the config to the current branch.

### `grove/register-repo`
Manages the registration of sub-projects.
- **Entry**: `RegisterRepo(repos []model.GGRepo, ggRepoPath string)`
- **Key Actions**:
  1. Validates no path conflicts or nested repositories.
  2. Updates `gg.json` and commits it to the trunk.
  3. Executes `git subtree split` to create the initial orphan branch (`gg/<trunk>/<repo>`).

### `grove/prepare-merge`
Automates the creation of a merge-ready branch from an orphan branch.
- **Entry**: `PrepareMerge(ggRepoPath string, repoNameArg string)`
- **Key Actions**:
  1. Detects context (Orphan vs Trunk).
  2. Switches to Trunk (`main`).
  3. Creates `gg/merge-prep/<repoName>/<timestamp>` branch.
  4. Merges orphan branch using `git merge -s subtree --allow-unrelated-histories`.

### `grove/hooks`
The enforcement layer.

#### `PreCommit`
- **Trigger**: `git commit`
- **Logic**:
  1. Checks `.gg/gg.json`.
  2. Analyzes staged files.
  3. **Blocking Rule**: Rejects commits that touch multiple registered repositories, or mix a registered repository with root files.

#### `PrepareCommitMsg`
- **Trigger**: `git commit` (before message editing)
- **Logic**:
  1. If `RepoAwareContextMessage` is enabled in config:
  2. Checks staged files.
  3. If all files belong to a single registered repo (e.g., `serviceA`), prepends `[serviceA]` to the commit message.

### `tui`
The Terminal User Interface (BubbleTea).
- **Entry**: `InitialModel()`
- **States**:
  - `Startup`: Checks if CWD is a GG repo. If initialized (Trunk or Orphan), enters `Idle`.
  - `Init`: Prompts for path and Atomic Commit preference.
  - `ActionSelection` (Idle): Main menu.
    - **Trunk**: Shows Registered Repos. Options: Register Repo, Prepare Merge.
    - **Orphan**: Shows Orphan Context. Options: Prepare Merge.
  - `OpenRepo`: Allows opening an existing GG repo from a non-root CWD.
  - `RegisterRepo`: Interactive flow to add repositories.

## Data Structure: `gg.json`

Located at `.gg/gg.json`.

```json
{
  "version": "1.0",
  "repo_aware_context_message": true,
  "repositories": {
    "serviceA": {
      "Name": "serviceA",
      "Path": "backend/services/serviceA"
    }
  }
}
```
