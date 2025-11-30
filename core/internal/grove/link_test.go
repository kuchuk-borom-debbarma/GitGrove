package grove_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

func TestLink(t *testing.T) {
	// Setup
	temp := t.TempDir()
	createDummyProject(t, temp)
	execGit(t, temp, "init")
	execGit(t, temp, "add", ".")
	execGit(t, temp, "commit", "-m", "initial commit")

	// Capture default branch
	defaultBranch := strings.TrimSpace(execGit(t, temp, "branch", "--show-current"))

	if err := grove.Init(temp); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Switch back to default branch because Init leaves us on gitgroove/internal
	// and Register/Link expect to run from a user branch.
	// If we stay on gitgroove/internal, Register's update-ref will make the index appear dirty.
	execGit(t, temp, "checkout", defaultBranch)

	// Create directories for repos
	os.MkdirAll(filepath.Join(temp, "services", "backend"), 0755)
	os.MkdirAll(filepath.Join(temp, "services", "frontend"), 0755)
	os.MkdirAll(filepath.Join(temp, "libs", "shared"), 0755)

	// Register repos
	repos := map[string]string{
		"backend":  "services/backend",
		"frontend": "services/frontend",
		"shared":   "libs/shared",
	}
	if err := grove.Register(temp, repos); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Register creates .gitgroverepo files which are untracked.
	// We need to commit them to have a clean state for Link.
	execGit(t, temp, "add", ".")
	execGit(t, temp, "commit", "-m", "Add .gitgroverepo markers")

	// Test Case 1: Successful Linking
	// Hierarchy:
	// shared -> backend
	// shared -> frontend
	links := map[string]string{
		"backend":  "shared",
		"frontend": "shared",
	}
	if err := grove.Link(temp, links); err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// Verify .gg/repos content
	// Check backend parent
	backendParent, err := gitUtil.ShowFile(temp, "gitgroove/internal", ".gg/repos/backend/parent")
	if err != nil {
		t.Fatalf("failed to read backend parent: %v", err)
	}
	if strings.TrimSpace(backendParent) != "shared" {
		t.Errorf("expected backend parent 'shared', got '%s'", backendParent)
	}

	// Check shared children
	// .gg/repos/shared/children/backend should exist
	_, err = gitUtil.ShowFile(temp, "gitgroove/internal", ".gg/repos/shared/children/backend")
	if err != nil {
		t.Errorf("expected child entry for backend in shared, got error: %v", err)
	}

	// Verify repo branches exist (created by Register)
	// backend: refs/heads/gitgroove/repos/backend/branches/main
	backendBranch := "refs/heads/gitgroove/repos/backend/branches/main"
	if exists, _ := gitUtil.HasBranch(temp, backendBranch); !exists {
		t.Errorf("expected branch %s to exist", backendBranch)
	}

	// frontend: refs/heads/gitgroove/repos/frontend/branches/main
	frontendBranch := "refs/heads/gitgroove/repos/frontend/branches/main"
	if exists, _ := gitUtil.HasBranch(temp, frontendBranch); !exists {
		t.Errorf("expected branch %s to exist", frontendBranch)
	}

	// shared: refs/heads/gitgroove/repos/shared/branches/main
	sharedBranch := "refs/heads/gitgroove/repos/shared/branches/main"
	if exists, _ := gitUtil.HasBranch(temp, sharedBranch); !exists {
		t.Errorf("expected branch %s to exist", sharedBranch)
	}

	// Test Case 2: Cycle Detection
	// Try to make backend -> shared (creating a cycle shared -> backend -> shared)
	cycleLinks := map[string]string{
		"shared": "backend",
	}
	if err := grove.Link(temp, cycleLinks); err == nil {
		t.Fatal("expected error for cycle, got nil")
	}

	// Test Case 3: Non-existent Parent
	invalidParentLinks := map[string]string{
		"backend": "ghost",
	}
	if err := grove.Link(temp, invalidParentLinks); err == nil {
		t.Fatal("expected error for non-existent parent, got nil")
	}

	// Test Case 4: Non-existent Child
	invalidChildLinks := map[string]string{
		"ghost": "shared",
	}
	if err := grove.Link(temp, invalidChildLinks); err == nil {
		t.Fatal("expected error for non-existent child, got nil")
	}

	// Test Case 5: Self-reference
	selfLinks := map[string]string{
		"backend": "backend",
	}
	if err := grove.Link(temp, selfLinks); err == nil {
		t.Fatal("expected error for self-reference, got nil")
	}

	// Test Case 6: Existing Parent
	// backend already has parent 'shared' from Test Case 1
	// Try to assign it another parent
	reparentLinks := map[string]string{
		"backend": "frontend",
	}
	if err := grove.Link(temp, reparentLinks); err == nil {
		t.Fatal("expected error for existing parent, got nil")
	}

	// Test Case 7: Dangling Repo
	// Delete backend directory
	os.RemoveAll(filepath.Join(temp, "services", "backend"))
	// Try to link something to backend or use backend
	// Note: Link checks child existence.
	// Let's try to link backend to something else (if we could reparent)
	// Or better, register a new repo, delete its dir, then try to link it.

	// But we can't register if dir doesn't exist.
	// So we register, then delete, then link.
	os.MkdirAll(filepath.Join(temp, "dangling"), 0755)
	grove.Register(temp, map[string]string{"dangling": "dangling"})
	os.RemoveAll(filepath.Join(temp, "dangling"))

	danglingLinks := map[string]string{
		"dangling": "shared",
	}
	if err := grove.Link(temp, danglingLinks); err == nil {
		t.Fatal("expected error for dangling repo, got nil")
	}
}
