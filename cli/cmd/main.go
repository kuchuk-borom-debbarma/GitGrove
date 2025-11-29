package main

import (
	"fmt"
	"os"

	"github.com/kuchuk-borom-debbarma/GitGrove/cli/internal/commands"
	"github.com/kuchuk-borom-debbarma/GitGrove/cli/internal/util/arg"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gitgrove <command> [args]")
		return
	}

	commandName := os.Args[1]
	command, ok := commands.GetCommand(commandName)
	if !ok {
		fmt.Println("Unknown command:", commandName)
		return
	}

	args := arg.ParseArg(os.Args[2:])
	runner := commands.CommandRunner{}
	if err := runner.Run(command, args); err != nil {
		os.Exit(1)
	}
}
