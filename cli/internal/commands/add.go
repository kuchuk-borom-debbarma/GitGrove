package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuchuk-borom-debbarma/GitGrove/cli/internal/util/git"
	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type AddCommand struct{}

func init() {
	registerCommand(&AddCommand{})
}

func (c *AddCommand) Command() string {
	return "add"
}

func (c *AddCommand) Description() string {
	return "Add files with GitGrove validation (skips out-of-bound files with warning)"
}

func (c *AddCommand) ValidateArgs(args map[string]any) error {
	val, ok := args["args"]
	if !ok {
		return fmt.Errorf("missing required argument: files")
	}

	list, ok := val.([]string)
	if !ok || len(list) < 1 {
		return fmt.Errorf("missing required argument: files")
	}

	return nil
}

func (c *AddCommand) Execute(args map[string]any) error {
	files := args["args"].([]string)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	root, err := git.FindRepoRoot()
	if err != nil {
		return err
	}

	// Resolve all files to absolute paths
	var absFiles []string
	for _, f := range files {
		if f == "." {
			absFiles = append(absFiles, cwd)
		} else {
			if filepath.IsAbs(f) {
				absFiles = append(absFiles, f)
			} else {
				absFiles = append(absFiles, filepath.Join(cwd, f))
			}
		}
	}

	if err := core.Add(root, absFiles); err != nil {
		return fmt.Errorf("add failed: %w", err)
	}

	return nil
}
