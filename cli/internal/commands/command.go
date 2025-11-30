package commands

import (
	"fmt"
	"sort"
)

type Command interface {
	// return the name of the command such as init
	Command() string
	// description
	Description() string
	// Validate if the required args are present
	ValidateArgs(args map[string]any) error
	// Execute the command
	Execute(args map[string]any) error
}

var commandRegistry = make(map[string]Command)

func registerCommand(command Command) {
	commandRegistry[command.Command()] = command
}

func GetCommand(name string) (Command, bool) {
	cmd, ok := commandRegistry[name]
	return cmd, ok
}

func ListCommands() []string {
	keys := make([]string, 0, len(commandRegistry))
	for k := range commandRegistry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type CommandRunner struct{}

func (CommandRunner) Run(command Command, args map[string]any) error {
	err := command.ValidateArgs(args)
	if err != nil {
		fmt.Println("[ERROR]: Invalid Arguments ", err)
		return err
	}
	err = command.Execute(args)
	if err != nil {
		fmt.Println("[ERROR]: Execution failed:", err)
		return err
	}
	return nil
}
