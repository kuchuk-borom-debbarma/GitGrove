package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type linkCmd struct{}

func (linkCmd) Command() string {
	return "link"
}

func (linkCmd) Description() string {
	return "Link a child repository to a parent repository using --child and --parent"
}

func (linkCmd) ValidateArgs(args map[string]any) error {
	if _, ok := args["child"]; !ok {
		return fmt.Errorf("missing required flag: --child")
	}
	if _, ok := args["parent"]; !ok {
		return fmt.Errorf("missing required flag: --parent")
	}
	return nil
}

func (linkCmd) Execute(args map[string]any) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	child, ok := args["child"].(string)
	if !ok {
		return fmt.Errorf("--child must be a string")
	}

	parent, ok := args["parent"].(string)
	if !ok {
		return fmt.Errorf("--parent must be a string")
	}

	relationships := map[string]string{
		child: parent,
	}

	return core.Link(cwd, relationships)
}

func init() {
	registerCommand(linkCmd{})
}
