package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/hooks"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/initialize"
	preparemerge "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/prepare-merge"
	registerrepo "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/register-repo"
	grovesync "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/grove/sync"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/internal/tui"
	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
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
		case "register":
			if len(os.Args) < 4 {
				fmt.Println("Usage: gg register <name> <path>")
				os.Exit(1)
			}
			cwd, _ := os.Getwd()
			name := os.Args[2]
			path := os.Args[3]
			repo := model.GGRepo{Name: name, Path: path}
			if err := registerrepo.RegisterRepo([]model.GGRepo{repo}, cwd); err != nil {
				fmt.Fprintf(os.Stderr, "Error registering repo: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully registered repo '%s'\n", name)
			os.Exit(0)
		case "checkout":
			if len(os.Args) < 3 {
				fmt.Println("Usage: gg checkout <repo-name>")
				os.Exit(1)
			}
			cwd, _ := os.Getwd()
			repoName := os.Args[2]

			// Determine Trunk
			trunk, err := groveUtil.GetContextTrunk(cwd)
			if err != nil || trunk == "" {
				// Try falling back to current branch if we are on trunk
				cb, _ := gitUtil.CurrentBranch(cwd)
				if cb != "" {
					trunk = cb
				} else {
					fmt.Fprintf(os.Stderr, "Error: Could not determine trunk branch. Ensure you are in a GitGrove workspace.\n")
					os.Exit(1)
				}
			}

			// Construct target branch: gg/<trunk>/<repo>
			targetBranch := fmt.Sprintf("gg/%s/%s", trunk, repoName)

			if err := gitUtil.Checkout(cwd, targetBranch); err != nil {
				fmt.Fprintf(os.Stderr, "Error checking out %s: %v\n", targetBranch, err)
				os.Exit(1)
			}

			// Clean artifacts
			if err := gitUtil.Clean(cwd); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Checkout succeeded but clean failed: %v\n", err)
			}

			// Set sticky context
			_ = groveUtil.SetContextRepo(cwd, repoName)
			_ = groveUtil.SetContextTrunk(cwd, trunk)
			_ = groveUtil.SetContextOrphan(cwd, targetBranch)

			fmt.Printf("Switched to orphan branch: %s\n", targetBranch)
			os.Exit(0)
		case "sync":
			cwd, _ := os.Getwd()
			// Let SyncOrphanFromTrunk infer context
			if err := grovesync.SyncOrphanFromTrunk(cwd, "", "", ""); err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing from trunk: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Successfully synced from trunk.")
			os.Exit(0)
		case "trunk":
			cwd, _ := os.Getwd()
			trunk, err := groveUtil.GetContextTrunk(cwd)
			if err != nil || trunk == "" {
				fmt.Fprintf(os.Stderr, "Error: Unknown trunk branch. Are you in a GitGrove orphan branch?\n")
				os.Exit(1)
			}
			if err := gitUtil.Checkout(cwd, trunk); err != nil {
				fmt.Fprintf(os.Stderr, "Error returning to trunk: %v\n", err)
				os.Exit(1)
			}
			// Clear context
			groveUtil.ClearAllContext(cwd)

			fmt.Printf("Returned to trunk branch: %s\n", trunk)
			os.Exit(0)
		}
	}

	p := tea.NewProgram(tui.InitialModel(BuildTime), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
