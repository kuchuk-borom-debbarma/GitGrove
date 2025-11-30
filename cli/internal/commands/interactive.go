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
		fmt.Println("1. Info")
		fmt.Println("2. Register a Repository")
		fmt.Println("3. Link Repositories")
		fmt.Println("4. Create Branch")
		fmt.Println("5. Checkout Branch")
		fmt.Println("6. Switch Branch (Legacy)")
		fmt.Println("7. Stage Files")
		fmt.Println("8. Commit Changes")
		fmt.Println("9. Move Repository")
		fmt.Println("10. Exit")
		fmt.Print("Select an option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			handleInfo(absPath)
		case "2":
			handleRegister(reader, absPath)
		case "3":
			handleLink(reader, absPath)
		case "4":
			handleBranch(reader, absPath)
		case "5":
			handleCheckout(reader, absPath)
		case "6":
			handleSwitch(reader, absPath)
		case "7":
			handleStage(reader, absPath)
		case "8":
			handleCommit(reader, absPath)
		case "9":
			handleMove(reader, absPath)
		case "10":
			fmt.Println("Exiting...")
			return nil
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func handleInfo(rootPath string) {
	info, err := core.Info(rootPath)
	if err != nil {
		fmt.Printf("Error getting info: %v\n", err)
	} else {
		fmt.Println(info)
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

func handleBranch(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter Repository Name: ")
	repoName, _ := reader.ReadString('\n')
	repoName = strings.TrimSpace(repoName)

	fmt.Print("Enter New Branch Name: ")
	branch, _ := reader.ReadString('\n')
	branch = strings.TrimSpace(branch)

	if repoName == "" || branch == "" {
		fmt.Println("Error: Repository Name and Branch Name are required.")
		return
	}

	if err := core.CreateRepoBranch(rootPath, repoName, branch); err != nil {
		fmt.Printf("Error creating branch: %v\n", err)
	} else {
		fmt.Println("Branch created successfully.")
	}
}

func handleCheckout(reader *bufio.Reader, rootPath string) {
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

	if err := core.CheckoutRepo(rootPath, repoName, branch); err != nil {
		fmt.Printf("Error checking out branch: %v\n", err)
	} else {
		fmt.Println("Checked out branch successfully.")
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

func handleStage(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter file paths to stage (comma separated, or '.' for all): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("Error: File paths are required.")
		return
	}

	var files []string
	if input == "." {
		files = []string{"."}
	} else {
		parts := strings.Split(input, ",")
		for _, p := range parts {
			files = append(files, strings.TrimSpace(p))
		}
	}

	if err := core.Stage(rootPath, files); err != nil {
		fmt.Printf("Error staging files: %v\n", err)
	} else {
		fmt.Println("Files staged successfully.")
	}
}

func handleCommit(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter commit message: ")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	if message == "" {
		fmt.Println("Error: Commit message is required.")
		return
	}

	if err := core.Commit(rootPath, message); err != nil {
		fmt.Printf("Error committing changes: %v\n", err)
	} else {
		fmt.Println("Changes committed successfully.")
	}
}

func handleMove(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter Repository Name: ")
	repoName, _ := reader.ReadString('\n')
	repoName = strings.TrimSpace(repoName)

	fmt.Print("Enter New Relative Path: ")
	newPath, _ := reader.ReadString('\n')
	newPath = strings.TrimSpace(newPath)

	if repoName == "" || newPath == "" {
		fmt.Println("Error: Repository Name and New Path are required.")
		return
	}

	if err := core.Move(rootPath, repoName, newPath); err != nil {
		fmt.Printf("Error moving repository: %v\n", err)
	} else {
		fmt.Println("Repository moved successfully.")
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
