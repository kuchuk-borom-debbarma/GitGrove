package commands

import (
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type initCommand struct{}

func (initCommand) Command() string {
	return "init"
}

func (initCommand) Description() string {
	return `Initialize GitGrove in an existing Git repository

Usage:
  gitgrove init

Note:
  Must be run from within a Git repository
  Creates the gitgroove/internal branch for metadata`
}

func (initCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (initCommand) Execute(args map[string]any) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return core.Init(cwd)
}

func init() {
	registerCommand(initCommand{})
}
