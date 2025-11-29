package grove_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
)

func TestCheckoutRepo(t *testing.T) {
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

	// Switch back to default branch
	execGit(t, temp, "checkout", defaultBranch)

	// Create directories for repos
	os.MkdirAll(filepath.Join(temp, "services", "parent"), 0755)
	os.MkdirAll(filepath.Join(temp, "services", "child"), 0755)

	// Register repos
	repos := map[string]string{
		"parent": "services/parent",
		"child":  "services/child",
	}
	if err := grove.Register(temp, repos); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Link repos: parent -> child
	links := map[string]string{
		"child": "parent",
	}
	if err := grove.Link(temp, links); err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// Create a branch for the child repo
	branchName := "feature/checkout-test"
	if err := grove.CreateRepoBranch(temp, "child", branchName); err != nil {
		t.Fatalf("CreateRepoBranch failed: %v", err)
	}

	// Verify we are currently on default branch
	currentBranch := strings.TrimSpace(execGit(t, temp, "branch", "--show-current"))
	if currentBranch != defaultBranch {
		t.Fatalf("Expected to be on %s, got %s", defaultBranch, currentBranch)
	}

	// Test Case: Checkout the new branch
	if err := grove.CheckoutRepo(temp, "child", branchName); err != nil {
		t.Fatalf("CheckoutRepo failed: %v", err)
	}

	// Verify we switched to the correct branch
	// Expected branch name: gitgroove/repos/parent/children/child/branches/feature/checkout-test
	expectedBranch := "gitgroove/repos/parent/children/child/branches/feature/checkout-test"

	newBranch := strings.TrimSpace(execGit(t, temp, "branch", "--show-current"))
	if newBranch != expectedBranch {
		t.Errorf("Expected to be on branch %s, got %s", expectedBranch, newBranch)
	}
}
