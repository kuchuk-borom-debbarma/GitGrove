# GitGrove Technical Documentation

Deep dive into GitGrove's architecture, implementation, and design decisions.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Data Model](#data-model)
3. [Branch Strategy](#branch-strategy)
4. [Core Operations](#core-operations)
5. [Safety Mechanisms](#safety-mechanisms)
6. [Git Plumbing Details](#git-plumbing-details)
7. [Design Decisions](#design-decisions)
8. [Performance Considerations](#performance-considerations)

## Architecture Overview

### High-Level Design

GitGrove implements a **metadata-driven hierarchical repository system** on top of standard Git. It achieves isolation through:

1. **Metadata isolation**: All GitGrove state lives on a dedicated branch
2. **Branch isolation**: Each repository gets its own branch namespace
3. **Scope validation**: Programmatic checks prevent cross-repository changes

```
┌─────────────────────────────────────────┐
│         User Working Tree               │
│  (Flattened view of current repository) │
└─────────────────────────────────────────┘
              ↕ checkout
┌─────────────────────────────────────────┐
│    GitGrove Repository Branches         │
│  gitgroove/repos/<n>/branches/<b>       │
└─────────────────────────────────────────┘
              ↕ metadata references
┌─────────────────────────────────────────┐
│      System Branch (Metadata)           │
│    gitgroove/system                     │
│    (.gg/repos/*/path, parent, children) │
└─────────────────────────────────────────┘
```

### Component Layers

**Layer 1: Git Utilities** (`internal/util/git/`)
- Low-level Git operations
- Plumbing commands wrapper
- Error handling and normalization

**Layer 2: File Utilities** (`internal/util/file/`)
- Path normalization (cross-platform)
- File I/O operations
- Directory management

**Layer 3: Grove Core** (`internal/grove/`)
- High-level GitGrove operations
- Validation and safety checks
- Orchestration of git/file utilities

**Layer 4: API** (`core/api.go`)
- Public-facing interface
- Thin wrapper around grove core
- Stable API contract

**Layer 5: CLI** (`cmd/gg/`)
- Command-line interface
- Argument parsing
- User interaction

## Data Model

### Metadata Structure

GitGrove stores metadata in `.gg/` directory on the `gitgroove/system` branch:

```
.gg/
└── repos/
    ├── .gitkeep                    # Ensures directory is tracked
    ├── backend/
    │   ├── path                    # "services/backend"
    │   ├── parent                  # "shared"
    │   └── children/
    │       ├── auth-service        # Empty marker file
    │       └── payment-service     # Empty marker file
    ├── frontend/
    │   ├── path                    # "apps/frontend"
    │   └── parent                  # "shared"
    └── shared/
        ├── path                    # "libs/shared"
        └── children/
            ├── backend             # Empty marker file
            └── frontend            # Empty marker file
```

**File Formats:**
- `path`: Plain text, single line, relative path with `/` separators
- `parent`: Plain text, single line, parent repository name
- `children/<name>`: Empty file serving as existence marker

**Why this structure?**
- Simple text files = human-readable, git-friendly
- Flat structure = easy to parse, no complex JSON/YAML
- Children as separate files = supports incremental updates
- Git tracks everything = full history of structural changes

### System Root View

The `gitgroove/system` branch serves as the "System Root" view. In addition to the `.gg/` metadata directory, it contains:

- **Directory Stubs**: Empty directories (with `.gitkeep`) for each registered **root** repository.
- **Purpose**: Allows users to see and navigate to top-level repositories using standard shell commands (`ls`, `cd`).
- **Maintenance**:
    - `register`: Adds stub for new root repos.
    - `link`: Removes stub if a root repo becomes a child.

### In-Memory Model

```go
type Repo struct {
    Name   string // Unique identifier (e.g., "backend")
    Path   string // Relative path (e.g., "services/backend")
    Parent string // Parent repo name (empty if root)
}

type RepoState struct {
    Repo       Repo
    PathExists bool // Physical directory exists?
}

type RepoInfo struct {
    Repos map[string]RepoState
}
```

**Loading process:**
1. Read `gitgroove/system` branch at specific commit
2. List directories in `.gg/repos/`
3. For each directory (repo):
   - Read `path` file → `Repo.Path`
   - Read `parent` file (if exists) → `Repo.Parent`
   - Check physical path existence → `RepoState.PathExists`

### Marker Files

Each registered repository gets a `.gitgroverepo` marker file:

```
# File location: <repo-path>/.gitgroverepo
# Content: repository name
backend
```

**Purposes:**
1. **Identity**: Identifies which repository a directory belongs to
2. **Scope boundary**: Used by validation to detect nested repos
3. **Navigation**: Enables `up`/`down`/`ls` commands
4. **User visibility**: Makes GitGrove structure visible in file browser

**Lifecycle:**
- Created by `register` command
- Updated if repository is moved
- Should be committed to user branches (optional but recommended)

## Branch Strategy

### Branch Naming Convention

**System branch:**
```
gitgroove/system
```

**Repository branches:**
```
gitgroove/repos/<repo-name>/branches/<branch-name>
```

Examples:
```
gitgroove/repos/backend/branches/main
gitgroove/repos/backend/branches/feature-auth
gitgroove/repos/backend/branches/hotfix-critical
```

**Why this structure?**
- Clear namespace separation
- Avoids conflicts with user branches
- Supports multiple branches per repository
- Easy to list all branches for a repository

### Branch Content Model

**System branch commits:**
- **Tree**: Contains only `.gg/` directory
- **Parent**: Previous system commit (linear history)
- **Message**: Describes metadata change

**Repository branch commits:**
- **Tree**: Flattened repository content (repository files at root level)
- **Parent**: Previous commit on same branch
- **Message**: User's commit message

**Example tree comparison:**

Project HEAD tree:
```
.
├── .gg/
├── services/
│   └── backend/
│       ├── main.go
│       └── utils/
│           └── helper.go
└── libs/
    └── shared/
        └── lib.go
```

Repository branch tree (backend):
```
.
├── main.go
├── utils/
│   └── helper.go
└── .gitgroverepo
```

### Branch Creation

**Initial branch creation** (during `register`):

```go
// 1. Get repository subtree from HEAD
headCommit := getHeadCommit()
subtreeHash := getSubtreeHash(headCommit, repoPath)

// 2. Add marker file to tree
treeHash := addFileToTree(subtreeHash, ".gitgroverepo", repoName)

// 3. Create commit with this tree
commitHash := commitTree(treeHash, "Initial repo structure")

// 4. Set branch ref
setRef("refs/heads/gitgroove/repos/backend/branches/main", commitHash)
```

**New branch creation** (via `gg branch`):
- Same process, starting from current HEAD
- Copies repository's current state to new branch

## Core Operations

### Init

**Goal:** Bootstrap GitGrove metadata structure.

**Algorithm:**
```
1. Validate:
   - Is git repository?
   - Is working tree clean?
   - Does .gg/ NOT exist?
   - Does gitgroove/system NOT exist?

2. Create directories:
   - .gg/repos/
   - .gg/repos/.gitkeep

3. Create system branch:
   - git checkout -b gitgroove/system

4. Commit metadata:
   - git add .gg
   - git commit -m "Initialize GitGroove system branch"
```

**Result:** User is left on `gitgroove/system` branch.

---

### Register

**Goal:** Add repositories to GitGrove metadata.

**Algorithm:**
```
1. Validate environment:
   - Is git repository?
   - Is working tree clean?
   - Not on gitgroove/system (must be on user branch)

2. Load existing repos from gitgroove/system

3. Validate inputs:
   - Names: unique, valid pattern
   - Paths: exist, unique, no .git/, no escaping

4. Create temporary worktree at gitgroove/system tip:
   - git worktree add --detach <temp> <system-ref>

5. In temporary worktree:
   - Write .gg/repos/<name>/path files
   - Write .gitgroverepo marker files at repo paths
   - git add .gg/repos
   - git add <path>/.gitgroverepo (for each repo)
   - git commit -m "Register N repositories"

6. Atomically update system ref:
   - git update-ref gitgroove/system <new> <old>
   - CAS (compare-and-swap) prevents races

7. If currently on gitgroove/system:
   - git reset --hard HEAD (sync working tree)

8. Create orphan branches:
   - For each repo: gitgroove/repos/<n>/branches/main
   - Tree: repository subtree + .gitgroverepo marker
   - Commit: "Initial repo structure"
```

**Atomicity:** Uses git's CAS mechanism via `update-ref`.

**Rollback:** If update-ref fails, changes are not applied.

---

### Link

**Goal:** Create parent-child relationships.

**Algorithm:**
```
1. Validate environment (same as register)

2. Load existing repos

3. Validate relationships:
   - Both repos exist
   - No self-references
   - Child doesn't have parent already
   - No cycles (build full graph, DFS cycle detection)
   - Child path exists on disk

4. Create temporary worktree at system tip

5. In temporary worktree:
   - Write .gg/repos/<child>/parent files
   - Write .gg/repos/<parent>/children/<child> files
   - git add .gg/repos
   - git commit -m "Link N repositories"

6. Atomically update system ref (CAS)

7. If on gitgroove/system: sync working tree
```

**Cycle detection:**
```
graph = {}
for repo in existing:
    if repo.parent:
        graph[parent].append(repo.name)

for child, parent in new_relationships:
    graph[parent].append(child)

visited = {}
recursion_stack = {}

def has_cycle(node):
    visited[node] = True
    recursion_stack[node] = True
    
    for child in graph[node]:
        if not visited[child]:
            if has_cycle(child):
                return True
        elif recursion_stack[child]:
            return True  # Cycle detected!
    
    recursion_stack[node] = False
    return False

for node in graph:
    if not visited[node]:
        if has_cycle(node):
            return Error("cycle detected")
```

---

### Switch

**Goal:** Change working tree to repository's flattened view.

**Algorithm:**
```
1. Validate:
   - Is git repository?
   - Is working tree clean?

2. Checkout gitgroove/system:
   - git checkout gitgroove/system
   - This loads fresh metadata

3. Load repos from HEAD (now gitgroove/system)

4. Validate:
   - Repo exists?
   - Target branch exists?

5. Checkout target branch:
   - git checkout gitgroove/repos/<n>/branches/<b>

6. Working tree now shows flattened view

7. Configure sparse-checkout:
   - Identify nested registered repositories
   - If any found (and --flat not used):
     - git sparse-checkout set --no-cone "/*" "!nested/repo/"
   - Else:
     - git sparse-checkout disable

8. Clean up empty directories (unless requested otherwise)
   - Recursively remove empty directories to keep workspace clean

9. Ensure marker existence:
   - Check if .gitgroverepo exists
   - If missing, recreate it (self-healing)
```

**Critical:** Always checkout system branch first to ensure metadata is fresh.

---

### Add

**Goal:** Stage files with scope validation.

**Algorithm:**
```
1. Validate:
   - Is git repository?

2. Get current branch

3. Parse repo name from branch:
   - Must match: gitgroove/repos/<name>/branches/<branch>

4. Load repo metadata (without checkout):
   - Read gitgroove/system at tip

5. Expand file arguments to changed files:
   - git status --porcelain -u -z
   - Match input paths against changed files

6. Validate each file:
   - Not .gg/**
   - Not from nested repo (check for .gitgroverepo markers)

7. Stage valid files:
   - git add -- <file1> <file2> ...
```

**Key insight:** No need to checkout system branch, just read metadata from ref.

---

### Commit

**Goal:** Commit staged changes with validation.

**Algorithm:**
```
1. Validate:
   - Is git repository?

2. Get current branch and parse repo name

3. Load repo metadata

4. Verify .gitgroverepo marker matches:
   - Read .gitgroverepo
   - Compare to expected repo name

5. Validate staged files:
   - git diff --cached --name-only
   - Check each file not .gg/**

6. Commit:
   - git commit -m "<message>"
```

**Why validate marker?** Ensures you're in correct directory and haven't manually switched branches.

---

### Move

**Goal:** Relocate repository to new path.

**Algorithm:**
```
1. Validate environment

2. Load repos from system tip

3. Validate:
   - Repo exists?
   - New path available?
   - New path != old path

4. Physical move:
   - os.Rename(oldPath, newPath)

5. Update metadata:
   - Create temporary worktree at system tip
   - Update .gg/repos/<name>/path
   - git add .gg/repos
   - git commit -m "Move repo"

6. Atomically update system ref

7. If metadata update fails:
   - os.Rename(newPath, oldPath)  # Rollback
```

**Identity preservation:** Repository name, branches, and hierarchy unchanged.

---

## Safety Mechanisms

### 1. Clean State Requirement

**Rationale:** Prevents data loss and conflicting changes.

**Check:**
```go
func VerifyCleanState(path string) error {
    if IsDetachedHEAD(path) {
        return error("HEAD is detached")
    }
    if HasStagedChanges(path) {
        return error("staged changes exist")
    }
    if HasUnstagedChanges(path) {
        return error("unstaged changes exist")
    }
    if HasUntrackedFiles(path) {
        return error("untracked files exist")
    }
    return nil
}
```

**Applied to:** `init`, `register`, `link`, `switch`, `branch`, `move`

### 2. Atomic Metadata Updates

**Mechanism:** Git's `update-ref` with old value (CAS).

```bash
git update-ref refs/heads/gitgroove/system <new-hash> <old-hash>
```

**Behavior:**
- ✅ Succeeds if current value == old-hash
- ❌ Fails if current value != old-hash (concurrent modification)

**Protection against:**
- Race conditions (multiple simultaneous operations)
- Lost updates (overwriting concurrent changes)

### 3. Scope Validation

**For add/commit operations:**

```go
func validateStagingFiles(root, targetRepo, files) {
    for file in files:
        // 1. Not .gg/**
        if isMetadata(file):
            skip

        // 2. Not in nested repo
        if hasNestedRepoMarker(file):
            skip

        stage(file)
}

func hasNestedRepoMarker(file):
    dir = parent(file)
    while dir != root:
        if exists(dir + "/.gitgroverepo"):
            return true
        dir = parent(dir)
    return false
```

**Prevents:** Accidentally staging files from other repositories.

### 4. Input Validation

**Repository names:**
```go
validNameRegex = `^[a-zA-Z0-9._-]+$`
```

**Paths:**
- Must exist
- Must be directories
- Must be unique
- Must not contain `.git/`
- Must not escape project root
- Must not overlap (no repo contains another)

### 5. Cycle Detection

**Algorithm:** Depth-first search with recursion stack.

**Time complexity:** O(V + E) where V = repos, E = relationships

**Catches:** All cycles, including transitive ones (A→B→C→D→A)

### 6. Path Normalization

**Cross-platform safety:**

```go
func NormalizePath(path string) string {
    clean := filepath.Clean(path)        // Resolve .., ., //
    return strings.ReplaceAll(clean, "\\", "/")  // Force /
}
```

**Ensures:** Consistent path handling across Windows/Linux/macOS.

## Git Plumbing Details

### Commands Used

**Porcelain (high-level):**
- `git init`
- `git add`
- `git commit`
- `git checkout`
- `git branch`

**Plumbing (low-level):**
- `git rev-parse` - Resolve refs, check state
- `git update-ref` - Atomic ref updates
- `git worktree add/remove` - Temporary workspaces
- `git ls-tree` - List tree contents
- `git show` - Read file at specific commit
- `git diff` - Check for changes
- `git hash-object` - Create blobs
- `git commit-tree` - Create commits directly
- `git write-tree` - Create tree objects
- `git update-index` - Modify index without working tree

### Temporary Worktrees

**Purpose:** Manipulate system branch without affecting user's working tree.

**Creation:**
```bash
git worktree add --detach <path> <ref>
```

**Usage flow:**
```
1. Create worktree at specific commit
2. Make changes in worktree
3. Commit changes
4. Update ref to new commit
5. Remove worktree
```

**Benefits:**
- User's working tree untouched
- Can work on detached state safely
- Automatically cleaned up

### Tree Manipulation

**Challenge:** Create commits with custom tree structures without affecting working tree.

**Solution:** Use plumbing commands to build trees programmatically.

**Example (flattened tree creation):**

```bash
# 1. Get repository subtree hash from HEAD
REPO_TREE=$(git rev-parse HEAD:services/backend)

# 2. Create blob for marker file
MARKER_BLOB=$(echo "backend" | git hash-object -w --stdin)

# 3. Create temporary index
export GIT_INDEX_FILE=.git/index.temp

# 4. Read base tree into index
git read-tree $REPO_TREE

# 5. Add marker file to index
git update-index --add --cacheinfo 100644 $MARKER_BLOB .gitgroverepo

# 6. Write tree
NEW_TREE=$(git write-tree)

# 7. Create commit
NEW_COMMIT=$(git commit-tree $NEW_TREE -m "Initial repo structure")

# 8. Update ref
git update-ref refs/heads/gitgroove/repos/backend/branches/main $NEW_COMMIT
```

This is what `CreateRepoBranch` does internally.

### CAS via update-ref

**Standard ref update (unsafe):**
```bash
git update-ref refs/heads/mybranch $NEW_HASH
```
→ Always succeeds, can overwrite concurrent changes

**CAS ref update (safe):**
```bash
git update-ref refs/heads/mybranch $NEW_HASH $OLD_HASH
```
→ Only succeeds if current value is $OLD_HASH

**GitGrove usage:**
```go
oldTip := resolveRef("gitgroove/system")
// ... make changes ...
newTip := createCommit()
err := updateRef("gitgroove/system", newTip, oldTip)
if err != nil {
    // Concurrent modification detected!
}
```

## Design Decisions

### Why Not Git Submodules?

**Submodules:**
- ✅ Native Git feature
- ❌ Separate repositories (different histories)
- ❌ Complex to manage (nested .git/)
- ❌ No flattened view
- ❌ No hierarchical relationships

**GitGrove:**
- ✅ Single repository (unified history)
- ✅ Flattened views
- ✅ Explicit hierarchy
- ❌ Requires tool/CLI

### Why Not Git Worktrees (alone)?

**Worktrees:**
- ✅ Multiple working trees
- ❌ No automatic scoping
- ❌ No hierarchy
- ❌ No validation

**GitGrove:**
- Uses worktrees internally (for atomic updates)
- Adds scoping and validation on top

### Why Custom Branch Namespace?

**Alternative:** Use regular branches like `backend`, `frontend`.

**Problems:**
- Conflicts with user branch names
- No clear separation of concerns
- Hard to list all branches for a repo

**GitGrove approach:**
- Clear namespace: `gitgroove/repos/<n>/branches/<b>`
- No conflicts
- Easy to query (list all branches with prefix)

### Why Marker Files?

**Alternative:** Track everything in `.gg/` metadata.

**Problems:**
- Scope validation requires metadata lookups
- Not visible to users browsing files
- Can't detect nested repos without parsing all paths

**Marker file benefits:**
- Fast scope checks (just walk up directory tree)
- Visible identity (users see which repo they're in)
- Enables filesystem-based detection

### Why System Branch?

**Alternative:** Store metadata as JSON/config file in main branch.

**Problems:**
- Metadata changes appear in every branch
- Conflicts during merges
- Can't have different metadata per branch

**System branch benefits:**
- Isolated metadata history
- No interference with user branches
- Can experiment with metadata without affecting code

### Why Flat Text Files?

**Alternative:** JSON/YAML/binary format.

**Flat text benefits:**
- Human-readable (cat .gg/repos/backend/path)
- Git-friendly (good diffs, merging)
- Simple parsing (no library needed)
- Extensible (just add new files)

**Tradeoffs:**
- More files (but not many)
- No schema validation (must validate in code)

## Performance Considerations

### Metadata Loading

**Current approach:**
```go
entries := listTree(systemRef, ".gg/repos")
for name in entries:
    path := showFile(systemRef, ".gg/repos/" + name + "/path")
    parent := showFile(systemRef, ".gg/repos/" + name + "/parent")
```

**Time complexity:** O(N) where N = number of repos

**Git operations:** O(N) × `git show` calls

**Optimization opportunities:**
1. Batch reads using `git cat-file --batch`
2. Cache metadata in memory
3. Incremental loading (only changed repos)

**Current performance:** Fine for <100 repos. Optimization needed for 100+.

### Temporary Worktrees

**Overhead:**
- Create: ~50-100ms
- Remove: ~10-20ms

**Usage frequency:**
- `register`: 1 worktree per operation
- `link`: 1 worktree per operation
- `move`: 1 worktree per operation

**Optimization:** Reuse worktree for batch operations (currently not implemented).

### Branch Creation

**Current approach:** Create subtree-based commits.

**Cost:**
- Extract subtree: O(tree size)
- Create blob for marker: O(1)
- Build tree: O(files in repo)
- Create commit: O(1)

**Total:** O(files in repo) – acceptable for most repos.

### Scope Validation

**Check for nested repos:**
```
Walk up from file to root:
    Check for .gitgroverepo marker
```

**Worst case:** O(depth × N files)

**Typical case:** O(depth) – only one check per unique directory

**Optimization:** Cache directory→repo mappings.

## Implementation Notes

### Error Handling

**Strategy:** Fail fast with descriptive errors.

**Example:**
```go
func Register(root, repos) error {
    // Validate ALL inputs before making ANY changes
    if err := validateEnvironment(root); err != nil {
        return err
    }
    if err := validateRepos(repos); err != nil {
        return err
    }
    
    // Only proceed if all validations pass
    return applyChanges(root, repos)
}
```

**Rollback:** Limited (physical directory moves are rolled back on metadata failure).

### Logging

**Levels:**
- `Debug`: Internal state, detailed flow
- `Info`: Major operations, user-visible actions
- `Warn`: Skipped files, recoverable issues
- `Error`: Failures, validation errors

**Library:** `zerolog` (structured, fast)

### Testing

**Unit tests:** Cover core logic (validation, cycle detection, path normalization)

**Integration tests:** Full workflows (init → register → link → switch)

**Test utilities:**
- `createDummyProject`: Setup test repository
- `execGit`: Run git commands in tests
- `t.TempDir()`: Isolated test environments

## Future Enhancements

### Potential Features

1. **Batch operations:**
   ```bash
   gg register -f repos.txt
   ```

2. **Merge support:**
   - Merge repo branches while maintaining scope
   - Conflict resolution within repository boundaries

3. **Stash support:**
   ```bash
   gg stash
   gg switch other-repo
   gg stash pop
   ```

4. **Remote synchronization:**
   - Push/pull repo branches
   - Sync system branch metadata

5. **Performance optimizations:**
   - Metadata caching
   - Batch git operations
   - Incremental validation

6. **Repo aliasing:**
   ```bash
   gg alias be=backend
   gg switch be
   ```

7. **Status command:**
   ```bash
   gg status
   # Shows: current repo, dirty files, branch
   ```

### Non-Goals

- ❌ Supporting nested .git directories
- ❌ Automatic conflict resolution across repos
- ❌ Submodule compatibility layer
- ❌ GUI/web interface (CLI-first design)

## Contributing

### Code Organization

```
core/
├── internal/
│   ├── grove/              # Core operations
│   │   ├── init.go
│   │   ├── register.go
│   │   ├── link.go
│   │   ├── switch.go
│   │   ├── add.go
│   │   ├── commit.go
│   │   ├── model/          # Data structures
│   │   └── info/           # Status/info commands
│   └── util/               # Utilities
│       ├── git/            # Git operations
│       └── file/           # File operations
├── api.go                  # Public API
└── go.mod

cmd/
└── gg/
    └── main.go             # CLI entry point
```

### Adding New Commands

1. **Add core function** in `internal/grove/<command>.go`:
   ```go
   func NewCommand(rootAbsPath string, args...) error {
       // 1. Validate environment
       // 2. Load metadata
       // 3. Perform operation
       // 4. Update metadata (if needed)
       return nil
   }
   ```

2. **Expose via API** in `core/api.go`:
   ```go
   func NewCommand(rootAbsPath string, args...) error {
       return grove.NewCommand(rootAbsPath, args...)
   }
   ```

3. **Add CLI handler** in `cmd/gg/main.go`:
   ```go
   case "newcommand":
       err = core.NewCommand(root, args...)
   ```

4. **Add tests** in `internal/grove/<command>_test.go`

5. **Update documentation**

### Testing Guidelines

- **Always use `t.TempDir()`** for test isolation
- **Create clean git state** before each test
- **Test error cases** as well as happy paths
- **Use test helpers** (`execGit`, `createDummyProject`)

### Pull Request Checklist

- [ ] Tests pass
- [ ] Code follows existing style
- [ ] Documentation updated
- [ ] Commit messages are descriptive
- [ ] No breaking changes (or clearly documented)

## Appendix: Git Internals Primer

### Objects

- **Blob**: File content
- **Tree**: Directory structure
- **Commit**: Snapshot with metadata
- **Tag**: Named reference to commit

### References

- **Branch**: Mutable pointer to commit (`refs/heads/main`)
- **Tag**: Immutable pointer (`refs/tags/v1.0`)
- **HEAD**: Current branch/commit (`.git/HEAD`)

### Working Tree vs Index vs Objects

```
Working Tree → Index → Objects
(files)      (staging) (committed)

git add:  Working Tree → Index
git commit: Index → Objects
git checkout: Objects → Working Tree + Index
```

### Detached HEAD

Normal:
```
HEAD → refs/heads/main → commit-hash
```

Detached:
```
HEAD → commit-hash
```

GitGrove uses detached worktrees for metadata manipulation.

## References

- [Git Internals](https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain)
- [Git worktrees](https://git-scm.com/docs/git-worktree)
- [Git update-ref](https://git-scm.com/docs/git-update-ref)
- [Go filepath package](https://pkg.go.dev/path/filepath)

