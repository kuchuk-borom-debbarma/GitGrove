package grove_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

func TestCreateRepoBranch(t *testing.T) {
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
	os.MkdirAll(filepath.Join(temp, "services", "grandparent"), 0755)
	os.MkdirAll(filepath.Join(temp, "services", "parent"), 0755)
	os.MkdirAll(filepath.Join(temp, "services", "child"), 0755)

	// Register repos
	repos := map[string]string{
		"grandparent": "services/grandparent",
		"parent":      "services/parent",
		"child":       "services/child",
	}
	if err := grove.Register(temp, repos); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Register creates .gitgroverepo files which are untracked.
	// We need to commit them to have a clean state for Link.
	execGit(t, temp, "add", ".")
	execGit(t, temp, "commit", "-m", "Add .gitgroverepo markers")

	// Link repos: grandparent -> parent -> child
	links := map[string]string{
		"parent": "grandparent",
		"child":  "parent",
	}
	if err := grove.Link(temp, links); err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// Test Case: Create branch for nested child
	branchName := "feature/login"
	if err := grove.CreateRepoBranch(temp, "child", branchName); err != nil {
		t.Fatalf("CreateRepoBranch failed: %v", err)
	}

	// Verify the branch ref exists
	// Expected path: gitgroove/repos/grandparent/children/parent/children/child/branches/feature/login
	expectedRef := "refs/heads/gitgroove/repos/grandparent/children/parent/children/child/branches/feature/login"

	exists, err := gitUtil.HasBranch(temp, expectedRef)
	if err != nil {
		t.Fatalf("HasBranch failed: %v", err)
	}
	if !exists {
		t.Errorf("Expected branch ref %s to exist, but it does not", expectedRef)
	}

	// Verify it points to HEAD
	headCommit, _ := gitUtil.GetHeadCommit(temp)
	branchCommit, _ := gitUtil.ResolveRef(temp, expectedRef)
	if branchCommit != headCommit {
		t.Errorf("Expected branch to point to %s, got %s", headCommit, branchCommit)
	}
}
