package commands

import (
	"fmt"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/cli/internal/util/git"
	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type linkCmd struct{}

func (linkCmd) Command() string {
	return "link"
}

func (linkCmd) Description() string {
	return "Link a child repository to a parent using --child and --parent, or positional args like child;parent"
}

func (linkCmd) ValidateArgs(args map[string]any) error {
	hasChild := false
	if _, ok := args["child"]; ok {
		hasChild = true
	}
	hasParent := false
	if _, ok := args["parent"]; ok {
		hasParent = true
	}

	hasPositional := false
	if val, ok := args["args"]; ok {
		if list, ok := val.([]string); ok && len(list) > 0 {
			hasPositional = true
		}
	}

	if !hasChild && !hasParent && !hasPositional {
		return fmt.Errorf("missing required arguments: provide either --child and --parent flags, or child;parent positional arguments")
	}

	if (hasChild && !hasParent) || (!hasChild && hasParent) {
		return fmt.Errorf("both --child and --parent must be provided together")
	}

	return nil
}

func (linkCmd) Execute(args map[string]any) error {
	root, err := git.FindRepoRoot()
	if err != nil {
		return err
	}

	relationships := make(map[string]string)

	// Handle flags
	if child, ok := args["child"].(string); ok {
		if parent, ok := args["parent"].(string); ok {
			relationships[child] = parent
		}
	}

	// Handle positional args
	if val, ok := args["args"]; ok {
		if list, ok := val.([]string); ok {
			for _, arg := range list {
				parts := strings.SplitN(arg, ";", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid argument format '%s': expected child;parent", arg)
				}
				child := strings.TrimSpace(parts[0])
				parent := strings.TrimSpace(parts[1])
				if child == "" || parent == "" {
					return fmt.Errorf("invalid argument '%s': child and parent cannot be empty", arg)
				}
				relationships[child] = parent
			}
		}
	}

	if len(relationships) == 0 {
		return fmt.Errorf("no relationships specified")
	}

	return core.Link(root, relationships)
}

func init() {
	registerCommand(linkCmd{})
}
