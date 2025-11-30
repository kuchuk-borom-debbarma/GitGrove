package grove_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

func TestMove(t *testing.T) {
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

	execGit(t, temp, "checkout", defaultBranch)

	// Create physical directories
	backendDir := filepath.Join(temp, "services/backend")
	if err := os.MkdirAll(backendDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Register repo
	if err := grove.Register(temp, map[string]string{"backend": "services/backend"}); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Note: Markers are now automatically tracked in git via plumbing API
	// No need to manually commit them

	// Move backend to services/backend
	newPath := "services/backend"
	if err := grove.Move(temp, "backend", newPath); err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	// Verify physical move
	if _, err := os.Stat(filepath.Join(temp, newPath)); err != nil {
		t.Errorf("expected directory at %s, got error: %v", newPath, err)
	}
	if _, err := os.Stat(filepath.Join(temp, "backend")); err == nil {
		t.Errorf("expected directory at backend to be gone")
	}

	// Verify metadata
	pathContent, err := gitUtil.ShowFile(temp, "gitgroove/internal", ".gg/repos/backend/path")
	if err != nil {
		t.Fatalf("failed to read metadata: %v", err)
	}
	if strings.TrimSpace(pathContent) != newPath {
		t.Errorf("expected path '%s', got '%s'", newPath, pathContent)
	}

	// Verify repo branch still exists and is accessible
	// It should be unchanged
	branchRef := "refs/heads/gitgroove/repos/backend/branches/main"
	if exists, _ := gitUtil.HasBranch(temp, branchRef); !exists {
		t.Errorf("expected repo branch %s to exist", branchRef)
	}
}
