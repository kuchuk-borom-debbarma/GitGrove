package registerrepo

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gg-test-repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "you@example.com")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Your Name")
	cmd.Dir = dir
	cmd.Run()

	// Initialize Grove
	if err := initialize.Initialize(dir); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to initialize grove: %v", err)
	}

	return dir
}

func TestRegisterRepo(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer os.RemoveAll(repoPath)

	// Create a dummy service directory and file
	servicePath := filepath.Join(repoPath, "backend", "serviceA")
	if err := os.MkdirAll(servicePath, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(servicePath, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Commit the file so git knows about it
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "Add serviceA")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Define repo to register
	newRepo := model.GGRepo{
		Name: "service-a",
		Path: "backend/serviceA",
	}

	// Register Repo
	if err := RegisterRepo([]model.GGRepo{newRepo}, repoPath); err != nil {
		t.Fatalf("RegisterRepo failed: %v", err)
	}

	// Verify gg.json
	configPath := filepath.Join(repoPath, ".gg", "gg.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read gg.json: %v", err)
	}

	var config groveUtil.GGConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse gg.json: %v", err)
	}

	if _, exists := config.Repositories["service-a"]; !exists {
		t.Errorf("Repo 'service-a' not found in gg.json")
	}

	// Verify branch creation
	cmd = exec.Command("git", "branch", "--list", "gg/service-a")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	if string(output) == "" {
		t.Errorf("Branch 'gg/service-a' was not created")
	}
}
