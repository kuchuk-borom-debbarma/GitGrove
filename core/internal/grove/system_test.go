package grove_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

func TestSwitchToSystem_CleansUntrackedFiles(t *testing.T) {
	// 1. Setup a temp directory as the root repo
	rootDir := t.TempDir()
	if err := gitUtil.Init(rootDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 2. Create gitgroove/system branch
	// We need to commit something first to have a branch
	if err := os.WriteFile(filepath.Join(rootDir, "README.md"), []byte("# Root"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if _, err := gitUtil.RunGit(rootDir, "add", "."); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	if _, err := gitUtil.RunGit(rootDir, "commit", "-m", "Initial commit"); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Create system branch
	if err := gitUtil.CreateBranch(rootDir, "gitgroove/system", "HEAD"); err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// 3. Create a child branch (simulating a registered repo)
	if err := gitUtil.CreateAndCheckoutBranch(rootDir, "child-branch"); err != nil {
		t.Fatalf("CreateAndCheckoutBranch failed: %v", err)
	}

	// 4. Create a file that is ignored in child-branch but NOT in system branch
	// Add .gitignore to child-branch
	if err := os.WriteFile(filepath.Join(rootDir, ".gitignore"), []byte("temp.log\n"), 0644); err != nil {
		t.Fatalf("WriteFile .gitignore failed: %v", err)
	}
	if _, err := gitUtil.RunGit(rootDir, "add", ".gitignore"); err != nil {
		t.Fatalf("git add .gitignore failed: %v", err)
	}
	if _, err := gitUtil.RunGit(rootDir, "commit", "-m", "Add gitignore"); err != nil {
		t.Fatalf("git commit .gitignore failed: %v", err)
	}

	// Create the ignored file
	tempFile := filepath.Join(rootDir, "temp.log")
	if err := os.WriteFile(tempFile, []byte("junk"), 0644); err != nil {
		t.Fatalf("WriteFile temp.log failed: %v", err)
	}

	// Verify it is ignored (clean state)
	if err := gitUtil.VerifyCleanState(rootDir); err != nil {
		t.Fatalf("VerifyCleanState failed (should be clean): %v", err)
	}

	// 5. Switch to system
	// Note: gitgroove/system does NOT have the .gitignore we just added to child-branch
	// So temp.log will become untracked.
	if err := grove.SwitchToSystem(rootDir); err != nil {
		t.Fatalf("SwitchToSystem failed: %v", err)
	}

	// 6. Verify we are on system branch
	currentBranch, err := gitUtil.GetCurrentBranch(rootDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if currentBranch != "gitgroove/system" {
		t.Errorf("Expected branch gitgroove/system, got %s", currentBranch)
	}

	// 7. Verify temp.log is GONE (cleaned)
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Errorf("temp.log should have been cleaned up, but it exists")
	}
}
