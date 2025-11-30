package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type lsCommand struct{}

func (lsCommand) Command() string {
	return "ls"
}

func (lsCommand) Description() string {
	return "List child repositories"
}

func (lsCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (lsCommand) Execute(args map[string]any) error {
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	gitRoot, err := findGitRoot(rootPath)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	children, err := core.Ls(gitRoot)
	if err != nil {
		return fmt.Errorf("failed to list: %w", err)
	}

	if len(children) == 0 {
		fmt.Println("(no children)")
		return nil
	}

	for _, child := range children {
		fmt.Println(child)
	}
	return nil
}

func init() {
	registerCommand(lsCommand{})
}
