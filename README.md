# GitGrove

### *Multi-Repository Version Control on Top of Git*

GitGrove is a powerful tool that transforms a single Git repository into a **structured forest of independent repositories**. It allows you to manage complex monorepos with the simplicity of a single Git backend, while giving each component its own isolated history, lifecycle, and workspace.

To Git, everything remains a normal repository. To you, your project becomes a hierarchy of focused workspaces.

---

## Why GitGrove?

Managing large monorepos often leads to:
*   **Noise**: Every `git status` shows irrelevant changes from other teams.
*   **Complexity**: Submodules are painful; subtrees are complex.
*   **Loss of Context**: It's hard to tell where one service ends and another begins.

**GitGrove solves this by:**
*   **Virtualizing Repositories**: Treat any subdirectory as a standalone repo.
*   **Flattening Views**: When you work on `backend`, your root directory *becomes* `backend`. No more `cd backend/src/...`.
*   **Preserving History**: Each virtual repo has its own commit history, isolated from the rest.
*   **Staying 100% Git Compatible**: Under the hood, it's just Git. You can push, pull, and merge as usual.

---

## Who Is It For?

*   **Microservices Teams**: Manage 50 services in one repo without the chaos.
*   **Full-Stack Developers**: Keep frontend and backend coupled in versioning but decoupled in development.
*   **Modular Projects**: Library authors managing multiple packages.

---

## Installation

Currently, GitGrove is built from source.

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/kuchuk-borom-debbarma/GitGrove.git
    cd GitGrove
    ```

2.  **Build the CLI**:
    ```bash
    cd cli
    go build -o ../gitgrove ./cmd/main.go
    ```

3.  **Add to PATH** (Optional):
    ```bash
    export PATH=$PATH:$(pwd)/..
    ```

---

## User Guide

### 1. Initialization
Turn any Git repository into a GitGrove project.

```bash
# Inside your git repo
gitgrove init
```
This sets up the internal metadata tracking on a hidden `gitgroove/system` branch.

### 2. Registering Repositories
Tell GitGrove which folders should be treated as repositories.

```bash
# Register 'backend' located at ./backend
gitgrove register --name backend --path backend

# Register 'frontend' located at ./web/ui
gitgrove register --name frontend --path web/ui
```

### 3. Creating Hierarchy
Link repositories to define structure. This is useful for navigation.

```bash
# 'service-a' is a child of 'backend'
gitgrove link --child service-a --parent backend
```

### 4. Navigation & Workflow
This is where GitGrove shines. You can move between repositories seamlessly.

#### Interactive Mode
The easiest way to explore.
```bash
gitgrove interactive
```
This launches a menu where you can browse your repo hierarchy and perform actions.

#### Command Line Navigation
*   **Check Status**: See where you are and the full hierarchy.
    ```bash
    gitgrove info
    ```
    *Output:*
    ```text
    ...
    └── backend
        ├── service-a * [CURRENT]
        └── service-b
    ```

*   **Switch Context**: Jump to a specific repo.
    ```bash
    gitgrove switch backend main
    ```
    *Your working directory now contains ONLY the files from `backend`.*

*   **Move Up/Down**: Traverse the tree like a filesystem.
    ```bash
    gitgrove ls            # List child repositories
    gitgrove cd service-b  # Go to service-b
    gitgrove cd ..         # Go up to parent
    ```

### 5. Daily Work
Work exactly as you would in Git, but with GitGrove's safety wrappers.

*   **Stage Files**:
    ```bash
    gitgrove add .
    ```
    *Only stages files belonging to the current virtual repo.*

*   **Commit**:
    ```bash
    gitgrove commit "Fix login bug"
    ```
    *Ensures you are committing to the correct repository context.*

---

## Technical Details

Want to know how the magic works? Check out our [Technical Internals Documentation](docs/internals.md).

---

## License

MIT
