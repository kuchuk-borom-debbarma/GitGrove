package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/hooks"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/tui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hook":
			if len(os.Args) > 2 && os.Args[2] == "pre-commit" {
				if err := hooks.PreCommit(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			}
			fmt.Println("Usage: git-grove hook pre-commit")
			os.Exit(1)
		case "init":
			cwd, _ := os.Getwd()
			if err := initialize.Initialize(cwd); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing GitGrove: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("GitGrove initialized successfully!")
			os.Exit(0)
		}
	}

	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
