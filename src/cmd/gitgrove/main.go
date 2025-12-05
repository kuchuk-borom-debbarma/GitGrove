package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/hooks"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/sync"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/tui"
)

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
			// Default CLI init to no atomic commit enforcement for now, or TODO: add flag
			if err := initialize.Initialize(cwd, false); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing GitGrove: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("GitGrove initialized successfully!")
			os.Exit(0)
		case "sync":
			targetArg := ""
			squash := true
			commit := false // Default: Squash and No Commit (Stage only)

			// Basic arg parsing: check args starting from index 2
			for i := 2; i < len(os.Args); i++ {
				arg := os.Args[i]
				if arg == "--no-squash" {
					squash = false
				} else if arg == "--commit" {
					commit = true
				} else if arg == "--repo" {
					if i+1 < len(os.Args) {
						targetArg = os.Args[i+1]
						i++ // Skip value
					}
				} else if !strings.HasPrefix(arg, "-") && targetArg == "" {
					targetArg = arg
				}
			}

			if err := sync.Sync(targetArg, squash, commit); err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
