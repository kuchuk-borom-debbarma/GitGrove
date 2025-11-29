package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type statusCommand struct{}

func (statusCommand) Command() string {
	return "status"
}

func (statusCommand) Description() string {
	return "Show the working tree status"
}

func (statusCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (statusCommand) Execute(args map[string]any) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	output, err := core.Status(cwd)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}

func init() {
	registerCommand(statusCommand{})
}
