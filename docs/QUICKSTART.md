# GitGrove Quick Start Guide

Get started with GitGrove in 5 minutes.

## Prerequisites

Before using GitGrove, ensure you have:

1. **Git installed** (version 2.0+)
2. **An initialized Git repository** with at least one commit
3. **A clean working tree** (no uncommitted changes)
4. **GitGrove CLI** installed and in your PATH

## Step 1: Initialize GitGrove

Navigate to your Git repository and initialize GitGrove:

```bash
cd /path/to/your/repo
gg init
```

**What happens:**
- Creates `.gg/` directory for metadata
- Creates `gitgroove/system` branch for tracking metadata
- Leaves you on the system branch

**Important:** After initialization, switch back to your main branch:

```bash
git checkout main  # or your default branch
```

## Step 2: Register Repositories

Register logical repositories within your project:

```bash
# Register individual repositories
gg register backend=./services/backend
gg register frontend=./services/frontend
gg register shared=./libs/shared
```

**Rules:**
- Repository names must be unique (alphanumeric, dots, dashes, underscores only)
- Paths must exist as directories
- Paths must be unique (no overlapping repositories)
- Paths cannot contain nested `.git` directories
- Working tree must be clean before registering

**What happens:**
- Creates metadata in `.gg/repos/<name>/path`
- Creates `.gitgroverepo` marker files in each directory
- Creates initial orphan branches for each repo: `gitgroove/repos/<name>/branches/main`
- Automatically commits marker files to `gitgroove/system`

**After registration:** Commit the marker files to your working branch if you want them tracked:

```bash
git add .
git commit -m "Add GitGrove repository markers"
```

## Step 3: Create Hierarchy (Optional)

Link repositories into parent-child relationships:

```bash
# Make 'shared' the parent of 'backend' and 'frontend'
gg link backend=shared
gg link frontend=shared
```

**Rules:**
- Both child and parent must be registered
- A repository can have only one parent
- No cycles allowed (e.g., A→B→C→A is forbidden)
- Repositories can be reparented later
- Working tree must be clean before linking

**What happens:**
- Updates metadata: `.gg/repos/backend/parent` → "shared"
- Creates child pointers: `.gg/repos/shared/children/backend`
- Commits changes to `gitgroove/system`

## Step 4: Switch to a Repository

Switch your working tree to a repository:

```bash
gg switch backend
```

**What you see:**
- **Flattened view**: Repository contents appear at the root level
- Only files from `backend` are visible (plus `.gitgroverepo` marker)
- You're on branch: `gitgroove/repos/backend/branches/main`

**Rules:**
- Working tree must be clean before switching
- Repository must exist and be registered
- GitGrove always checks out `gitgroove/system` first to load fresh metadata

## Step 5: Work with Files

Work normally with Git commands, enhanced by GitGrove:

```bash
# Make changes to files
echo "new feature" > feature.go

# Stage changes (GitGrove validates scope)
gg add .

# Commit (GitGrove validates you're on a repo branch)
gg commit -m "Add new feature"
```

**Validation rules for `gg add`:**
- Must be on a GitGrove repo branch (not `main`, not `gitgroove/system`)
- Files must be within the current repository's scope
- Cannot stage `.gg/` metadata files
- Cannot stage files from nested repositories

**Validation rules for `gg commit`:**
- Must be on a GitGrove repo branch
- `.gitgroverepo` marker must match the branch's repository
- Staged files must belong to current repository
- Cannot commit `.gg/` metadata

## Step 6: Navigate the Hierarchy

Move between parent and child repositories:

```bash
# List children of current repository
gg ls

# Move down to a child
gg down backend

# Move up to parent
gg up

# Or use cd-style navigation
gg cd backend    # Move to child
gg cd ..         # Move to parent
```

**Requirements:**
- `.gitgroverepo` marker must exist and be readable
- Working tree must be clean
- Target repository must exist in the hierarchy

## Step 7: Check Status

View your GitGrove project status:

```bash
gg info
```

**Output includes:**
- Current branch and repository
- Clean/dirty working tree status
- System branch commit hash
- Hierarchical tree of all repositories
- Physical path existence status

## Common Workflows

### Creating a New Feature Branch

```bash
# Switch to the repository
gg switch backend

# Create a new branch for this repo
gg branch feature-auth

# Checkout the new branch
gg checkout backend feature-auth

# Or use the shorthand
gg switch backend feature-auth
```

### Moving a Repository

```bash
# Physical path will change, but identity preserved
gg move backend services/api/backend
```

**Requirements:**
- Working tree must be clean
- New path must not exist
- New path must not conflict with other repositories

### Working with Standard Git

You can use regular Git commands, but be aware:

```bash
# These work normally
git status
git log
git diff

# Use GitGrove versions for safety
gg add .          # Instead of: git add .
gg commit -m "x"  # Instead of: git commit -m "x"

# Avoid these on repo branches
git checkout main  # Can break GitGrove state
git add .gg/       # Metadata should not be manually edited
```

## Troubleshooting

### "Not a git repository"
Ensure you're inside a Git repository. Run `git init` if needed.

### "Working tree is not clean"
Commit or stash your changes before GitGrove operations:
```bash
git stash
gg switch backend
git stash pop
```

### "Repo branch format invalid"
You're on the wrong branch. GitGrove commands like `add` and `commit` require you to be on a repository branch (format: `gitgroove/repos/<name>/branches/<branch>`).

### Marker file mismatch
The `.gitgroverepo` marker doesn't match your current branch. This usually means:
- You manually switched branches
- You're in the wrong directory

Solution:
```bash
gg switch <correct-repo>
```

## Next Steps

- Read the **[Full Documentation](FULL_DOCUMENTATION.md)** for advanced features
- See **[Technical Documentation](TECHNICAL.md)** to understand internals
- Explore edge cases and best practices in the full docs

## Quick Reference

```bash
# Setup
gg init                              # Initialize GitGrove
gg register <name>=<path>            # Register repository
gg link <child>=<parent>             # Create hierarchy

# Navigation
gg switch <repo> [branch]            # Switch to repository
gg up                                # Navigate to parent
gg down <child>                      # Navigate to child
gg ls                                # List children

# Changes
gg add <files>                       # Stage changes
gg commit -m "message"               # Commit changes
gg branch <branch-name>              # Create repo branch
gg checkout <repo> <branch> [--keep-empty-dirs] # Switch repo branch

# Information
gg info                              # Show status
gg move <repo> <new-path>            # Relocate repository
gg push <repo|*>                     # Push to remote
```
