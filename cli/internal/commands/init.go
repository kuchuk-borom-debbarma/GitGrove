package commands

import (
	"os"
	"os/exec"
)

type initCommand struct{}

func (initCommand) Command() string {
	return "init"
}

func (initCommand) Description() string {
	return "Initialize a new GitGrove repository"
}

func (initCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (initCommand) Execute(args map[string]any) error {
	cmd := exec.Command("./gitgrove", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func init() {
	registerCommand(initCommand{})
}
