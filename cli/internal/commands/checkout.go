package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type checkoutCommand struct{}

func (checkoutCommand) Command() string {
	return "checkout"
}

func (checkoutCommand) Description() string {
	return "Checkout a specific branch of a nested repository"
}

func (checkoutCommand) ValidateArgs(args map[string]any) error {
	val, ok := args["args"]
	if !ok {
		return fmt.Errorf("missing required arguments: repo-name branch-name")
	}
	list, ok := val.([]string)
	if !ok || len(list) < 2 {
		return fmt.Errorf("usage: git grove checkout <repo-name> <branch-name>")
	}
	return nil
}

func (checkoutCommand) Execute(args map[string]any) error {
	list := args["args"].([]string)
	repoName := list[0]
	branchName := list[1]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if err := core.CheckoutRepo(cwd, repoName, branchName); err != nil {
		return fmt.Errorf("failed to checkout repo: %w", err)
	}

	fmt.Printf("Successfully checked out repo '%s' to branch '%s'\n", repoName, branchName)
	return nil
}

func init() {
	registerCommand(checkoutCommand{})
}
