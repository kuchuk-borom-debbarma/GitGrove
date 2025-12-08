# GitGrove (GG) -> The Isolated-History Monorepo Manager

GitGrove is a powerful tool designed to manage multiple logical repositories within a single Git Monorepo. It enforces strict history isolation for sub-projects while maintaining a unified integration trunk.

By using GitGrove, developers can enjoy the "feel" of working in a small, isolated repository (via orphan branches) while the codebase physically resides in a monorepo.

## 🚀 Key Features

*   **Modular Monorepo**: Manage distinct projects (`services/a`, `lib/b`) as if they were separate repos.
*   **The TUI**: A terminal user interface to manage the complex states easily.
*   **History Isolation**: Work on orphan branches that contain *only* the files for your specific project.
*   **Context Aware Commits**: Automatically prefixes commit messages (e.g., `[ServiceA] Fix bug`) when you commit changes scoped to a single repo.
*   **Atomic Commit Enforcement**: Prevents "spaghetti history" by blocking commits that touch multiple registered repositories simultaneously.
*   **Safe Integration**: Automates the complex process of merging isolated history back into the main monorepo trunk.

---

## Installation

You can install GitGrove by downloading a pre-built release or by building from source.

### Option 1: Pre-built Release (Recommended)
This is the easiest way to get started.

1. **Download** the latest release `.zip` for your OS and architecture (e.g., `GitGrove_darwin_arm64.zip` for Apple Silicon Mac).
2. **Unzip** the archive.
3. **Run the Setup Script** inside the extracted folder to add `gg` to your PATH:
   - **MacOS/Linux**: Open terminal in the folder and run `./setup.sh`
   - **Windows**: Right-click `setup.ps1` and "Run with PowerShell" (or run `.\setup.ps1` in a terminal).
4. **Restart your terminal** to apply the changes.

### Option 2: Build from Source
Use this method if you want to develop GitGrove or validte the latest source code.
**Prerequisite**: [Go](https://go.dev/doc/install) 1.22 or higher.

1. **Clone** the repository.
2. **Run the Installer** for your OS from the project root:
   - **MacOS**: `./scripts/install_mac.sh`
   - **Linux**: `./scripts/install_linux.sh`
   - **Windows**: `.\scripts\install_windows.ps1`
3. **Restart your terminal**.

## 📖 How to Use

GitGrove can be controlled via its **Terminal User Interface (TUI)** or via specific **CLI commands**.

### 1. Initialization
Turn your current git repository into a GitGrove workspace.

**Using CLI:**
```bash
gg init
# OR strictly enforce atomic commits from the start
gg init --atomic
```

**Using TUI:**
Simply run `gg` in your repo folder. If it's not initialized, you will be prompted to set it up.

### 2. Registering a Repository
Slice a sub-folder into its own logical repository.

*Currently, registration is done via the TUI.*

1.  Run `gg`.
2.  Select **"Register Repo"**.
3.  Enter the name (e.g., `service-a`) and the relative path (e.g., `backend/service-a`).
4.  GitGrove will:
    *   Update configuration.
    *   Create an **orphan branch** (`gg/main/service-a`) containing only that folder's history.

### 3. The Workflow (Development)

To work on a specific repository using its isolated history:

1.  Checkout the orphan branch:
    ```bash
    git checkout gg/main/service-a
    ```
    *(Or check the TUI for branch names)*

2.  **Work as normal!** You will see only the files for `service-a` at the root level.
3.  **Commit**:
    ```bash
    git add .
    git commit -m "My new feature"
    ```
    *   **Context Aware**: If configured, `gg` will automatically prepend `[service-a]` to your message.
    *   **Atomic Check**: If you somehow staged files from outside the scope (unlikely in orphan branch, but possible in Trunk), `gg` will block the commit.

### 4. Merging Back (Integration)

When your feature is ready to be merged back into the main trunk:

**Using CLI:**
```bash
# Run this from your orphan branch
gg prepare-merge
```

**Using TUI:**
1.  Run `gg`.
2.  Select **"Prepare Merge"**.

**What happens?**
GitGrove will:
1.  Switch to the **trunk branch** (e.g., `main`).
2.  Create a timestamped integration branch (e.g., `gg/merge-prep/service-a/12345`).
3.  Merge your orphan branch changes back into the correct nested file structure.
4.  You can now open a **Pull Request** from this branch to your trunk.

---

## 🧠 Architecture Overview

*   **The Trunk**: Your `main` branch. Contains everything.
*   **The Split**: Isolated branches created via `git subtree`.
*   **The Guard**: Git hooks (`pre-commit`) that ensure you don't accidental mix histories.

For deep details, see [docs/architecture.md](docs/architecture.md).
