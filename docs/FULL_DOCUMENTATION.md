# GitGrove Full Documentation

Complete guide to using GitGrove for hierarchical monorepo management.

## Table of Contents

1. [Introduction](#introduction)
2. [Core Concepts](#core-concepts)
3. [Installation](#installation)
4. [Command Reference](#command-reference)
5. [Workflows](#workflows)
6. [Advanced Topics](#advanced-topics)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

## Introduction

GitGrove enables hierarchical organization of logical repositories within a single Git repository. It provides:

- **Isolation**: Work on specific repositories without seeing unrelated files
- **Hierarchy**: Define parent-child relationships between repositories
- **Validation**: Automatic checks to prevent accidental cross-repository changes
- **Navigation**: Move between repositories using intuitive commands

### When to Use GitGrove

**Good use cases:**
- Monorepos with multiple services/libraries
- Projects with shared dependencies
- Teams wanting logical separation without repository overhead

**Not recommended for:**
- Simple single-purpose repositories
- Repositories with frequent large binary file changes

## Core Concepts

### Repositories

A **repository** in GitGrove is a logical subdivision of your project. Each repository:
- Has a unique name (e.g., `backend`, `frontend`)
- Maps to a physical directory path
- Can have a parent repository
- Has its own branch namespace

### Hierarchy

Repositories can form parent-child relationships:

```
shared (libs/shared)
├── backend (services/backend)
│   └── auth-service (services/backend/auth)
└── frontend (apps/frontend)
```

**Rules:**
- Each repository can have at most one parent
- No cycles allowed
- Hierarchy is optional (flat is fine)

### Flattened View

When you switch to a repository, GitGrove creates a **flattened view**:

**Physical structure:**
```
project/
├── services/
│   └── backend/
│       ├── main.go
│       └── utils/
└── libs/
    └── shared/
```

**View when switched to `backend`:**
```
project/
├── main.go
├── utils/
└── .gitgroverepo  (marker file)
```

The repository's contents appear at the root level, isolating you from other parts of the project.

### Branches

GitGrove uses a special branch naming convention:

- **Internal branch**: `gitgroove/internal` (stores metadata)
- **Repo branches**: `gitgroove/repos/<repo>/branches/<branch>`

Example: `gitgroove/repos/backend/branches/feature-auth`

### Metadata

GitGrove stores metadata in `.gg/` directory on the `gitgroove/internal` branch:

```
.gg/
└── repos/
    ├── backend/
    │   ├── path          (contains: "services/backend")
    │   ├── parent        (contains: "shared")
    │   └── children/
    │       └── auth-service
    └── frontend/
        ├── path
        └── parent
```

Additionally, each registered repository gets a `.gitgroverepo` marker file containing its name.

## Installation

### From Source

```bash
git clone https://github.com/kuchuk-borom-debbarma/GitGrove
cd GitGrove
go build -o gg ./cmd/gg

# Add to PATH (Linux/macOS)
sudo mv gg /usr/local/bin/

# Or Windows
# Move gg.exe to a directory in your PATH
```

### Prerequisites

- Go 1.25.4 or later
- Git 2.0 or later
- Existing Git repository (or ability to create one)

## Command Reference

### `gg init`

Initializes GitGrove in the current Git repository.

**Prerequisites:**
- Must be inside a Git repository
- Working tree must be clean (no uncommitted changes)
- `.gg/` directory must not exist
- `gitgroove/internal` branch must not exist

**Usage:**
```bash
gg init
```

**What it does:**
1. Creates `.gg/repos/` directory structure
2. Creates `.gg/repos/.gitkeep` file
3. Creates and checks out `gitgroove/internal` branch
4. Commits initial metadata structure

**After initialization:**
- You'll be on the `gitgroove/internal` branch
- Switch back to your working branch: `git checkout main`

**Errors:**
- "not a valid Git repository" - Not inside a Git repo
- "working tree is not clean" - Uncommitted changes exist
- "GitGroove already initialized" - `.gg/` or internal branch exists

---

### `gg register`

Registers one or more repositories in GitGrove.

**Prerequisites:**
- GitGrove must be initialized (`gg init`)
- Working tree must be clean
- Must be on a user branch (not `gitgroove/internal`)
- Target directories must exist

**Usage:**
```bash
# Single repository
gg register backend=./services/backend

# Multiple repositories
gg register backend=./services/backend frontend=./apps/frontend

# Root directory as repository
gg register root=.
```

**Rules:**
- Repository names must match `^[a-zA-Z0-9._-]+$`
- Names must be unique
- Paths must be unique (no duplicates within batch or with existing repos)
- Paths must be directories that exist
- Paths must not contain `.git/` subdirectories (except at project root)
- Paths must not escape the project root

**What it does:**
1. Validates all inputs
2. Creates `.gg/repos/<name>/path` files with paths
3. Creates `.gitgroverepo` marker files in each directory
4. Creates orphan branches: `gitgroove/repos/<name>/branches/main`
5. Commits metadata to `gitgroove/internal`
6. Stages marker files in current branch (if on `gitgroove/internal`)

**After registration:**
- Marker files (`.gitgroverepo`) are created but may be untracked
- Commit them to your working branch if desired

**Errors:**
- "invalid repo name" - Name contains invalid characters
- "repo name already registered" - Duplicate name
- "path already registered" - Duplicate path
- "path does not exist" - Directory not found
- "path escapes project root" - Path uses `..` to escape
- "nested git repos not allowed" - Directory contains `.git/`

---

### `gg link`

Creates parent-child relationships between repositories.

**Prerequisites:**
- GitGrove must be initialized
- Working tree must be clean
- Both child and parent repositories must be registered

**Usage:**
```bash
# Make 'shared' the parent of 'backend'
gg link backend=shared

# Multiple relationships
gg link backend=shared frontend=shared api=backend
```

**Rules:**
- Both repositories must exist
- A repository can have only one parent
- No cycles allowed (e.g., A→B→A is forbidden)
- No self-references (repo cannot be its own parent)
- Parent directory must exist on disk

**What it does:**
1. Validates relationships (existence, no cycles)
2. Creates `.gg/repos/<child>/parent` files
3. Creates `.gg/repos/<parent>/children/<child>` files
4. Commits to `gitgroove/internal`

**Cycle detection:**
GitGrove builds the full graph (existing + new edges) and detects cycles using depth-first search.

**Errors:**
- "repo not registered" - Unknown repository name
- "cannot be its own parent" - Self-reference detected
- "already has a parent" - Trying to assign second parent
- "cycle detected" - Would create A→B→...→A cycle
- "dangling repo" - Child directory doesn't exist

---

### `gg switch`

Switches the working tree to a repository's branch.

**Prerequisites:**
- Working tree must be clean
- Repository must be registered
- Target branch must exist

**Usage:**
```bash
# Switch to repository's main branch
gg switch backend

# Switch to specific branch
gg switch backend feature-auth
```

**What it does:**
1. Checks out `gitgroove/internal` to load fresh metadata
2. Validates repository exists
3. Constructs target branch: `gitgroove/repos/<repo>/branches/<branch>`
4. Checks out the target branch
5. You now see the flattened view of that repository

**Default branch:**
If no branch specified, defaults to `main`.

**Errors:**
- "working tree is not clean" - Uncommitted changes
- "repo not found in metadata" - Unknown repository
- "target branch does not exist" - Branch not created yet

---

### `gg add`

Stages files with GitGrove-specific validation.

**Prerequisites:**
- Must be on a GitGrove repo branch (not `main`, not `gitgroove/internal`)
- Files must exist or be tracked

**Usage:**
```bash
# Stage specific files
gg add file1.go file2.go

# Stage all changes
gg add .

# Stage directory
gg add src/
```

**Validation:**
- Expands inputs to actual changed files (handles `.` correctly)
- Validates each file is within current repository scope
- Prevents staging `.gg/` metadata files
- Prevents staging files from nested repositories (via `.gitgroverepo` markers)

**What it does:**
1. Gets current branch and extracts repository name
2. Loads repository metadata
3. Expands file patterns to changed files (modified + untracked)
4. Filters files against restrictions
5. Stages valid files using `git add --`

**Warnings:**
Invalid files are skipped with warnings, valid files are still staged.

**Errors:**
- "not a git repository" - Not in Git repo
- "not a valid GitGrove repo branch" - On wrong branch
- "failed to load repo metadata" - Internal branch issue

---

### `gg commit`

Commits staged changes with validation.

**Prerequisites:**
- Must be on a GitGrove repo branch
- `.gitgroverepo` marker must match branch's repository
- Staged changes must exist

**Usage:**
```bash
gg commit -m "Add new feature"
```

**Validation:**
- Verifies current branch is a repo branch
- Loads repository metadata
- Checks `.gitgroverepo` marker matches expectations
- Validates staged files belong to current repository
- Prevents committing `.gg/` metadata

**What it does:**
1. Validates environment and staged files
2. Delegates to `git commit -m "<message>"`

**Errors:**
- "not a valid GitGrove repo branch" - On wrong branch
- "repo marker not found" - `.gitgroverepo` missing
- "repo marker mismatch" - Marker doesn't match branch
- "cannot commit GitGroove metadata" - Staged `.gg/` files

---

### `gg branch`

Creates a new branch for a specific repository.

**Prerequisites:**
- GitGrove must be initialized
- Working tree must be clean
- Repository must be registered

**Usage:**
```bash
# Create a branch for backend repository
gg branch backend feature-auth
```

**What it does:**
1. Checks out `gitgroove/internal` to load metadata
2. Validates repository exists
3. Gets the repository's subtree from current HEAD
4. Creates a commit with that subtree as root tree
5. Creates branch ref: `gitgroove/repos/<repo>/branches/<branch>`

**Branch naming:**
Full ref: `refs/heads/gitgroove/repos/backend/branches/feature-auth`

**Errors:**
- "working tree is not clean" - Uncommitted changes
- "repo not found" - Unknown repository
- "failed to get subtree hash" - Repository path not in HEAD

---

### `gg checkout`

Checks out a specific branch of a repository.

**Prerequisites:**
- Working tree must be clean
- Repository must be registered
- Branch must exist

**Usage:**
```bash
gg checkout backend feature-auth

# Preserve empty directories
gg checkout backend feature-auth --keep-empty-dirs

# Keep nested repositories visible (flat view)
gg checkout backend feature-auth --flat
```

**What it does:**
1. Checks out `gitgroove/internal` to load metadata
2. Validates repository and branch exist
3. Checks out the target branch
4. You see the flattened view of that repository
5. Recursively removes empty directories (unless `--keep-empty-dirs` is used)
6. Hides directories belonging to nested registered repositories (unless `--flat` is used)

**Equivalent to:**
```bash
gg switch backend feature-auth
```

**Errors:**
Same as `gg switch`

---

### `gg move`

Moves a repository to a new path within the project.

**Prerequisites:**
- Working tree must be clean
- Repository must be registered
- New path must not exist
- New path must not be used by another repository

**Usage:**
```bash
# Move backend from services/backend to api/backend
gg move backend api/backend
```

**What it does:**
1. Validates repository exists and new path is available
2. Moves the directory physically on disk
3. Updates `.gg/repos/<repo>/path` metadata
4. Commits to `gitgroove/internal`

**Identity preserved:**
- Repository name stays the same
- Repository branches stay the same
- Hierarchy relationships stay the same
- Only the physical path changes

**Errors:**
- "working tree is not clean" - Uncommitted changes
- "repo not found" - Unknown repository
- "path already used" - Another repo uses that path
- "destination path already exists" - Directory exists
- "invalid destination path" - Empty or invalid path

**Rollback:**
If metadata update fails, GitGrove attempts to move the directory back.

---

### `gg up`

Navigates to the parent repository in the hierarchy.

**Prerequisites:**
- Must be inside a registered repository (`.gitgroverepo` exists)
- Current repository must have a parent

**Usage:**
```bash
gg up
```

**What it does:**
1. Reads `.gitgroverepo` marker to identify current repository
2. Loads repository metadata
3. Finds parent repository
4. Switches to parent's `main` branch

**Errors:**
- "not inside a registered repository" - No `.gitgroverepo` marker
- "has no parent" - Repository is a root
- "failed to load repo info" - Metadata issue

---

### `gg down`

Navigates to a child repository in the hierarchy.

**Prerequisites:**
- Must be inside a registered repository
- Target child must exist
- Target child must be a direct child (not grandchild)

**Usage:**
```bash
gg down backend
```

**What it does:**
1. Reads `.gitgroverepo` marker to identify current repository
2. Loads repository metadata
3. Validates target is a direct child
4. Switches to child's `main` branch

**Errors:**
- "not inside a registered repository" - No marker
- "child repo not found" - Unknown child name
- "is not a child of" - Not a direct child

---

### `gg cd`

Changes repository context using familiar `cd` syntax.

**Usage:**
```bash
# Navigate to parent
gg cd ..

# Navigate to child
gg cd backend
```

**What it does:**
- `gg cd ..` → calls `gg up` (or switches to System Root if at root repo)
- `gg cd ~` → switches to System Root
- `gg cd <name>` → calls `gg down <name>`

This is syntactic sugar for up/down navigation.

**System Root View:**
When you navigate to the System Root (`~` or `..` from a root repo), you will see:
- A clean view with directory stubs for all **root** repositories.
- Nested repositories are hidden in this view.
- You can `cd` into any root repository from here.

---

### `gg ls`

Lists the direct children of the current repository.

**Prerequisites:**
- Must be inside a registered repository

**Usage:**
```bash
gg ls
```

**Output:**
```
backend
frontend
shared
```

**What it does:**
1. Reads `.gitgroverepo` marker
2. Loads repository metadata
3. Finds all repositories where `parent == current_repo`
4. Returns sorted list

**Errors:**
- "not inside a registered repository" - No marker

---

### `gg info`

Displays comprehensive project status.

**Usage:**
```bash
gg info
```

**Output:**
```
GitGrove Info
=============

Root:   /home/user/project
Branch: gitgroove/repos/backend/branches/main
State:  Clean
System: a1b2c3d4...

Registered Repositories:
------------------------
shared (libs/shared)
├── backend (services/backend)
│   └── auth-service (services/backend/auth)
└── frontend (apps/frontend)
```

**What it shows:**
- Current root directory
- Current branch
- Working tree state (Clean/Dirty)
- Internal branch commit hash
- Hierarchical tree of all repositories
- Current repository marked with `*`
- Missing repositories marked with `[MISSING]`

**No prerequisites** - works from any branch/state.

---

## Workflows

### Setting Up a New Project

```bash
# 1. Create Git repository
git init my-project
cd my-project

# 2. Create initial structure
mkdir -p services/{backend,frontend} libs/shared
echo "# My Project" > README.md
git add .
git commit -m "Initial commit"

# 3. Initialize GitGrove
gg init
git checkout main

# 4. Register repositories
gg register backend=services/backend
gg register frontend=services/frontend
gg register shared=libs/shared

# 5. Create hierarchy
gg link backend=shared
gg link frontend=shared

# 6. Commit markers (optional)
git add .
git commit -m "Setup GitGrove structure"
```

### Feature Development Workflow

```bash
# 1. Switch to the repository
gg switch backend

# 2. Create feature branch
gg branch backend feature-auth
gg checkout backend feature-auth

# 3. Make changes
echo "package auth" > auth.go

# 4. Stage and commit
gg add .
gg commit -m "Add authentication module"

# 5. Switch to test in parent context
gg up  # Move to 'shared' repository
```

### Working Across Multiple Repositories

```bash
# 1. Start in shared library
gg switch shared main

# 2. Make library changes
gg add .
gg commit -m "Update shared utility"

# 3. Test in dependent service
gg down backend
gg add .
gg commit -m "Integrate updated shared library"

# 4. Check overall status
gg info
```

### Reorganizing Repository Structure

```bash
# Move repository to new location
gg move backend api/core/backend

# Physical directory is moved
# All branches and hierarchy preserved
# Update code references if needed
```

## Advanced Topics

### Understanding Branch Structure

**Internal branch:**
```
gitgroove/internal
└── commits: metadata changes only
    └── files: .gg/repos/*/path, parent, children
```

**Repository branches:**
```
gitgroove/repos/backend/branches/main
gitgroove/repos/backend/branches/feature-auth
└── commits: flattened repository content
    └── tree: repository files at root level
```

### Marker Files

The `.gitgroverepo` file serves as:
1. **Identity marker** - Contains repository name
2. **Scope boundary** - Prevents staging nested repo files
3. **Navigation aid** - Enables `up`/`down`/`ls` commands

**Best practice:** Commit marker files to your working branches:
- Enables detection of repository boundaries
- Prevents accidental cross-repository changes
- Makes GitGrove structure visible in regular Git

### Metadata Consistency

GitGrove maintains consistency through:

1. **Atomic updates** - Metadata commits are CAS-based
2. **Pre-validation** - All inputs checked before changes
3. **Internal branch isolation** - Metadata never mixed with user data
4. **Fresh metadata loading** - Always reads from `gitgroove/internal`

### Working with Standard Git

You can use standard Git commands, but be careful:

**Safe operations:**
```bash
git status
git log
git diff
git show
git branch -a  # View all branches
```

**Risky operations:**
```bash
git checkout main          # Breaks flattened view
git add .gg/              # Manual metadata editing (dangerous)
git branch -d gitgroove/*  # Deletes GitGrove branches
git reset --hard          # Can lose GitGrove state
```

**Recommendation:** Use `gg` commands for all structural operations.

### Integration with CI/CD

GitGrove branches are regular Git branches, so CI/CD integration works normally:

```yaml
# .github/workflows/backend.yml
name: Backend Tests
on:
  push:
    branches:
      - 'gitgroove/repos/backend/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - run: go test ./...
```

## Best Practices

### Repository Organization

**Do:**
- Use semantic names: `backend`, `frontend`, `shared`
- Group by function: `services/`, `libs/`, `apps/`
- Keep hierarchy shallow (2-3 levels max)
- Register frequently-changed code as separate repos

**Don't:**
- Create too many tiny repositories (overhead)
- Create deep hierarchies (navigation complexity)
- Use overlapping paths (validation prevents this)

### Branching Strategy

**Recommended pattern:**
```
gitgroove/repos/<repo>/branches/main           # Stable
gitgroove/repos/<repo>/branches/develop        # Integration
gitgroove/repos/<repo>/branches/feature-*      # Features
gitgroove/repos/<repo>/branches/hotfix-*       # Fixes
```

Match your normal Git workflow, just in repository-specific namespaces.

### Committing Marker Files

**Option 1: Track markers (recommended)**
```bash
git add .gitgroverepo
git commit -m "Add repository markers"
```

**Benefits:**
- GitGrove structure visible in regular Git
- Easier onboarding for new team members
- Prevents accidental cross-repository commits

**Option 2: Ignore markers**
```bash
echo ".gitgroverepo" >> .gitignore
```

**Benefits:**
- Cleaner commit history
- GitGrove is "invisible" to non-users

### Team Collaboration

**Coordination:**
1. Only one person should run `register`/`link`/`move` at a time
2. Pull `gitgroove/internal` branch before structural changes
3. Communicate repository structure changes to team

**Onboarding:**
```bash
# New team member setup
git clone <repo>
gg switch backend  # Automatically works!
```

### Migration from Existing Monorepo

```bash
# 1. Initialize in existing repo
gg init
git checkout main

# 2. Identify logical boundaries
# Map out: where are the natural dividing lines?

# 3. Register incrementally
gg register shared=libs/shared
# Test: gg switch shared

gg register backend=services/backend
gg link backend=shared
# Test: gg switch backend

# 4. Gradually migrate workflows
# Use gg commands for new work
# Keep git for existing workflows initially
```

## Troubleshooting

### "Not a git repository"

**Cause:** Not inside a Git repository.

**Solution:**
```bash
git init
git add .
git commit -m "Initial commit"
gg init
```

---

### "Working tree is not clean"

**Cause:** Uncommitted changes exist.

**Check status:**
```bash
git status
```

**Solutions:**
```bash
# Option 1: Commit changes
git add .
git commit -m "WIP"

# Option 2: Stash changes
git stash
gg switch backend
git stash pop

# Option 3: Discard changes (careful!)
git reset --hard
```

---

### "Not a valid GitGrove repo branch"

**Cause:** Trying to use `gg add` or `gg commit` on wrong branch.

**Check current branch:**
```bash
git branch --show-current
```

**Solution:**
```bash
gg switch <repo>
```

---

### "Repo marker mismatch"

**Cause:** `.gitgroverepo` file doesn't match current branch.

**Scenarios:**
1. Manually switched branches with `git checkout`
2. In wrong directory
3. Marker file corrupted/edited

**Solution:**
```bash
# Option 1: Switch to correct repo
gg switch <correct-repo>

# Option 2: Fix marker (if you're sure)
echo "<repo-name>" > .gitgroverepo
```

---

### "Cycle detected"

**Cause:** Trying to create circular dependency.

**Example:**
```bash
gg link A=B
gg link B=C
gg link C=A  # ❌ Creates cycle: A→B→C→A
```

**Solution:** Redesign hierarchy to be acyclic (tree structure).

---

### "Path already registered"

**Cause:** Duplicate path in registration.

**Check existing repos:**
```bash
gg info
```

**Solutions:**
```bash
# Option 1: Use different path
gg register newname=services/backend-v2

# Option 2: Move existing repo
gg move oldname services/legacy/backend
gg register newname=services/backend
```

---

### Lost in Branch Hierarchy

**Symptoms:**
- Don't know current repository
- Don't know available children
- Don't remember hierarchy

**Solutions:**
```bash
# Where am I?
gg info

# What's below me?
gg ls

# Go to known location
gg switch <repo>
```

---

### Metadata Corruption

**Symptoms:**
- Commands fail with "failed to load repo metadata"
- `gitgroove/internal` branch issues

**Recovery:**
```bash
# Check internal branch
git checkout gitgroove/internal
git log --oneline -10
git diff HEAD~1

# If needed, reset to last good state
git reset --hard <good-commit>

# Verify structure
ls -la .gg/repos/

# Return to working branch
git checkout main
```

---

### Accidentally Deleted `.gitgroverepo`

**Cause:** Manually deleted or lost during merge.

**Recovery:**
```bash
# Recreate from metadata
gg switch <repo>  # Recreates marker automatically

# Or manually
echo "<repo-name>" > .gitgroverepo
```

---

## Glossary

- **Repository**: Logical subdivision of your project
- **Hierarchy**: Parent-child relationships between repositories
- **Flattened View**: Repository contents displayed at root level
- **System Branch**: Special branch storing GitGrove metadata (`gitgroove/internal`)
- **Repo Branch**: Branch specific to a repository (`gitgroove/repos/<n>/branches/<b>`)
- **Marker File**: `.gitgroverepo` file identifying repository boundaries
- **Metadata**: GitGroove configuration stored in `.gg/` directory
- **Clean State**: Working tree with no uncommitted changes
- **Scope Validation**: Checking files belong to current repository


