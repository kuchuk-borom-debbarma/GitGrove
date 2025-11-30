package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core"
)

type interactiveCommand struct{}

func (interactiveCommand) Command() string {
	return "interactive"
}

func (interactiveCommand) Description() string {
	return "Start an interactive GitGrove session"
}

func (interactiveCommand) ValidateArgs(args map[string]any) error {
	return nil
}

func (interactiveCommand) Execute(args map[string]any) error {
	reader := bufio.NewReader(os.Stdin)

	// 1. Ask for repo path
	fmt.Print("Enter the GitHub repo path (default: current directory): ")
	repoPath, _ := reader.ReadString('\n')
	repoPath = strings.TrimSpace(repoPath)
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	fmt.Printf("Using repository at: %s\n", absPath)

	// 2. Detect initialization
	isInitialized, err := core.IsInitialized(absPath)
	if err != nil {
		// If check fails, treat as not initialized but log error if needed
		// For interactive session, we can just assume not initialized or print error
		fmt.Printf("Error checking initialization status: %v\n", err)
		isInitialized = false
	}

	if !isInitialized {
		fmt.Println("GitGrove is NOT initialized in this repository.")
		fmt.Print("Do you want to initialize it now? (y/n): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice == "y" || choice == "yes" {
			// Check if it's a git repo
			gitPath := filepath.Join(absPath, ".git")
			if _, err := os.Stat(gitPath); os.IsNotExist(err) {
				fmt.Println("This is not a git repository.")
				fmt.Print("Do you want to run 'git init' first? (y/n): ")
				gitChoice, _ := reader.ReadString('\n')
				gitChoice = strings.TrimSpace(strings.ToLower(gitChoice))
				if gitChoice == "y" || gitChoice == "yes" {
					if err := runGitInit(absPath); err != nil {
						return fmt.Errorf("failed to run git init: %w", err)
					}
					fmt.Println("Initialized empty Git repository.")
				} else {
					fmt.Println("Cannot initialize GitGrove without a git repository.")
					return nil
				}
			}

			if err := core.Init(absPath); err != nil {
				return fmt.Errorf("failed to initialize: %w", err)
			}
			fmt.Println("GitGrove initialized successfully!")
		} else {
			fmt.Println("Exiting interactive session.")
			return nil
		}
	} else {
		fmt.Println("GitGrove is initialized.")
	}

	// 3. Interactive Loop
	for {
		fmt.Println("\n--- GitGrove Interactive Session ---")
		fmt.Println("1. Register a Repository")
		fmt.Println("2. Link Repositories")
		fmt.Println("3. Switch Branch")
		fmt.Println("4. Exit")
		fmt.Print("Select an option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			handleRegister(reader, absPath)
		case "2":
			handleLink(reader, absPath)
		case "3":
			handleSwitch(reader, absPath)
		case "4":
			fmt.Println("Exiting...")
			return nil
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func handleRegister(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter Repository Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter Repository Path (relative to root): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	if name == "" || path == "" {
		fmt.Println("Error: Name and Path are required.")
		return
	}

	repos := map[string]string{name: path}
	if err := core.Register(rootPath, repos); err != nil {
		fmt.Printf("Error registering repository: %v\n", err)
	} else {
		fmt.Println("Repository registered successfully.")
	}
}

func handleLink(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter Child Repository Name: ")
	child, _ := reader.ReadString('\n')
	child = strings.TrimSpace(child)

	fmt.Print("Enter Parent Repository Name: ")
	parent, _ := reader.ReadString('\n')
	parent = strings.TrimSpace(parent)

	if child == "" || parent == "" {
		fmt.Println("Error: Child and Parent names are required.")
		return
	}

	relationships := map[string]string{child: parent}
	if err := core.Link(rootPath, relationships); err != nil {
		fmt.Printf("Error linking repositories: %v\n", err)
	} else {
		fmt.Println("Repositories linked successfully.")
	}
}

func handleSwitch(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter Repository Name: ")
	repoName, _ := reader.ReadString('\n')
	repoName = strings.TrimSpace(repoName)

	fmt.Print("Enter Branch Name: ")
	branch, _ := reader.ReadString('\n')
	branch = strings.TrimSpace(branch)

	if repoName == "" || branch == "" {
		fmt.Println("Error: Repository Name and Branch Name are required.")
		return
	}

	if err := core.Switch(rootPath, repoName, branch); err != nil {
		fmt.Printf("Error switching branch: %v\n", err)
	} else {
		fmt.Println("Switched branch successfully.")
	}
}

func runGitInit(path string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	registerCommand(interactiveCommand{})
}
