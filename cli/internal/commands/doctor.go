package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type doctorCommand struct{}

func (doctorCommand) Command() string {
	return "doctor"
}

func (doctorCommand) Description() string {
	return "Check the health and status of the GitGrove system"
}

func (doctorCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (doctorCommand) Execute(args map[string]any) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	output, err := core.Doctor(cwd)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}

func init() {
	registerCommand(doctorCommand{})
}
