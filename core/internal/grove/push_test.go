package grove

import (
	"os"
	"path/filepath"
	"testing"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

func TestPush(t *testing.T) {
	// 1. Setup temp directory
	tempDir, err := os.MkdirTemp("", "gitgrove-push-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 2. Initialize Git repo
	if err := gitUtil.Init(tempDir); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}
	// Configure user for commits
	gitUtil.RunGit(tempDir, "config", "user.email", "test@example.com")
	gitUtil.RunGit(tempDir, "config", "user.name", "Test User")

	// Initial commit
	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	gitUtil.RunGit(tempDir, "add", ".")
	gitUtil.RunGit(tempDir, "commit", "-m", "Initial commit")

	// 3. Setup bare remote
	remoteDir, err := os.MkdirTemp("", "gitgrove-push-remote")
	if err != nil {
		t.Fatalf("Failed to create remote dir: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	if _, err := gitUtil.RunGit(remoteDir, "init", "--bare"); err != nil {
		t.Fatalf("Failed to init bare remote: %v", err)
	}

	// Add remote to local repo
	if _, err := gitUtil.RunGit(tempDir, "remote", "add", "origin", remoteDir); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	// 4. Initialize GitGrove
	if err := Init(tempDir); err != nil {
		t.Fatalf("Failed to init GitGrove: %v", err)
	}

	// Switch back to main
	gitUtil.Checkout(tempDir, "main")

	// 5. Register a repo
	backendPath := filepath.Join(tempDir, "backend")
	if err := os.MkdirAll(backendPath, 0755); err != nil {
		t.Fatalf("Failed to create backend dir: %v", err)
	}
	repos := map[string]string{"backend": "backend"}
	if err := Register(tempDir, repos); err != nil {
		t.Fatalf("Failed to register repo: %v", err)
	}

	// 6. Push
	// We need to be on a clean state. Register might have left staged changes?
	// Register stages marker files but doesn't commit them to user branch.
	// We should commit them.
	gitUtil.RunGit(tempDir, "add", ".")
	gitUtil.RunGit(tempDir, "commit", "-m", "Add backend")

	// Now call Push
	targets := []string{"backend"}
	if err := Push(tempDir, targets); err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// 7. Verify remote has the branch
	// The branch pushed should be 'main' (DefaultRepoBranch)
	// But wait, Register creates 'gitgroove/repos/backend/branches/main'.
	// Push pushes that to 'origin/main'.

	// Check if 'main' exists in remote
	if valid, _ := gitUtil.HasBranch(remoteDir, "main"); !valid {
		// Bare repos don't have local branches checked out, but they have refs/heads/main
		// HasBranch checks rev-parse --verify --quiet main
		// In a bare repo, we might need to check refs/heads/main explicitly or just main.
		// Let's check if we can fetch it or see it.
		// Or just run git branch in remote
		out, _ := gitUtil.RunGit(remoteDir, "branch")
		t.Logf("Remote branches: %s", out)
		// If main is there, it should be listed.
	}

	// Verify ref exists in remote
	if !gitUtil.RefExists(remoteDir, "refs/heads/main") {
		t.Errorf("Remote does not have refs/heads/main")
	}
}
