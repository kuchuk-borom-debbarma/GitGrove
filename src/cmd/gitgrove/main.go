package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/hooks"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	preparemerge "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/prepare-merge"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/tui"
)

var BuildTime = "unknown"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hook":
			if len(os.Args) >= 3 {
				switch os.Args[2] {
				case "pre-commit":
					if err := hooks.PreCommit(); err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						os.Exit(1)
					}
					os.Exit(0)
				case "prepare-commit-msg":
					// args: <msgFile> <source> <sha>
					// os.Args[0] = git-grove, [1]=hook, [2]=prepare-commit-msg, [3]=msgFile, [4]=source, [5]=sha
					if len(os.Args) < 4 {
						// minimal is msgFile
						os.Exit(0)
					}
					msgFile := os.Args[3]
					source := ""
					if len(os.Args) > 4 {
						source = os.Args[4]
					}
					sha := ""
					if len(os.Args) > 5 {
						sha = os.Args[5]
					}
					if err := hooks.PrepareCommitMsg(msgFile, source, sha); err != nil {
						fmt.Fprintf(os.Stderr, "Error in prepare-commit-msg: %v\n", err)
						// We don't exit 1 on hook error usually unless we want to abort commit,
						// but message prep failure might be annoying if it aborts.
						// pre-commit hook aborts commit. prepare-commit-msg just preps it.
						// If we fail to prep, maybe we should let it pass?
						// But if we fail we might want to know.
						// For now, let's just print error and exit 0 to allow commit to proceed without mod?
						// Or Exit 1 to stop? Let's generic exit 1 if critical.
						os.Exit(0)
					}
					os.Exit(0)
				}
			}
			fmt.Println("Usage: git-grove hook <pre-commit|prepare-commit-msg>")
			os.Exit(1)
		case "init":
			cwd, _ := os.Getwd()
			atomic := false
			// Simple flag parsing for init command
			for _, arg := range os.Args[2:] {
				if arg == "--atomic" {
					atomic = true
				}
			}

			if err := initialize.Initialize(cwd, atomic); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing GitGrove: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("GitGrove initialized successfully!")
			os.Exit(0)
		case "prepare-merge":
			cwd, _ := os.Getwd()
			repoName := ""
			if len(os.Args) > 2 {
				repoName = os.Args[2]
			}
			if err := preparemerge.PrepareMerge(cwd, repoName); err != nil {
				fmt.Fprintf(os.Stderr, "Error preparing merge: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	p := tea.NewProgram(tui.InitialModel(BuildTime), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
