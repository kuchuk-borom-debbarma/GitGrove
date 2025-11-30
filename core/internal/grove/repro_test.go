package grove_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
)

func TestRegisterLeavesUntrackedMarkers(t *testing.T) {
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

	// Register multiple repos
	repos := map[string]string{
		"backend":  "backend",
		"frontend": "frontend",
	}
	if err := grove.Register(temp, repos); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Check for untracked files
	// We use git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = temp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}

	// If the bug is fixed, we expect NO untracked files (specifically no .gitgroverepo)
	output := string(out)
	if contains(output, "?? backend/.gitgroverepo") {
		t.Errorf("Expected backend/.gitgroverepo to be tracked, but got untracked:\n%s", output)
	} else {
		t.Log("Verified fix: backend/.gitgroverepo is tracked")
	}

	if contains(output, "?? frontend/.gitgroverepo") {
		t.Errorf("Expected frontend/.gitgroverepo to be tracked, but got untracked:\n%s", output)
	} else {
		t.Log("Verified fix: frontend/.gitgroverepo is tracked")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || contains(s[1:], substr))))
}

// Using strings.Contains would be better but need import.
// I'll add imports.

func createDummyProjectRepro(t *testing.T, root string) {
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# Demo"), 0644)
}

func execGitRepro(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
	}
}
