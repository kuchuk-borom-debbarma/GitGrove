package grove_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

func TestRegister(t *testing.T) {
	// Setup
	temp := t.TempDir()
	createDummyProject(t, temp)
	execGit(t, temp, "init")
	execGit(t, temp, "add", ".")
	execGit(t, temp, "commit", "-m", "initial commit")

	if err := grove.Init(temp); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create directories for repos
	os.MkdirAll(filepath.Join(temp, "backend"), 0755)
	os.MkdirAll(filepath.Join(temp, "frontend"), 0755)

	// Test Case 1: Successful Registration
	repos := map[string]string{
		"backend":  "backend",
		"frontend": "frontend",
	}
	if err := grove.Register(temp, repos); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify .gg/repos content
	// Check backend repo
	backendPathContent, err := gitUtil.ShowFile(temp, "gitgroove/system", ".gg/repos/backend/path")
	if err != nil {
		t.Fatalf("failed to read backend path: %v", err)
	}
	if strings.TrimSpace(backendPathContent) != "backend" {
		t.Errorf("expected backend path 'backend', got '%s'", backendPathContent)
	}

	// Check frontend repo
	frontendPathContent, err := gitUtil.ShowFile(temp, "gitgroove/system", ".gg/repos/frontend/path")
	if err != nil {
		t.Fatalf("failed to read frontend path: %v", err)
	}
	if strings.TrimSpace(frontendPathContent) != "frontend" {
		t.Errorf("expected frontend path 'frontend', got '%s'", frontendPathContent)
	}

	// Test Case 2: Duplicate Name
	err = grove.Register(temp, map[string]string{"backend": "other"})
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}

	// Test Case 3: Duplicate Path
	os.MkdirAll(filepath.Join(temp, "other"), 0755)
	err = grove.Register(temp, map[string]string{"other": "backend"})
	if err == nil {
		t.Fatal("expected error for duplicate path, got nil")
	}

	// Test Case 4: Non-existent Path
	err = grove.Register(temp, map[string]string{"ghost": "ghost"})
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}

	// Test Case 5: Nested .git
	os.MkdirAll(filepath.Join(temp, "nested"), 0755)
	os.MkdirAll(filepath.Join(temp, "nested", ".git"), 0755)
	err = grove.Register(temp, map[string]string{"nested": "nested"})
	if err == nil {
		t.Fatal("expected error for nested .git, got nil")
	}
}
