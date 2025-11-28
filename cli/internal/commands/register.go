package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type registerCmd struct{}

func (registerCmd) Command() string {
	return "register"
}

func (registerCmd) Description() string {
	return "Register a repository with --name and --path"
}

func (registerCmd) ValidateArgs(args map[string]any) error {
	if _, ok := args["name"]; !ok {
		return fmt.Errorf("missing required flag: --name")
	}
	if _, ok := args["path"]; !ok {
		return fmt.Errorf("missing required flag: --path")
	}
	return nil
}

func (registerCmd) Execute(args map[string]any) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	name, ok := args["name"].(string)
	if !ok {
		return fmt.Errorf("--name must be a string")
	}

	path, ok := args["path"].(string)
	if !ok {
		return fmt.Errorf("--path must be a string")
	}

	repos := map[string]string{
		name: path,
	}

	return core.Register(cwd, repos)
}

func init() {
	registerCommand(registerCmd{})
}
