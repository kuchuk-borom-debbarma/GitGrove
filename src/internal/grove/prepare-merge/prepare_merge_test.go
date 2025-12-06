package preparemerge

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	registerrepo "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/register-repo"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gg-test-pm")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	// Set default branch to main
	cmd.Args = append(cmd.Args, "--initial-branch=main")
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run()
	// Allow subtree merge of unrelated histories if needed (though here it shouldn't be unrelated)
	// But actually subtree merge works fine.

	// Initialize Grove
	if err := initialize.Initialize(dir, false); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to initialize grove: %v", err)
	}

	return dir
}

func TestPrepareMerge_FromOrphanBranch(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer os.RemoveAll(repoPath)

	// 1. Create a dummy service directory and file
	servicePath := filepath.Join(repoPath, "backend", "serviceA")
	if err := os.MkdirAll(servicePath, 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}
	mainGoPath := filepath.Join(servicePath, "main.go")
	if err := os.WriteFile(mainGoPath, []byte("package main\n\nfunc main() {}"), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Commit on main
	if err := gitUtil.Commit(repoPath, []string{"."}, "Add serviceA scaffold"); err != nil {
		t.Fatalf("Failed to commit scaffold: %v", err)
	}

	// 2. Register Repo
	newRepo := model.GGRepo{
		Name: "service-a",
		Path: "backend/serviceA",
	}
	if err := registerrepo.RegisterRepo([]model.GGRepo{newRepo}, repoPath); err != nil {
		t.Fatalf("RegisterRepo failed: %v", err)
	}

	// Commit gg.json so we can switch branches cleanly
	if err := gitUtil.Commit(repoPath, []string{".gg/gg.json"}, "Register service-a"); err != nil {
		t.Fatalf("Failed to commit gg.json: %v", err)
	}

	// 3. Checkout Orphan Branch: gg/main/service-a
	orphanBranch := "gg/main/service-a"
	if err := gitUtil.Checkout(repoPath, orphanBranch); err != nil {
		t.Fatalf("Failed to checkout orphan branch: %v", err)
	}

	// 4. Make changes in Orphan Branch
	// Note: In orphan branch, files are at root.
	// So "backend/serviceA/main.go" becomes "main.go".
	orphanMainGo := filepath.Join(repoPath, "main.go")
	newContent := "package main\n\nfunc main() { fmt.Println(\"Updated in Orphan\") }"
	if err := os.WriteFile(orphanMainGo, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to update file in orphan: %v", err)
	}
	// Add .gg/trunk to ensure it gets removed
	if err := os.MkdirAll(filepath.Join(repoPath, ".gg"), 0755); err != nil {
		t.Fatalf("Failed to create .gg dir in orphan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, ".gg", "trunk"), []byte("should be removed"), 0644); err != nil {
		t.Fatalf("Failed to create .gg/trunk in orphan: %v", err)
	}

	if err := gitUtil.Commit(repoPath, []string{"main.go", ".gg/trunk"}, "Update main.go and add .gg/trunk in orphan"); err != nil {
		t.Fatalf("Failed to commit in orphan: %v", err)
	}

	// 5. Run PrepareMerge (detect context)
	// We are on "gg/main/service-a".
	// The function should detect trunk="main" and repo="service-a".
	if err := PrepareMerge(repoPath, ""); err != nil {
		t.Fatalf("PrepareMerge failed: %v", err)
	}

	// 6. Verify
	// Check current branch is gg/merge-prep/service-a/...
	currentBranch, err := gitUtil.CurrentBranch(repoPath)
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if !strings.HasPrefix(currentBranch, "gg/merge-prep/service-a/") {
		t.Errorf("Expected current branch to start with gg/merge-prep/service-a/, got %s", currentBranch)
	}

	// Check content of backend/serviceA/main.go in this new branch
	// It should contain the update.
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read project file: %v", err)
	}
	if string(content) != newContent {
		t.Errorf("Content mismatch. Expected:\n%s\nGot:\n%s", newContent, string(content))
	}

	// Verify .gg/trunk is NOT present
	trunkFile := filepath.Join(repoPath, ".gg", "trunk")
	if _, err := os.Stat(trunkFile); !os.IsNotExist(err) {
		t.Errorf("Expected .gg/trunk to be removed, but it exists")
	}
}
