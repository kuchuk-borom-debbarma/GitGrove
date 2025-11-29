package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type CommitCommand struct{}

func init() {
	registerCommand(&CommitCommand{})
}

func (c *CommitCommand) Command() string {
	return "commit"
}

func (c *CommitCommand) Description() string {
	return "Commit staged changes with GitGrove validation"
}

func (c *CommitCommand) ValidateArgs(args map[string]any) error {
	if _, ok := args["message"]; ok {
		return nil
	}
	if _, ok := args["m"]; ok {
		return nil
	}
	return fmt.Errorf("missing required flag: -m (message)")
}

func (c *CommitCommand) Execute(args map[string]any) error {
	var message string
	if val, ok := args["message"].(string); ok {
		message = val
	} else if val, ok := args["m"].(string); ok {
		message = val
	} else {
		return fmt.Errorf("-m must be a string")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if err := core.Commit(cwd, message); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
