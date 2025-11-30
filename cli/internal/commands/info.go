package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type infoCommand struct{}

func (infoCommand) Command() string {
	return "info"
}

func (infoCommand) Description() string {
	return "Check the health and status of the GitGrove system"
}

func (infoCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (infoCommand) Execute(args map[string]any) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	output, err := core.Info(cwd)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}

func init() {
	registerCommand(infoCommand{})
}
