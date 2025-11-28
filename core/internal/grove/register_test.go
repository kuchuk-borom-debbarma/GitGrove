package grove_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
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

	// Verify repos.jsonl content
	content, err := gitUtil.ShowFile(temp, "gitgroove/system", ".gg/repos.jsonl")
	if err != nil {
		t.Fatalf("failed to read repos.jsonl: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines in repos.jsonl, got %d", len(lines))
	}

	var r1, r2 model.Repo
	json.Unmarshal([]byte(lines[0]), &r1)
	json.Unmarshal([]byte(lines[1]), &r2)

	// Order isn't guaranteed in map iteration, so check existence
	registered := map[string]string{
		r1.Name: r1.Path,
		r2.Name: r2.Path,
	}

	if registered["backend"] != "backend" || registered["frontend"] != "frontend" {
		t.Fatalf("unexpected registration result: %v", registered)
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
