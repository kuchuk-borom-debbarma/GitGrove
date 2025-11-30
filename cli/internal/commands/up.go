package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type upCommand struct{}

func (upCommand) Command() string {
	return "up"
}

func (upCommand) Description() string {
	return "Switch to the parent repository"
}

func (upCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (upCommand) Execute(args map[string]any) error {
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Resolve root path if we are in a subdirectory (though GitGrove ops usually handle this)
	// Actually, Up needs to find the .git root first.
	// core.Up handles this via rootAbsPath passed to it.
	// But we need to find the root of the git repo first.
	// Let's assume current dir is fine and let core handle it, or find git root.
	// core.Up expects rootAbsPath to be the root of the git repo.

	// We should find the git root.
	// Simple helper to find git root?
	// Or just pass current dir and let core resolve?
	// core functions usually expect the root of the repo.

	// Let's find the git root.
	gitRoot, err := findGitRoot(rootPath)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	if err := core.Up(gitRoot); err != nil {
		return fmt.Errorf("failed to go up: %w", err)
	}

	fmt.Println("Switched to parent repository.")
	return nil
}

func findGitRoot(path string) (string, error) {
	// Simple traversal up
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("git root not found")
		}
		dir = parent
	}
}

func init() {
	registerCommand(upCommand{})
}
