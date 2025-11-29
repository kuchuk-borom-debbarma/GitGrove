package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type SwitchCommand struct{}

func init() {
	registerCommand(&SwitchCommand{})
}

func (c *SwitchCommand) Command() string {
	return "switch"
}

func (c *SwitchCommand) Description() string {
	return "Switch to a specific repo branch"
}

func (c *SwitchCommand) ValidateArgs(args map[string]any) error {
	val, ok := args["args"]
	if !ok {
		return fmt.Errorf("missing required argument: repo")
	}

	list, ok := val.([]string)
	if !ok || len(list) < 1 {
		return fmt.Errorf("missing required argument: repo")
	}

	return nil
}

func (c *SwitchCommand) Execute(args map[string]any) error {
	list := args["args"].([]string)
	repoName := list[0]

	branch := ""
	if len(list) > 1 {
		branch = list[1]
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if err := core.Switch(cwd, repoName, branch); err != nil {
		return fmt.Errorf("switch failed: %w", err)
	}

	fmt.Printf("Switched to %s\n", repoName)
	return nil
}
