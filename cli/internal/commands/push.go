package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type pushCommand struct{}

func (pushCommand) Command() string {
	return "push"
}

func (pushCommand) Description() string {
	return "Push repositories to remote"
}

func (pushCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (pushCommand) Execute(args map[string]any) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	rootPath, err := findGitRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// Parse arguments
	// We expect args["args"] to contain positional arguments
	var targets []string
	if posArgs, ok := args["args"].([]string); ok && len(posArgs) > 0 {
		// Handle comma-separated values if user provides them like "repo1,repo2"
		for _, arg := range posArgs {
			parts := strings.Split(arg, ",")
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					targets = append(targets, trimmed)
				}
			}
		}
	}

	// If no targets provided, default to interactive selection?
	// Or maybe fail? The user said "interactive ui show what we want to push repo".
	// But this is the CLI command.
	// If I run `grove push`, it should probably ask or push current?
	// Let's assume if no args, we push the current repo if we are inside one?
	// Or maybe just error and say "specify repo or *".
	// Wait, the user requirement: "add the cli, interaction, grove command".
	// Let's support:
	// `grove push *` -> push all
	// `grove push repo1 repo2` -> push specific
	// `grove push` -> if inside a repo, push that repo? Or show help?

	if len(targets) == 0 {
		return fmt.Errorf("please specify repositories to push (e.g., 'grove push *' or 'grove push repo1,repo2')")
	}

	if err := core.Push(rootPath, targets); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	fmt.Println("✅ Push completed successfully.")
	return nil
}

func init() {
	registerCommand(pushCommand{})
}
