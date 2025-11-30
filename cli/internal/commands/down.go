package commands

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type downCommand struct{}

func (downCommand) Command() string {
	return "down"
}

func (downCommand) Description() string {
	return "Switch to a child repository"
}

func (downCommand) ValidateArgs(args map[string]any) error {
	if len(args) == 0 {
		return fmt.Errorf("child repository name is required")
	}
	return nil
}

func (downCommand) Execute(args map[string]any) error {
	// Args parsing is a bit manual here since we don't have a strict schema in this framework yet
	// But main.go parses flags.
	// Wait, the current command framework passes a map.
	// We need to know how arguments are passed.
	// Looking at main.go, it seems it uses `args` map populated by flags?
	// Or positional args?
	// The `ValidateArgs` signature suggests a map.
	// Let's check `main.go` or `link.go` to see how they handle args.

	// `link` uses flags: --child, --parent.
	// `down` should probably take a positional argument or a flag.
	// Let's use a flag `--child` or just `child` if positional is supported.
	// The current framework seems to rely on flags defined in main.go?
	// No, main.go defines flags for specific commands.

	// I need to update main.go to parse arguments for `down`.
	// For now, let's assume I'll add `--repo` or similar to main.go for `down`.
	// Or better, let's use `child` as the key in the map.

	childName, ok := args["child"].(string)
	if !ok || childName == "" {
		// Check positional arguments
		if posArgs, ok := args["args"].([]string); ok && len(posArgs) > 0 {
			childName = posArgs[0]
		} else {
			return fmt.Errorf("child repository name is required (use --child or positional argument)")
		}
	}

	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	gitRoot, err := findGitRoot(rootPath)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	if err := core.Down(gitRoot, childName); err != nil {
		return fmt.Errorf("failed to go down: %w", err)
	}

	fmt.Printf("Switched to child repository '%s'.\n", childName)
	return nil
}

func init() {
	registerCommand(downCommand{})
}
