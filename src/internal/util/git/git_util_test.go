package gitUtil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClean(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "git-clean-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Init git repo
	exec.Command("git", "init", tmpDir).Run()

	// Create untracked file
	untrackedFile := filepath.Join(tmpDir, "untracked.txt")
	os.WriteFile(untrackedFile, []byte("garbage"), 0644)

	// Create untracked dir
	untrackedDir := filepath.Join(tmpDir, "garbage_dir")
	os.Mkdir(untrackedDir, 0755)
	os.WriteFile(filepath.Join(untrackedDir, "file.txt"), []byte("more garbage"), 0644)

	// Verify they exist
	assert.FileExists(t, untrackedFile)
	assert.DirExists(t, untrackedDir)

	// Add .gitignore and an ignored file
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("ignored.txt"), 0644)
	ignoredFile := filepath.Join(tmpDir, "ignored.txt")
	os.WriteFile(ignoredFile, []byte("ignored"), 0644)

	// Execute Clean
	err = Clean(tmpDir)
	assert.NoError(t, err)

	// Verify they are gone
	assert.NoFileExists(t, untrackedFile)
	assert.NoDirExists(t, untrackedDir)

	// This assertion is expected to FAIL with current implementation (-fd) which keeps ignored files
	// If the user wants "dirs and all" gone, they imply ignored files too.
	assert.NoFileExists(t, ignoredFile)
}
