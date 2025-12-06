# GitGrove (GG): Architecture & Workflow
**The Isolated-History Monorepo Manager**

GitGrove (GG) is a client-side tool designed to manage multiple logical repositories within a single Git Monorepo. It achieves strict history isolation for sub-projects while maintaining a unified integration trunk, using Git's native Subtree mechanics under the hood.

---

## 1. Initialization (The Trunk)
The process begins by establishing the "Main" branch as the source of truth (The Trunk).

*   **Command:** `gg init`
*   **Action:**
    *   Initializes the `gg` environment.
    *   Creates a `gg.json` metadata file to store repo paths and relationships.
    *   Commits this configuration to `main`.

## 2. Registration (The Split)
When a folder is registered as a "repo," GG creates a parallel history for it.

*   **Command:** `gg register serviceA ./backend/services/serviceA`
*   **Action:**
    *   **Orphan Creation:** GG scans the history of the target folder.
    *   **Path Translation:** It uses `git subtree split` to create a new orphan branch (e.g., `gg/main/serviceA`).
    *   **Root Projection:** Inside this new branch, the files are moved to the Root Directory.
        *   *Trunk View:* `./backend/services/serviceA/main.go`
        *   *Orphan View:* `./main.go`

> **Note:** Nested directories cannot be registered as repositories at this time. Strict validation prevents registering a repository inside another registered repository.

## 3. The Development Loop (Independent Tracking)
Developers work inside these isolated orphan branches. This provides the "Per-Repo Commit Tracking" feature.

*   **Command:** `gg checkout serviceA`
*   **Experience:**
    *   The developer sees the project at the root level.
    *   They commit changes normally (`git add`, `git commit`).
    *   **Result:** The commit history on this branch is pure. It contains only changes relevant to `serviceA`, with no noise from the rest of the Monorepo.

## 4. The Safety Guard (The Hook)
To ensure the Monorepo history remains compatible with the Orphan history, GG enforces a strict workflow.

*   **Mechanism:** A Pre-Commit Hook.
*   **Logic:**
    *   If a user is on `main` and attempts to manually modify files inside a registered path (e.g., `./backend/services/serviceA`), GG blocks the commit.
*   **Why?** This forces users to make changes in the Orphan branch first. "Ghost Changes" in Main are prevented, ensuring the Orphan branch is always the source of truth, making merging easy and conflict-free.

## 5. Integration (Prepare for Merge)
When the work in the isolated branch is ready to be integrated, we prepare a branch for merging into the trunk.

*   **Command:** `gg prepare-merge [repoName]`
*   **Goal:** Create a bridge between the isolated orphan branch and the main trunk.
*   **Action:**
    *   **Context Detection:** Automatically detects if run from an Orphan Branch.
    *   Creates a new branch off the trunk.
    *   Brings in the changes from the registered repository's orphaned branch.
    *   Applies path translation to ensure files land in their correct nested location (e.g., `./main.go` -> `./backend/services/serviceA/main.go`).
*   **Result:** A standard feature branch that contains the updates.
*   **Next Step:** The user can now merge this branch into the trunk using standard Git workflows (Pull Request or manual merge).

### Summary of Benefits
1.  **True Isolation:** Developers work with root-level files, ignoring the complexity of the massive Monorepo folder structure.
2.  **Clean History:** `git log` on the orphan branch shows a perfect, linear history of just that project.
3.  **Rock-Solid Merging:** By using `git subtree` logic and blocking direct edits to Main, we guarantee that the "Main" tree and "Orphan" tree always share a compatible structure.

### Current Limitations
1.  **Cross-Repo Dependencies:** (Status: Out of Scope for v1)
    *   The system designs for *Total Isolation*. If `serviceA` depends on `../backend/utils`, those files are not present in the orphan branch (since `gg checkout serviceA` only shows `serviceA` files).
    *   Repositories must be self-contained in v1. Support for "Composite Workspaces" is reserved for future iterations.

---

# GitGrove: Context-Aware Commit (GGC)
**The Integrated Discipline Layer**

"GGC" refers to the Context-Aware Commit features integrated into GitGrove. It enforces strict commit hygiene in a Monorepo. While it is fully integrated into the standard `gg` workflow, it can be conceptually thought of as a lightweight discipline layer.

**If you plan on just having commit context then use GGC. This moves away from complex branching and focuses purely on Workflow Discipline within a standard Monorepo.**

It ensures that although your code lives in one giant folder, your commit history looks like it came from isolated micro-services.

## 1. Setup: Registration
You tell GGC which folders represent logical "repositories."

*   **Command:** `gg register serviceA ./backend/services/serviceA`
*   **State:** GGC updates `gg.json` to map that specific path to the tag `[serviceA]`.

## 2. The Core Feature: Smart Tagging & Validation
When a developer runs `gg commit`, the tool analyzes the staged files before Git creates the commit.

### Scenario A: The Clean Commit
*   **User Action:** Modifies 3 files inside `./backend/services/serviceA`.
*   **Command:** `gg commit -m "Fix login bug"`
*   **GGC Logic:**
    1.  Detects all files belong to the registered `serviceA` path.
    2.  **Auto-Tag:** Modifies the message to `[serviceA] Fix login bug`.
*   **Result:** The history is automatically categorized without user effort.

### Scenario B: The Mixed Commit (The Block)
*   **User Action:** Modifies files in `serviceA` AND `frontend/ui`.
*   **Command:** `gg commit -m "Update login"`
*   **GGC Logic:**
    1.  Detects paths from two distinct registered repos.
    2.  **BLOCK:** 🛑 Error: "Atomic Commit Violation. You are attempting to commit to 'serviceA' and 'frontend' simultaneously. Please split your commits."
*   **Result:** Prevents "Spaghetti History" where unrelated features are entangled in one commit hash.

## 3. Advanced Feature: Repo Bounding (Active Scope)
This allows a developer to "Lock In" their focus, ensuring they don't accidentally touch other parts of the codebase.

*   **Command:** `gg scope serviceA`
*   **Effect:**
    *   GGC enters "Strict Mode" for `serviceA`.
    *   If you try to commit anything outside of `./backend/services/serviceA`, GGC throws an error immediately—even if you haven't touched ServiceA files.
*   **Use Case:** Ideal for specific tasks where a developer wants to ensure zero accidental pollution of other modules.

### Summary: Why use GGC?
This is the "Low Tech, High Value" solution.

1.  **Zero Risk:** No complex Git tree rewriting or hidden branches. It's just a standard repo.
2.  **Perfect Logs:** Running `git log --grep="[serviceA]"` gives you a perfect, linear history of that specific service.
3.  **Standalone:** You can use this logic without adopting the full GitGrove architecture. It is essentially a super-powered Git Hook.
