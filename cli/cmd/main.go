package main

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/cli/internal/commands"
	"github.com/kuchuk-borom-debbarma/GitGrove/cli/internal/util/arg"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	commandName := os.Args[1]
	command, ok := commands.GetCommand(commandName)
	if !ok {
		fmt.Printf("Unknown command: %s\n\n", commandName)
		printUsage()
		return
	}

	args := arg.ParseArg(os.Args[2:])
	runner := commands.CommandRunner{}
	if err := runner.Run(command, args); err != nil {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: gitgrove <command> [args]")
	fmt.Println("\nAvailable commands:")
	for _, cmdName := range commands.ListCommands() {
		cmd, _ := commands.GetCommand(cmdName)
		fmt.Printf("  %-12s %s\n", cmdName, cmd.Description())
	}
}

//TODO cd ~ should place you to gitgrove/system latest head and ls in this branch should show all repos with no parent i.e root repos
