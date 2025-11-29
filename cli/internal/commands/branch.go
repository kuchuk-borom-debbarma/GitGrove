package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type branchCommand struct{}

func (branchCommand) Command() string {
	return "branch"
}

func (branchCommand) Description() string {
	return "Create a new branch for a nested repository"
}

func (branchCommand) ValidateArgs(args map[string]any) error {
	val, ok := args["args"]
	if !ok {
		return fmt.Errorf("missing required arguments: repo-name branch-name")
	}
	list, ok := val.([]string)
	if !ok || len(list) < 2 {
		return fmt.Errorf("usage: git grove branch <repo-name> <branch-name>")
	}
	return nil
}

func (branchCommand) Execute(args map[string]any) error {
	list := args["args"].([]string)
	repoName := list[0]
	branchName := list[1]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if err := core.CreateRepoBranch(cwd, repoName, branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Printf("Successfully created branch '%s' for repo '%s'\n", branchName, repoName)
	return nil
}

func init() {
	registerCommand(branchCommand{})
}
