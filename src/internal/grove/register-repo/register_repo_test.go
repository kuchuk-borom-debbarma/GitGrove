package registerrepo

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if err := initialize.Initialize(dir, false); err != nil {
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

	// Get current branch for verification
	currentBranch, err := exec.Command("git", "-C", repoPath, "branch", "--show-current").Output()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	currentBranchStr := string(currentBranch)
	// Remove newline
	if len(currentBranchStr) > 0 {
		currentBranchStr = currentBranchStr[:len(currentBranchStr)-1]
	}

	// Verify branch creation
	expectedBranch := "gg/service-a"
	if len(currentBranchStr) > 0 {
		expectedBranch = "gg/" + currentBranchStr + "/service-a"
	}

	cmd = exec.Command("git", "branch", "--list", expectedBranch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	if string(output) == "" {
		t.Errorf("Branch '%s' was not created", expectedBranch)
	}
}

func TestRegisterRepo_PathValidation(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer os.RemoveAll(repoPath)

	// Attempt to register a path outside the repo (e.g. ../outside)
	// Since we are mocking, we just pass the path string.
	// But RegisterRepo checks relative to repoPath.
	// So we pass "../outside" as path.

	newRepo := model.GGRepo{
		Name: "outside-repo",
		Path: "../outside",
	}

	err := RegisterRepo([]model.GGRepo{newRepo}, repoPath)
	if err == nil {
		t.Fatal("Expected RegisterRepo to fail for path '../outside', but it succeeded")
	}

	expectedError := "must be within repository root"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%v'", expectedError, err)
	}
}
