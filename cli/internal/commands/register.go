package commands

import (
	"fmt"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/cli/internal/util/git"
	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type registerCmd struct{}

func (registerCmd) Command() string {
	return "register"
}

func (registerCmd) Description() string {
	return "Register a repository using --name and --path, or positional args like name;path"
}

func (registerCmd) ValidateArgs(args map[string]any) error {
	hasName := false
	if _, ok := args["name"]; ok {
		hasName = true
	}
	hasPath := false
	if _, ok := args["path"]; ok {
		hasPath = true
	}

	hasPositional := false
	if val, ok := args["args"]; ok {
		if list, ok := val.([]string); ok && len(list) > 0 {
			hasPositional = true
		}
	}

	if !hasName && !hasPath && !hasPositional {
		return fmt.Errorf("missing required arguments: provide either --name and --path flags, or name;path positional arguments")
	}

	if (hasName && !hasPath) || (!hasName && hasPath) {
		return fmt.Errorf("both --name and --path must be provided together")
	}

	return nil
}

func (registerCmd) Execute(args map[string]any) error {
	root, err := git.FindRepoRoot()
	if err != nil {
		return err
	}

	repos := make(map[string]string)

	// Handle flags
	if name, ok := args["name"].(string); ok {
		if path, ok := args["path"].(string); ok {
			repos[name] = path
		}
	}

	// Handle positional args
	if val, ok := args["args"]; ok {
		if list, ok := val.([]string); ok {
			for _, arg := range list {
				parts := strings.SplitN(arg, ";", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid argument format '%s': expected name;path", arg)
				}
				name := strings.TrimSpace(parts[0])
				path := strings.TrimSpace(parts[1])
				if name == "" || path == "" {
					return fmt.Errorf("invalid argument '%s': name and path cannot be empty", arg)
				}
				repos[name] = path
			}
		}
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories specified")
	}

	return core.Register(root, repos)
}

func init() {
	registerCommand(registerCmd{})
}
