package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

func init() {
	registerCommand(&MoveCommand{})
}

type MoveCommand struct{}

func (c *MoveCommand) Command() string {
	return "move"
}

func (c *MoveCommand) Description() string {
	return "Move a registered repository to a new path using --repo and --to"
}

func (c *MoveCommand) ValidateArgs(args map[string]any) error {
	if _, ok := args["repo"]; !ok {
		return fmt.Errorf("missing required flag: --repo")
	}
	if _, ok := args["to"]; !ok {
		return fmt.Errorf("missing required flag: --to")
	}
	return nil
}

func (c *MoveCommand) Execute(args map[string]any) error {
	repoName, ok := args["repo"].(string)
	if !ok {
		return fmt.Errorf("--repo must be a string")
	}

	newPath, ok := args["to"].(string)
	if !ok {
		return fmt.Errorf("--to must be a string")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if err := core.Move(cwd, repoName, newPath); err != nil {
		return err
	}

	fmt.Printf("Successfully moved repo '%s' to '%s'\n", repoName, newPath)
	return nil
}
