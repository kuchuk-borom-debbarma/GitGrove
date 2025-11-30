package grove_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

func TestSwitch(t *testing.T) {
	// Setup temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "gitgroove-switch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repo
	if err := gitUtil.Init(tmpDir); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create .gg directory to simulate GitGroove repo
	if err := os.Mkdir(filepath.Join(tmpDir, ".gg"), 0755); err != nil {
		t.Fatalf("failed to create .gg dir: %v", err)
	}

	// Create initial commit
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Repo"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	if err := gitUtil.StagePath(tmpDir, "."); err != nil {
		t.Fatalf("failed to stage files: %v", err)
	}
	if err := gitUtil.Commit(tmpDir, "Initial commit"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Ensure we are on 'main'
	if _, err := gitUtil.RunGit(tmpDir, "branch", "-m", "main"); err != nil {
		t.Fatalf("failed to rename branch to main: %v", err)
	}

	// Create gitgroove/internal branch
	if err := gitUtil.CreateBranch(tmpDir, "gitgroove/internal", "HEAD"); err != nil {
		t.Fatalf("failed to create system branch: %v", err)
	}

	// Register repos
	repos := map[string]string{
		"root":     ".",
		"backend":  "backend",
		"frontend": "frontend",
	}
	// Create directories for repos
	if err := os.Mkdir(filepath.Join(tmpDir, "backend"), 0755); err != nil {
		t.Fatalf("failed to create backend dir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "frontend"), 0755); err != nil {
		t.Fatalf("failed to create frontend dir: %v", err)
	}

	// Switch to gitgroove/internal to register
	if err := gitUtil.Checkout(tmpDir, "gitgroove/internal"); err != nil {
		t.Fatalf("failed to checkout system branch: %v", err)
	}

	if err := grove.Register(tmpDir, repos); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Register now automatically commits .gitgroverepo markers when on system branch.
	// So we don't need to manually commit them.

	// Link repos
	relationships := map[string]string{
		"backend":  "root",
		"frontend": "root",
	}
	if err := grove.Link(tmpDir, relationships); err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// Switch back to main to simulate user working
	if err := gitUtil.Checkout(tmpDir, "main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	// Test Switch
	t.Run("Switch to backend main", func(t *testing.T) {
		if err := grove.Switch(tmpDir, "backend", ""); err != nil {
			t.Fatalf("Switch failed: %v", err)
		}

		// Verify HEAD
		head, err := gitUtil.GetCurrentBranch(tmpDir)
		if err != nil {
			t.Fatalf("failed to get current branch: %v", err)
		}
		expected := "gitgroove/repos/backend/branches/main"
		if head != expected {
			t.Errorf("expected HEAD to be %s, got %s", expected, head)
		}
	})

	t.Run("Switch to frontend main", func(t *testing.T) {
		// Switch requires clean state.
		// We are currently on backend branch. It should be clean.

		if err := grove.Switch(tmpDir, "frontend", "main"); err != nil {
			t.Fatalf("Switch failed: %v", err)
		}

		head, err := gitUtil.GetCurrentBranch(tmpDir)
		if err != nil {
			t.Fatalf("failed to get current branch: %v", err)
		}
		expected := "gitgroove/repos/frontend/branches/main"
		if head != expected {
			t.Errorf("expected HEAD to be %s, got %s", expected, head)
		}
	})

	t.Run("Switch to non-existent repo", func(t *testing.T) {
		if err := grove.Switch(tmpDir, "non-existent", ""); err == nil {
			t.Error("expected error for non-existent repo, got nil")
		}
	})
}
