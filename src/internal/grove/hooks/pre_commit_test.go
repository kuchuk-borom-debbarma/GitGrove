package hooks

import (
	"os"
	"os/exec"
	"testing"

	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

func TestPreCommit(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "gg-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to tmpDir
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Init git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "you@example.com").Run()
	exec.Command("git", "config", "user.name", "Your Name").Run()

	// Case 1: No gg.json -> Should pass
	if err := PreCommit(); err != nil {
		t.Errorf("expected pass when no gg.json, got error: %v", err)
	}

	// 1. Setup mock environment
	err = groveUtil.CreateGroveConfig(tmpDir, false) // Changed setup.Cwd to tmpDir to match test context
	// assert.NoError(t, err) repos // This line seems to be a partial instruction, removing "repos" and adding error check
	if err != nil {
		t.Fatalf("failed to create grove config: %v", err)
	}
	repos := []model.GGRepo{
		{Name: "repoA", Path: "services/repoA"},
		{Name: "repoB", Path: "services/repoB"},
	}
	groveUtil.RegisterRepoInConfig(tmpDir, repos)

	// Create directories
	os.MkdirAll("services/repoA", 0755)
	os.MkdirAll("services/repoB", 0755)

	// Case 2: Modify repoA only -> Should pass
	os.WriteFile("services/repoA/file.txt", []byte("content"), 0644)
	exec.Command("git", "add", "services/repoA/file.txt").Run()
	if err := PreCommit(); err != nil {
		t.Errorf("expected pass for single repo commit, got error: %v", err)
	}
	exec.Command("git", "commit", "-m", "repoA commit").Run()

	// Case 3: Modify repoA AND repoB -> Should fail
	os.WriteFile("services/repoA/file2.txt", []byte("content"), 0644)
	os.WriteFile("services/repoB/file.txt", []byte("content"), 0644)
	exec.Command("git", "add", "services/repoA/file2.txt", "services/repoB/file.txt").Run()
	if err := PreCommit(); err == nil {
		t.Error("expected fail for mixed repo commit, got nil")
	} else {
		t.Logf("Got expected error: %v", err)
	}
	// Reset staging
	exec.Command("git", "reset").Run()

	// Case 4: Modify repoA AND root file -> Should fail
	os.WriteFile("services/repoA/file3.txt", []byte("content"), 0644)
	os.WriteFile("README.md", []byte("content"), 0644)
	exec.Command("git", "add", "services/repoA/file3.txt", "README.md").Run()
	if err := PreCommit(); err == nil {
		t.Error("expected fail for repo+root commit, got nil")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}
