package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type cdCommand struct{}

func (cdCommand) Command() string {
	return "cd"
}

func (cdCommand) Description() string {
	return `Navigate repository hierarchy like a filesystem

Usage:
  gitgrove cd <target>

Examples:
  gitgrove cd auth        # Navigate to auth repository
  gitgrove cd ..          # Go up to parent repository`
}

func (cdCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (cdCommand) Execute(args map[string]any) error {
	// Get target from positional args
	target := ""
	if posArgs, ok := args["args"].([]string); ok && len(posArgs) > 0 {
		target = posArgs[0]
	}

	if target == "" {
		return fmt.Errorf("target required: use '..' for parent or repository name for child")
	}

	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	gitRoot, err := findGitRoot(rootPath)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	if err := core.Cd(gitRoot, target); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	if target == ".." {
		fmt.Println("Switched to parent repository.")
	} else {
		fmt.Printf("Switched to '%s'.\n", target)
	}
	return nil
}

func init() {
	registerCommand(cdCommand{})
}
