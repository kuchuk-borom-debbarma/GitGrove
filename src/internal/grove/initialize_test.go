package grove_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove"
)

func TestInitialize(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gitgrove-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a git repo in the temp dir
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commit to work
	cmd = exec.Command("git", "config", "user.email", "you@example.com")
	cmd.Dir = tempDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Your Name")
	cmd.Dir = tempDir
	cmd.Run()

	// Run Initialize
	if err := grove.Initialize(tempDir); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify .gg/gg.json exists
	configPath := filepath.Join(tempDir, ".gg", "gg.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("gg.json was not created")
	}

	// Verify commit was made
	cmd = exec.Command("git", "log", "-1", "--pretty=%B")
	cmd.Dir = tempDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get git log: %v", err)
	}
	if string(output) != "Initialize GitGrove\n" && string(output) != "Initialize GitGrove\n\n" {
		t.Errorf("Unexpected commit message: %q", string(output))
	}
}
