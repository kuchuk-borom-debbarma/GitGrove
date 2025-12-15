# GitGrove (GG) -> The Isolated-History Monorepo Manager

GitGrove is a powerful tool designed to manage multiple logical repositories within a single Git Monorepo. It enforces strict history isolation for sub-projects while maintaining a unified integration trunk.

By using GitGrove, developers can enjoy the "feel" of working in a small, isolated repository (via orphan branches) while the codebase physically resides in a monorepo.

## 🚀 Key Features

*   **Modular Monorepo**: Manage distinct projects (`services/a`, `lib/b`) as if they were separate repos.
*   **The TUI**: A terminal user interface to manage the complex states easily.
*   **Isolate & Focus**: Work on a single folder as if it were a standalone repository.
*   **Reset to Trunk**: Safely hard-reset your isolated workspace to match the latest trunk state (discarding local changes).
*   **Context-Aware Commits**: Commits are automatically prefixed with the component name (e.g., `[service-a] feat: new API`).
*   **Atomic Commit Enforcement**: Prevents "spaghetti history" by blocking commits that touch multiple registered repositories simultaneously.
*   **Safe Integration**: Automates the complex process of merging isolated history back into the main monorepo trunk.

---

## Installation

You can install GitGrove by downloading a pre-built release or by building from source.

### Option 1: Pre-built Release (Recommended)
This is the easiest way to get started.

1. **Download** the latest release (`v1.1.2`) .zip for your OS from the Releases page.
2. **Extract** the archive.
3. **Add the binary to your PATH**:
   - **MacOS/Linux**: Move the `gg` binary to `/usr/local/bin` (or another directory in your PATH).
     ```bash
     sudo mv gg /usr/local/bin/
     ```
   - **Windows**: Move `gg.exe` to a folder in your PATH (e.g., `C:\Program Files\GitGrove`) and add that folder to your User PATH environment variable.
4. **Restart your terminal** to apply the changes.

### Option 2: Build from Source
Use this method if you want to develop GitGrove or validate the latest source code.
**Prerequisite**: [Go](https://go.dev/doc/install) 1.22 or higher.

1. **Clone** the repository:
   ```bash
   git clone https://github.com/kuchuk-borom-debbarma/GitGrove.git
   cd GitGrove
   ```
2. **Run the Build Script** for your OS:
   - **MacOS**: `./scripts/install_mac.sh`
   - **Linux**: `./scripts/install_linux.sh`
   - **Windows**: `.\scripts\install_windows.ps1`
3. **Follow the on-screen instructions** to add the built binary to your PATH.

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

**Using CLI:**
```bash
gg register <repo-name> <relative-path>
# Example: gg register service-a backend/service-a
```

**Using TUI:**
1.  Run `gg`.
2.  Select **"Register Repo"**.
3.  Enter the name (e.g., `service-a`) and the relative path (e.g., `backend/service-a`).

GitGrove will:
*   Update configuration (`gg.json`).
*   Create an **orphan branch** (`gg/<trunk>/<repoName>`) containing only that folder's history.

### 3. The Workflow (Development)

To work on a specific repository using its isolated history:

1.  **Checkout the orphan branch**:
    
    **Using CLI:**
    ```bash
    gg checkout <repo-name>
    ```

    **Using TUI:**
    Select **"Checkout Repo Branch"** and choose the repository.

2.  **Work as normal!** You will see only the files for `service-a` at the root level.
3.  **Commit**:
    ```bash
    git add .
    git commit -m "My new feature"
    ```
    *   **Context Aware**: If configured, `gg` will automatically prepend `[service-a]` to your message.
    *   **Sticky Context** (New): If you checkout a repo via the TUI, you can freely create feature branches (e.g., `git checkout -b feature/login`) and your commits will *still* be automatically prefixed.
    *   **Atomic Check**: If you somehow staged files from outside the scope (unlikely in orphan branch, but possible in Trunk), `gg` will block the commit.

### 4. Return to Trunk
When you are done, simply switch back to the main branch.

**Using CLI:**
```bash
gg trunk
```

**Using TUI:**
Select **"Return to Trunk"**.

### 5. Resetting to Trunk
If updates have been made to your component in the main branch (e.g., by other team members) and you want to start fresh:

**Using CLI:**
```bash
gg reset
```

**Using TUI:**
1.  Inside your orphan branch, select **"Reset to Trunk"**.
2.  Confirm the warning prompt.

GitGrove will **hard reset** your workspace to match the trunk's version of the component, discarding any local changes.

### 6. Merging Back (Integration)

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
