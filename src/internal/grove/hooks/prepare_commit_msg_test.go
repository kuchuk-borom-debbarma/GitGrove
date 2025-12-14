package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/stretchr/testify/assert"
)

func TestPrepareCommitMsg(t *testing.T) {
	// Setup temporary directory acting as repo root
	tmpDir, err := os.MkdirTemp("", "gitgrove-test-prepare-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Switch CWD
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Init git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "you@example.com").Run()
	exec.Command("git", "config", "user.name", "Your Name").Run()

	// Init Grove Config with RepoAwareContextMessage = true
	repoA := "services/repoA"
	// Init Grove Config with AtomicCommit = true
	// config declaration removed as it was unused and we write manually

	// Manually write config because CreateGroveConfig doesn't let us inject repos easily without RegisterRepo
	os.MkdirAll(filepath.Join(tmpDir, ".gg"), 0755)
	// We can use RegisterRepoInConfig but manual is faster for test
	// Actually we can use groveUtil.RegisterRepoInConfig if we had it exported or similar.
	// We'll just write the file manually to ensure exact state.
	// But wait, groveUtil has LoadConfig. We need to save it.
	// groveUtil.CreateGroveConfig creates empty.
	// Let's use CreateGroveConfig then manually append.
	groveUtil.CreateGroveConfig(tmpDir, true)
	// Add repo manually
	// Re-load to modify? No, just overwrite.
	// grove_util.go doesn't export the struct if I didn't verify?
	// It is exported: type GGConfig struct
	// Writing manually.
	importJSON := `{"repositories": {"repoA": {"name": "repoA", "path": "services/repoA"}}, "repo_aware_context_message": true}`
	os.WriteFile(filepath.Join(tmpDir, ".gg", "gg.json"), []byte(importJSON), 0644)

	// Commit gg.json first so it doesn't interfere with "atomic" check for repoA
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "init gg.json").Run()

	// Create repo directory
	os.MkdirAll(filepath.Join(tmpDir, repoA), 0755)

	// Case 1: Modification of file in repoA -> Should prepend [repoA]
	testFile := filepath.Join(repoA, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)
	exec.Command("git", "add", ".").Run()

	msgFile := filepath.Join(tmpDir, "COMMIT_EDITMSG")
	os.WriteFile(msgFile, []byte("initial commit"), 0644)

	err = PrepareCommitMsg(msgFile, "", "")
	assert.NoError(t, err)

	content, _ := os.ReadFile(msgFile)
	assert.Equal(t, "[repoA] initial commit", string(content))

	// Case 2: Root file modification -> No prepend
	// Clean up
	exec.Command("git", "commit", "-m", "committing").Run()

	rootFile := filepath.Join(tmpDir, "root.txt")
	os.WriteFile(rootFile, []byte("root content"), 0644)
	exec.Command("git", "add", ".").Run()

	os.WriteFile(msgFile, []byte("root change"), 0644)
	err = PrepareCommitMsg(msgFile, "", "")
	assert.NoError(t, err)
	content, _ = os.ReadFile(msgFile)
	assert.Equal(t, "root change", string(content))

	// Case 3: Atomic Commit Disabled -> No prepend
	// Update config
	importJSON = `{"repositories": {"repoA": {"name": "repoA", "path": "services/repoA"}}, "repo_aware_context_message": false}`
	os.WriteFile(filepath.Join(tmpDir, ".gg", "gg.json"), []byte(importJSON), 0644)
	// Commit config change
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "disable atomic").Run()

	// Modify repoA again
	os.WriteFile(testFile, []byte("content2"), 0644)
	exec.Command("git", "add", ".").Run()

	os.WriteFile(msgFile, []byte("another commit"), 0644)
	err = PrepareCommitMsg(msgFile, "", "")
	assert.NoError(t, err)
	content, _ = os.ReadFile(msgFile)
	assert.Equal(t, "another commit", string(content))
	// Case 4: Orphan branch with missing .gg/gg.json but present in trunk
	// Switch back to master (has config) to act as trunk
	exec.Command("git", "checkout", "master").Run() // "master" is default in our test setup?
	// Wait, we didn't explicitly name the trunk branch, just committed to it.
	// git init likely created "master" or "main".
	// Let's create a known trunk "trunk" with config.
	exec.Command("git", "checkout", "-b", "trunk").Run()

	// Reset config to true for this test case
	importJSON = `{"repositories": {"repoA": {"name": "repoA", "path": "services/repoA"}}, "repo_aware_context_message": true}`
	// Ensure config is present in trunk
	os.WriteFile(filepath.Join(tmpDir, ".gg", "gg.json"), []byte(importJSON), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "add config to trunk").Run()

	// Switch to orphan branch gg/trunk/repoA
	orphanBranch := "gg/trunk/repoA"
	exec.Command("git", "checkout", "--orphan", orphanBranch).Run()
	// Remove all files from previous branch (standard orphan behavior usually leaves files staged)
	// We want to simulate a clean state for the repo except for what we are working on.
	// But in GitGrove workspace, we are in a separate worktree or clone usually.
	// Here we simulate just missing config.
	os.RemoveAll(filepath.Join(tmpDir, ".gg")) // Delete .gg dir

	// Modify file in repoA (which is now root relative in some setups, but here we just need to match repoName)
	// Wait, if it's orphan, EVERYTHING is the repo.
	// PrepareCommitMsg logic: if orphan branch gg/<trunk>/<repoName>, ALWAYS prepend [repoName].
	// It doesn't check file paths.

	msgFile4 := filepath.Join(tmpDir, "COMMIT_EDITMSG_4")
	os.WriteFile(msgFile4, []byte("orphan commit"), 0644)

	err = PrepareCommitMsg(msgFile4, "", "")
	assert.NoError(t, err)

	content, _ = os.ReadFile(msgFile4)
	assert.Equal(t, "[repoA] orphan commit", string(content))

	// Case 5: Sticky Context Logic
	// Switch to a new branch off trunk (no orphan prefix)
	exec.Command("git", "checkout", "trunk").Run()
	exec.Command("git", "checkout", "-b", "feature/sticky-test").Run()

	// Set sticky context
	exec.Command("git", "config", "--local", "gitgrove.context.repo", "repoA").Run()

	msgFile5 := filepath.Join(tmpDir, "COMMIT_EDITMSG_5")
	os.WriteFile(msgFile5, []byte("sticky commit"), 0644)

	err = PrepareCommitMsg(msgFile5, "", "")
	assert.NoError(t, err)

	content, _ = os.ReadFile(msgFile5)
	assert.Equal(t, "[repoA] sticky commit", string(content))
}
