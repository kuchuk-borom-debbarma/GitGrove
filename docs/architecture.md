# GitGrove Architecture Documentation

Comprehensive architectural overview of GitGrove's design, implementation patterns, and system behavior.

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Architecture](#system-architecture)
3. [Component Design](#component-design)
4. [Data Architecture](#data-architecture)
5. [Branch Architecture](#branch-architecture)
6. [Operation Patterns](#operation-patterns)
7. [Concurrency and Atomicity](#concurrency-and-atomicity)
8. [Security and Validation](#security-and-validation)
9. [Performance Architecture](#performance-architecture)
10. [Extensibility and Modularity](#extensibility-and-modularity)

---

## Executive Summary

### Vision

GitGrove is a **metadata-driven hierarchical repository management system** built on Git. It enables developers to work with logically isolated repositories within a single Git repository, providing the benefits of monorepo organization with the isolation of polyrepo development.

### Core Principles

1. **Git-Native**: Built entirely on Git primitives, no database required
2. **Non-Destructive**: Never modifies user data or branches without explicit action
3. **Atomic**: All metadata operations succeed completely or fail without side effects
4. **Isolated**: Each repository operates in its own logical namespace
5. **Transparent**: All state is stored in Git and inspectable with standard tools

### Key Architectural Decisions

| Decision | Rationale |
|----------|-----------|
| **Single Git Repository** | Unified history, simplified tooling, atomic cross-repo operations |
| **Metadata Branch** | Isolated metadata history, no pollution of user branches |
| **Flattened Views** | Repository isolation without filesystem complexity |
| **Branch Namespaces** | Clear separation, no conflicts with user branches |
| **Text-Based Metadata** | Human-readable, Git-friendly, simple parsing |
| **Marker Files** | Fast scope detection, user-visible structure |

### System Boundaries

**What GitGrove Does:**
- Organizes code into logical repositories within a single Git repo
- Provides isolated working views of repositories
- Enforces scope boundaries during staging/committing
- Manages hierarchical relationships between repositories
- Tracks metadata history alongside code history

**What GitGrove Doesn't Do:**
- Replace Git (it's a layer on top)
- Manage multiple physical repositories
- Provide distributed repository coordination
- Handle binary artifact management
- Implement build/deployment systems

---

## System Architecture

### Layered Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     CLI Layer (cmd/gg)                  │
│  • Argument parsing  • User interaction  • Help text    │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                   API Layer (core/api.go)               │
│  • Public interface  • Stable contracts  • Thin wrapper │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│            Core Business Logic (internal/grove)          │
│  • Operations orchestration  • Validation  • Safety     │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│              Utilities Layer (internal/util)             │
│  Git Operations (git/)        File Operations (file/)   │
│  • Plumbing commands          • Path normalization      │
│  • Ref management             • Directory management    │
│  • Tree manipulation          • File I/O                │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                      Git Engine                          │
│  • Object database  • Refs  • Index  • Worktrees        │
└─────────────────────────────────────────────────────────┘
```

### Data Flow

**Read Path (Query Operations):**
```
User Command
    ↓
CLI Parser
    ↓
API Gateway
    ↓
Core Logic: Load Metadata
    ↓
Git Util: Read refs/heads/gitgroove/system
    ↓
Git Util: Parse .gg/repos/* files
    ↓
Core Logic: Build in-memory model
    ↓
Core Logic: Execute query logic
    ↓
Format & Return to User
```

**Write Path (Mutation Operations):**
```
User Command
    ↓
CLI Parser
    ↓
API Gateway
    ↓
Core Logic: Validate Environment
    ↓
Core Logic: Load Current Metadata
    ↓
Core Logic: Validate Inputs
    ↓
Git Util: Create Temporary Worktree
    ↓
Core Logic: Apply Changes in Worktree
    ↓
Git Util: Commit Changes
    ↓
Git Util: Atomic Update-Ref (CAS)
    ↓
Core Logic: Cleanup Worktree
    ↓
Return Success/Failure to User
```

### Process Model

GitGrove operates as a **single-process, synchronous CLI tool**:

- **No daemon**: Every command is a separate process
- **No state persistence**: All state lives in Git
- **No cache**: Metadata loaded fresh each time
- **No networking**: Pure local operations

**Concurrency Model:**
- Multiple GitGrove processes CAN run simultaneously
- Conflicts detected via Git's CAS mechanism
- Last-write-wins for file operations
- No distributed coordination needed

---

## Component Design

### 1. CLI Layer (`cmd/gg`)

**Responsibilities:**
- Parse command-line arguments
- Validate argument formats (basic)
- Dispatch to appropriate API calls
- Format output for terminal
- Handle user interrupts (Ctrl+C)

**Architecture Pattern:** Command Pattern

```go
func main() {
    command := os.Args[1]
    
    switch command {
    case "init":
        handleInit()
    case "register":
        handleRegister()
    case "switch":
        handleSwitch()
    // ...
    }
}

func handleInit() {
    root := getRootPath()
    if err := core.Init(root); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("GitGrove initialized successfully")
}
```

**Design Notes:**
- Thin layer, minimal logic
- No direct Git operations
- All business logic delegated to Core
- User-facing error messages

---

### 2. API Layer (`core/api.go`)

**Responsibilities:**
- Provide stable public interface
- Version boundaries (future: versioned APIs)
- Type conversions (internal ↔ external models)
- Documentation surface

**Architecture Pattern:** Facade Pattern

```go
// Public API
func Init(absolutePath string) error {
    return grove.Init(absolutePath)
}

func Register(rootAbsPath string, repos map[string]string) error {
    return grove.Register(rootAbsPath, repos)
}

// Public type (different from internal model)
type Repo struct {
    Name   string
    Path   string
    Parent string
}

func GetRepositories(rootAbsPath string) ([]Repo, error) {
    repoInfo, err := info.GetRepoInfo(rootAbsPath)
    if err != nil {
        return nil, err
    }
    
    // Convert internal model to public model
    var repos []Repo
    for _, state := range repoInfo.Repos {
        repos = append(repos, Repo{
            Name:   state.Repo.Name,
            Path:   state.Repo.Path,
            Parent: state.Repo.Parent,
        })
    }
    return repos, nil
}
```

**Design Notes:**
- No business logic
- Type conversions only
- Stable API even if internals change
- Future: versioning (v1, v2)

---

### 3. Core Business Logic (`internal/grove`)

**Responsibilities:**
- Orchestrate operations
- Enforce business rules
- Validate inputs and state
- Ensure safety guarantees
- Coordinate Git/File utilities

**Architecture Pattern:** Service Layer + Domain Model

**Sub-components:**

#### 3.1. Operations (`*.go` files)

Each operation is a separate file implementing a specific command:

```
init.go       - Initialize GitGrove
register.go   - Register repositories
link.go       - Create hierarchy
switch.go     - Change repository view
add.go        - Stage files with validation
commit.go     - Commit with validation
branch.go     - Create repository branches
checkout.go   - Checkout repository branches
move.go       - Relocate repositories
up.go         - Navigate to parent
down.go       - Navigate to child
cd.go         - Navigate (up/down wrapper)
ls.go         - List children
```

**Standard Operation Structure:**
```go
func Operation(rootAbsPath string, args...) error {
    // 1. Validate Environment
    if err := validateEnvironment(rootAbsPath); err != nil {
        return err
    }
    
    // 2. Load Current State
    state, err := loadState(rootAbsPath)
    if err != nil {
        return err
    }
    
    // 3. Validate Inputs
    if err := validateInputs(args, state); err != nil {
        return err
    }
    
    // 4. Execute Operation (with atomicity)
    newState, err := executeAtomically(rootAbsPath, state, args)
    if err != nil {
        return err
    }
    
    // 5. Post-Operation Sync (if needed)
    if err := syncState(rootAbsPath, newState); err != nil {
        return err
    }
    
    return nil
}
```

#### 3.2. Model (`model/`)

Domain entities representing core concepts:

```go
// model/model.go
package model

type Repo struct {
    Name   string `json:"name"`
    Path   string `json:"path"`
    Parent string `json:"parent,omitempty"`
}

const DefaultRepoBranch = "main"
```

**Design Notes:**
- Simple, immutable data structures
- JSON tags for potential serialization
- No behavior (data only)

#### 3.3. Info (`info/`)

Query-side logic for reading and presenting state:

```go
// info/repo_info.go
type RepoState struct {
    Repo       model.Repo
    PathExists bool
}

type RepoInfo struct {
    Repos map[string]RepoState
}

func GetRepoInfo(rootAbsPath string) (*RepoInfo, error) {
    // Load from gitgroove/system
    // Build in-memory representation
    // Check physical state
    return &RepoInfo{...}, nil
}
```

```go
// info/link_info.go
type TreeNode struct {
    State    RepoState
    Children []*TreeNode
}

type LinkInfo struct {
    Roots []*TreeNode
}

func GetLinkInfo(repoInfo *RepoInfo) *LinkInfo {
    // Build tree structure from flat data
    // Sort for consistent output
    return &LinkInfo{...}
}
```

**Design Notes:**
- Separation of concerns (repo data vs hierarchy representation)
- Builder pattern for complex structures
- Read-only operations

#### 3.4. Branch Ref Management (`branch_ref.go`)

Utilities for constructing and parsing GitGrove branch names:

```go
const branchPrefix = "refs/heads/gitgroove/repos/"

func RepoBranchRef(repoName, branchName string) string {
    return fmt.Sprintf("%s%s/branches/%s", branchPrefix, repoName, branchName)
}

func ParseRepoFromBranch(branchName string) (string, error) {
    // Extract repo name from branch path
    // Validate format
    return repoName, nil
}
```

**Design Pattern:** Static Helper Functions

---

### 4. Git Utilities (`internal/util/git`)

**Responsibilities:**
- Wrap Git commands
- Normalize Git output
- Handle Git errors
- Provide Git plumbing operations

**Architecture Pattern:** Adapter Pattern (wraps Git CLI)

**Core Functions:**

```go
// Basic Operations
func Init(repoPath string) error
func IsInsideGitRepo(path string) bool
func GetCurrentBranch(repoPath string) (string, error)
func Checkout(repoPath, branch string, keepEmptyDirs bool) error

// Validation
func VerifyCleanState(path string) error
func IsDetachedHEAD(path string) bool
func HasStagedChanges(path string) bool
func HasUnstagedChanges(path string) bool
func HasUntrackedFiles(path string) bool

// References
func ResolveRef(repoPath, ref string) (string, error)
func RefExists(repoPath, ref string) bool
func UpdateRef(repoPath, ref, newHash, oldHash string) error  // CAS
func SetRef(repoPath, ref, newHash string) error

// Branches
func HasBranch(path, branch string) (bool, error)
func CreateBranch(repoPath, branch, startPoint string) error
func CreateAndCheckoutBranch(path, branch string) error

// Worktrees
func WorktreeAddDetached(repoPath, worktreePath, ref string) error
func WorktreeRemove(repoPath, worktreePath string) error

// Trees and Objects
func GetSubtreeHash(repoPath, ref, path string) (string, error)
func CreateBlob(repoPath, content string) (string, error)
func CreateTreeWithFile(repoPath, relPath, content string) (string, error)
func AddFileToTree(repoPath, baseTreeHash, filename, content string) (string, error)
func CommitTree(repoPath, treeHash, message string, parents ...string) (string, error)

// Files
func ShowFile(repoPath, ref, filePath string) (string, error)
func ListTree(repoPath, ref, path string) ([]string, error)
func GetStagedFiles(repoPath string) ([]string, error)

// Staging and Committing
func StagePath(repoPath, relativePath string) error
func UnstagePath(repoPath, path string) error
func Commit(repoPath, message string) error
```

**Implementation Pattern:**
```go
func runGit(dir string, args ...string) (string, error) {
    cmd := exec.Command("git", args...)
    cmd.Dir = dir
    
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    
    err := cmd.Run()
    return strings.TrimSpace(out.String()), err
}

func IsInsideGitRepo(path string) bool {
    out, err := runGit(path, "rev-parse", "--is-inside-work-tree")
    return err == nil && out == "true"
}
```

**Design Notes:**
- Each function wraps one Git operation
- Consistent error handling
- Output normalization (trim whitespace)
- No business logic

**Error Handling Strategy:**
```go
// Distinguish between "not found" and "error"
func HasBranch(path, branch string) (bool, error) {
    _, err := runGit(path, "rev-parse", "--verify", "--quiet", branch)
    
    if err == nil {
        return true, nil  // Branch exists
    }
    
    if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
        return false, nil  // Branch doesn't exist (not an error)
    }
    
    return false, err  // Actual error
}
```

---

### 5. File Utilities (`internal/util/file`)

**Responsibilities:**
- File system operations
- Path normalization
- Directory management
- Cross-platform compatibility

**Architecture Pattern:** Utility Module

**Core Functions:**

```go
// Path Operations
func NormalizePath(path string) string  // Cross-platform path normalization
func Exists(path string) bool
func EnsureNotExists(path string) error

// Directory Operations
func CreateDir(path string) error  // mkdir -p

// File Operations
func CreateEmptyFile(path string) error
func WriteTextFile(path string, content string) error
func ReadTextFile(path string) (string, error)
func AppendTextFile(path, content string) error
func WriteJSONFile(path string, v any) error

// Utilities
func RandomString(n int) string  // For temporary file names
```

**Path Normalization (Critical for Cross-Platform):**

```go
func NormalizePath(path string) string {
    if path == "" {
        return ""
    }
    clean := filepath.Clean(path)  // Resolve .., ., remove extra separators
    return strings.ReplaceAll(clean, "\\", "/")  // Force forward slashes
}
```

**Why Normalization Matters:**
- Windows uses backslashes (`\`)
- Linux/macOS use forward slashes (`/`)
- GitGrove metadata stores paths with forward slashes
- Comparisons must be consistent

**Design Notes:**
- All file operations auto-create parent directories
- Consistent error messages
- No Git-specific logic

---

## Data Architecture

### Metadata Storage Model

**Location:** `.gg/` directory on `gitgroove/system` branch

**Structure:**
```
.gg/
└── repos/
    ├── .gitkeep
    ├── <repo-name-1>/
    │   ├── path              # Plain text: relative/path/to/repo
    │   ├── parent            # Plain text: parent-repo-name (optional)
    │   └── children/
    │       ├── <child-1>     # Empty file (existence marker)
    │       └── <child-2>     # Empty file (existence marker)
    ├── <repo-name-2>/
    │   ├── path
    │   └── parent
    └── <repo-name-3>/
        └── path
```

**File Format Specifications:**

| File | Format | Example | Purpose |
|------|--------|---------|---------|
| `path` | Single line, relative path, `/` separator | `services/backend` | Maps name to location |
| `parent` | Single line, parent repo name | `shared` | Defines hierarchy edge |
| `children/<name>` | Empty file | (empty) | Inverse pointer for navigation |

**Design Rationale:**

1. **Plain Text:** Human-readable, Git-friendly diffs
2. **Flat Structure:** Easy to list (`ls .gg/repos/`)
3. **Separate Files:** Each concern is independent
4. **Children Directory:** Scales to many children, easy to add/remove
5. **No JSON/YAML:** Simpler parsing, no library dependency

### Metadata Evolution

**Loading Algorithm:**
```go
func loadExistingRepos(root, ref string) (map[string]model.Repo, error) {
    // 1. List all repo directories
    entries, err := gitUtil.ListTree(root, ref, ".gg/repos")
    if err != nil {
        return map[string]model.Repo{}, nil  // Empty if doesn't exist
    }
    
    repos := make(map[string]model.Repo)
    
    // 2. For each repo directory
    for _, name := range entries {
        if name == ".gitkeep" {
            continue
        }
        
        // 3. Read path file (required)
        pathFile := fmt.Sprintf(".gg/repos/%s/path", name)
        content, err := gitUtil.ShowFile(root, ref, pathFile)
        if err != nil {
            return nil, fmt.Errorf("failed to read path for repo %s: %w", name, err)
        }
        repoPath := strings.TrimSpace(content)
        
        // 4. Read parent file (optional)
        parentFile := fmt.Sprintf(".gg/repos/%s/parent", name)
        parentContent, err := gitUtil.ShowFile(root, ref, parentFile)
        parent := ""
        if err == nil {
            parent = strings.TrimSpace(parentContent)
        }
        
        // 5. Build repo object
        repos[name] = model.Repo{
            Name:   name,
            Path:   repoPath,
            Parent: parent,
        }
    }
    
    return repos, nil
}
```

**Time Complexity:** O(N) where N = number of repositories

**Future Optimization:** Batch read using `git cat-file --batch`

### Marker Files

**Purpose:** Repository boundary markers in working tree

**Location:** `<repo-path>/.gitgroverepo`

**Content:** Single line containing repository name
```
backend
```

**Lifecycle:**
1. **Created:** During `register` operation
2. **Updated:** During `move` operation (content stays same, location changes)
3. **Committed:** User's choice (recommended for full functionality)

**Usage:**

```go
// Check if path is inside a nested repo
func checkNestedRepo(rootAbsPath, fileAbsPath string) error {
    dir := filepath.Dir(fileAbsPath)
    root := filepath.Clean(rootAbsPath)
    current := filepath.Clean(dir)
    
    for {
        if current == root || len(current) < len(root) {
            break
        }
        
        markerPath := filepath.Join(current, ".gitgroverepo")
        if fileUtil.Exists(markerPath) {
            rel, _ := filepath.Rel(root, current)
            return fmt.Errorf("belongs to nested repo '%s'", rel)
        }
        
        parent := filepath.Dir(current)
        if parent == current {
            break
        }
        current = parent
    }
    
    return nil
}
```

**Performance:** O(depth) per file, where depth = directory nesting level

---

## Branch Architecture

### Branch Namespace Design

**System Branch:**
```
refs/heads/gitgroove/system
```

**Repository Branches:**
```
refs/heads/gitgroove/repos/<repo-name>/branches/<branch-name>
```

**Examples:**
```
refs/heads/gitgroove/system
refs/heads/gitgroove/repos/backend/branches/main
refs/heads/gitgroove/repos/backend/branches/feature-auth
refs/heads/gitgroove/repos/backend/branches/develop
refs/heads/gitgroove/repos/frontend/branches/main
refs/heads/gitgroove/repos/shared/branches/main
```

**Namespace Hierarchy:**
```
refs/heads/
├── main                                  (user branch)
├── develop                               (user branch)
└── gitgroove/
    ├── system                            (metadata)
    └── repos/
        ├── backend/
        │   └── branches/
        │       ├── main
        │       ├── develop
        │       └── feature-auth
        ├── frontend/
        │   └── branches/
        │       └── main
        └── shared/
            └── branches/
                └── main
```

### Branch Content Model

**System Branch Tree:**
```
<commit>
└── <tree>
    └── .gg/
        └── repos/
            ├── .gitkeep
            ├── backend/
            │   ├── path
            │   └── parent
            └── frontend/
                └── path
```

**Repository Branch Tree (Flattened):**
```
<commit>
└── <tree>
    ├── .gitgroverepo
    ├── main.go
    ├── utils/
    │   └── helper.go
    └── config/
        └── app.yaml
```

**Key Difference:** Repository branches show ONLY that repository's content at root level, not the full project structure.

### Branch Creation Process

**Initial Repository Branch (during `register`):**

```
1. Get HEAD commit hash
   HEAD_COMMIT = git rev-parse HEAD

2. Extract repository subtree
   REPO_TREE = git rev-parse HEAD:services/backend

3. Create marker blob
   MARKER_BLOB = echo "backend" | git hash-object -w --stdin

4. Build new tree with marker
   export GIT_INDEX_FILE=.git/index.temp
   git read-tree $REPO_TREE
   git update-index --add --cacheinfo 100644 $MARKER_BLOB .gitgroverepo
   NEW_TREE = git write-tree

5. Create commit with new tree
   NEW_COMMIT = git commit-tree $NEW_TREE -m "Initial repo structure"

6. Create branch ref
   git update-ref refs/heads/gitgroove/repos/backend/branches/main $NEW_COMMIT
```

**New Branch (via `gg branch`):**
- Same process, but starting from current HEAD's subtree
- Preserves current state of repository

### Branch Relationships

**Independence:** Repository branches are INDEPENDENT of each other and user branches.

```
main:                  A---B---C---D
                                    \
gitgroove/system:                    X---Y---Z (metadata history)

gitgroove/repos/backend/branches/main:        P---Q---R (backend history)
gitgroove/repos/frontend/branches/main:           S---T (frontend history)
```

**No Shared History:** Each branch namespace has its own commit graph.

**Benefit:** Clear separation, no merge conflicts between metadata and code.

---

## Operation Patterns

### Pattern 1: Metadata Mutation with Atomicity

Used by: `register`, `link`, `move`

**Flow:**
```
1. Validate Environment
   ├─ Is Git repo?
   ├─ Is working tree clean?
   └─ Not on system branch?

2. Load Current Metadata
   ├─ Resolve gitgroove/system ref → OLD_TIP
   └─ Parse .gg/repos/* → CURRENT_STATE

3. Validate Inputs
   ├─ Names/paths unique?
   ├─ No cycles?
   └─ Repos exist?

4. Create Temporary Worktree
   └─ git worktree add --detach <temp> OLD_TIP

5. Apply Changes in Worktree
   ├─ Modify .gg/repos/* files
   ├─ git add .gg/repos
   └─ git commit -m "..." → NEW_TIP

6. Atomic Update
   ├─ git update-ref gitgroove/system NEW_TIP OLD_TIP
   └─ [Succeeds only if OLD_TIP unchanged]

7. Cleanup
   └─ git worktree remove --force <temp>

8. Sync Working Tree (if on system branch)
   └─ git reset --hard HEAD
```

**Key Guarantees:**
- Atomicity via CAS (compare-and-swap)
- No working tree disruption
- Full rollback on failure (worktree discarded)
- Concurrent operation detection

**Example (Register):**
```go
func Register(rootAbsPath string, repos map[string]string) error {
    // 1. Validate
    if err := validateRegisterEnvironment(rootAbsPath); err != nil {
        return err
    }
    
    // 2. Load current state
    systemRef := "refs/heads/gitgroove/system"
    oldTip, err := gitUtil.ResolveRef(rootAbsPath, systemRef)
    if err != nil {
        return fmt.Errorf("failed to resolve %s: %w", systemRef, err)
    }
    
    existingRepos, err := loadExistingRepos(rootAbsPath, oldTip)
    if err != nil {
        return err
    }
    
    // 3. Validate inputs
    if err := validateNewRepos(rootAbsPath, repos, existingRepos); err != nil {
        return err
    }
    
    // 4-5. Create commit via temporary worktree
    newTip, err := createRegisterCommit(rootAbsPath, oldTip, repos)
    if err != nil {
        return err
    }
    
    // 6. Atomic update
    if err := gitUtil.UpdateRef(rootAbsPath, systemRef, newTip, oldTip); err != nil {
        return fmt.Errorf("failed to update %s (concurrent modification?): %w", systemRef, err)
    }
    
    // 8. Sync if needed
    currentBranch, _ := gitUtil.GetCurrentBranch(rootAbsPath)
    if currentBranch == "gitgroove/system" {
        gitUtil.ResetHard(rootAbsPath, "HEAD")
    }
    
    return nil
}
```

---

### Pattern 2: View Switching with Metadata Refresh

Used by: `switch`, `checkout`

**Flow:**
```
1. Validate Environment
   ├─ Is Git repo?
   └─ Is working tree clean?

2. Load Fresh Metadata
   ├─ git checkout gitgroove/system  [CRITICAL STEP]
   └─ Parse .gg/repos/* from working tree

3. Validate Target
   ├─ Repo exists?
   └─ Branch exists?

4. Checkout Target Branch
   └─ git checkout gitgroove/repos/<n>/branches/<b>

5. User Now Sees Flattened View
```

**Critical Design Point:** ALWAYS checkout system branch first.

**Rationale:**
- Metadata on user branches may be stale or absent
- System branch is the source of truth
- Guarantees consistent behavior
- Prevents using outdated metadata

**Example (Switch):**
```go
func Switch(rootAbsPath, repoName, branch string) error {
    // 1. Validate
    if err := validateSwitchEnvironment(rootAbsPath); err != nil {
        return err
    }
    
    // 2. CRITICAL: Load fresh metadata
    log.Info().Msg("Checking out gitgroove/system to load metadata")
    if err := gitUtil.Checkout(rootAbsPath, "gitgroove/system"); err != nil {
        return fmt.Errorf("failed to checkout gitgroove/system: %w", err)
    }
    
    repos, err := loadExistingRepos(rootAbsPath, "HEAD")
    if err != nil {
        return err
    }
    
    // 3. Validate target
    if _, ok := repos[repoName]; !ok {
        return fmt.Errorf("repo '%s' not found", repoName)
    }
    
    targetBranchName := branch
    if targetBranchName == "" {
        targetBranchName = model.DefaultRepoBranch
    }
    
    fullBranchRef := RepoBranchRef(repoName, targetBranchName)
    if !gitUtil.RefExists(rootAbsPath, fullBranchRef) {
        return fmt.Errorf("target branch '%s' does not exist", fullBranchRef)
    }
    
    // 4. Checkout target
    shortBranchName := strings.TrimPrefix(fullBranchRef, "refs/heads/")
    if err := gitUtil.Checkout(rootAbsPath, shortBranchName); err != nil {
        return fmt.Errorf("failed to checkout target branch: %w", err)
    }

    // 5. Cleanup empty directories
    if !keepEmptyDirs {
        fileUtil.CleanEmptyDirsRecursively(rootAbsPath)
    }
    
    return nil
}
```

---

### Pattern 3: Scoped Operations with Validation

Used by: `add`, `commit`

**Flow:**
```
1. Validate Environment
   ├─ Is Git repo?
   └─ On repo branch? (not main, not system)

2. Extract Context from Branch
   └─ Parse gitgroove/repos/<N>/branches/<b> → N

3. Load Metadata (No Checkout)
   ├─ Resolve gitgroove/system → tip
   └─ Read .gg/repos/* from tip

4. Validate Scope
   ├─ Files within repo?
   ├─ No .gg/* files?
   └─ No nested repo files?

5. Execute Git Operation
   └─ git add / git commit
```

**Key Design Point:** No system branch checkout needed for validation.


### Pattern 3: Scoped Operations with Validation

Used by: `add`, `commit`

**Flow:**
```
1. Validate Environment
   ├─ Is Git repo?
   └─ On repo branch? (not main, not system)

2. Extract Context from Branch
   └─ Parse gitgroove/repos/<N>/branches/<b> → N

3. Load Metadata (No Checkout)
   ├─ Resolve gitgroove/system → tip
   └─ Read .gg/repos/* from tip

4. Validate Scope
   ├─ Files within repo?
   ├─ No .gg/* files?
   └─ No nested repo files?

5. Execute Git Operation
   └─ git add / git commit
```

**Key Design Point:** No system branch checkout needed for validation.

---

## Concurrency and Atomicity

GitGrove is designed to handle multiple users working on the same repository simultaneously.

### 1. Metadata Concurrency

**Problem:** Two users register different repositories at the same time.
**Solution:** Optimistic Locking via Git's `update-ref`.

- Both users read `gitgroove/system` at commit `A`.
- User 1 creates commit `B` (parent `A`) and updates ref `A` -> `B`. Success.
- User 2 creates commit `C` (parent `A`) and tries to update ref `A` -> `C`.
- Git rejects the update because current ref is `B`, not `A`.
- User 2 receives an error: "Concurrent modification detected. Please retry."

### 2. File System Concurrency

**Problem:** User switches branches while another process is reading files.
**Solution:**
- Git's index lock prevents simultaneous index updates.
- GitGrove relies on Git's internal locking mechanisms for all working tree operations.

---

## Security and Validation

GitGrove enforces strict boundaries to prevent accidental data corruption or leakage.

### 1. Scope Validation

**Rule:** You cannot commit files that belong to another repository.

**Mechanism:**
- `gg add` checks every file path against the current repository's path.
- It also checks for `.gitgroverepo` markers in parent directories to ensure you aren't committing from a nested repository context unintentionally.

### 2. Metadata Integrity

**Rule:** Metadata must always be valid.

**Mechanism:**
- `register` and `link` validate the entire graph before committing.
- Cycle detection prevents infinite loops in the hierarchy.
- Path collision detection prevents overlapping repositories.

### 3. Clean State Enforcement

**Rule:** Major operations (`switch`, `register`) require a clean working tree.

**Mechanism:**
- Prevents uncommitted changes from being overwritten or lost during context switches.
- Forces users to commit or stash changes before performing structural operations.

---

## Performance Architecture

GitGrove is optimized for large monorepos.

### 1. Lazy Loading

- Metadata is only loaded when needed.
- `gg add` and `gg commit` read metadata directly from the object database without checking out the system branch, making them fast.

### 2. Sparse Operations

- `gg switch` uses Git's sparse-checkout mechanisms (simulated via tree manipulation) to only materialize relevant files.
- This keeps the working directory small and operations fast, even in massive repos.

### 3. Efficient Graph Algorithms

- Cycle detection and hierarchy traversal use efficient DFS/BFS algorithms.
- Metadata structure allows for O(1) lookup of repository paths.

---

## Extensibility and Modularity

### 1. Pluggable Architecture

The core logic is separated from the CLI, allowing for future integrations (e.g., IDE plugins, GUIs).

### 2. Versioned Metadata

The metadata structure is simple but extensible. Future versions can add new file types to `.gg/repos/` without breaking existing readers (which just ignore unknown files).

### 3. Hook System (Future)

The architecture allows for pre/post-operation hooks, enabling custom workflows like:
- Running tests before `gg commit`.
- Auto-generating documentation after `gg register`.
