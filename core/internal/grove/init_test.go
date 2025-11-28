package grove_test

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove"
	fileUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/file"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

// ------------------------------------------------------------
// HELPER: create dummy file structure
// ------------------------------------------------------------
func createDummyProject(t *testing.T, root string) {
	t.Helper()

	files := []string{
		"src/main.go",
		"src/utils/helper.go",
		"README.md",
		".gitignore",
		"config/app.yaml",
	}

	for _, f := range files {
		path := filepath.Join(root, f)
		if err := fileUtil.WriteTextFile(path, "// dummy content"); err != nil {
			t.Fatalf("failed to write dummy file %s: %v", f, err)
		}
	}
}

// ------------------------------------------------------------
// HELPER: run git inside test repo
// ------------------------------------------------------------
func execGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s", args, string(out))
	}
	return strings.TrimSpace(string(out))
}

// ------------------------------------------------------------
// Test 1: Init should succeed when repo is clean & initialized
// ------------------------------------------------------------
func TestInitOnCleanRepo(t *testing.T) {
	temp := t.TempDir()

	// Create dummy files
	createDummyProject(t, temp)

	// git init manually (Init() no longer does this)
	execGit(t, temp, "init")
	execGit(t, temp, "add", ".")
	execGit(t, temp, "commit", "-m", "initial commit")

	// Sanity check
	if !gitUtil.IsInsideGitRepo(temp) {
		t.Fatalf("repo not initialized")
	}

	// Make sure repo is clean
	if err := gitUtil.VerifyCleanState(temp); err != nil {
		t.Fatalf("repo not clean before init: %v", err)
	}

	// Run GitGroove init
	if err := grove.Init(temp); err != nil {
		t.Fatalf("Init() failed on clean repo: %v", err)
	}

	// Verify .gg exists
	if !fileUtil.Exists(filepath.Join(temp, ".gg")) {
		t.Fatalf(".gg directory missing after init")
	}

	// Verify .gg/repos exists
	if !fileUtil.Exists(filepath.Join(temp, ".gg/repos")) {
		t.Fatalf(".gg/repos directory missing after init")
	}

	// Verify .gitkeep exists
	if !fileUtil.Exists(filepath.Join(temp, ".gg/repos/.gitkeep")) {
		t.Fatalf(".gg/repos/.gitkeep missing after init")
	}

	// Verify system branch exists
	exists, err := gitUtil.HasBranch(temp, "gitgroove/system")
	if err != nil || !exists {
		t.Fatalf("gitgroove/system branch missing")
	}

	// Verify HEAD is on system branch
	head := execGit(t, temp, "branch", "--show-current")
	if head != "gitgroove/system" {
		t.Fatalf("expected HEAD on gitgroove/system, got %s", head)
	}

	// Metadata should be committed
	count := gitLogCount(t, temp)
	if count != 2 { // initial commit + system commit
		t.Fatalf("expected 2 commits, got %d", count)
	}
}

// ------------------------------------------------------------
// Test 2: Init should fail when git repo does NOT exist
// ------------------------------------------------------------
func TestInitFailsOnNoGitRepo(t *testing.T) {
	temp := t.TempDir()

	createDummyProject(t, temp)

	if gitUtil.IsInsideGitRepo(temp) {
		t.Fatalf("temp unexpectedly inside git repo")
	}

	err := grove.Init(temp)
	if err == nil {
		t.Fatalf("Init() should have failed on non-git directory")
	}
}

// ------------------------------------------------------------
// Test 3: Init should fail on dirty repo
// ------------------------------------------------------------
func TestInitFailsOnDirtyRepo(t *testing.T) {
	temp := t.TempDir()

	// Create dummy files
	createDummyProject(t, temp)

	// initialize git
	execGit(t, temp, "init")

	// DO NOT COMMIT — repo becomes dirty
	execGit(t, temp, "add", ".") // staged
	// Or: leave untracked files

	err := grove.Init(temp)
	if err == nil {
		t.Fatalf("Init() should fail on dirty repo")
	}
}

// ------------------------------------------------------------
// Helper: count commits
// ------------------------------------------------------------
func gitLogCount(t *testing.T, dir string) int {
	out := execGit(t, dir, "rev-list", "--count", "HEAD")
	n, err := strconv.Atoi(out)
	if err != nil {
		t.Fatalf("invalid commit count: %v", err)
	}
	return n
}
