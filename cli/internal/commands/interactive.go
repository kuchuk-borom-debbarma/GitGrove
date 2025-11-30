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
		fmt.Println("6. Add Files")
		fmt.Println("7. Commit Changes")
		fmt.Println("8. Move Repository")
		fmt.Println("9. Exit")
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
			handleAdd(reader, absPath)
		case "7":
			handleCommit(reader, absPath)
		case "8":
			handleMove(reader, absPath)
		case "9":
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
	fmt.Println("Select Child Repository:")
	child, err := selectRepo(reader, rootPath, "Select child")
	if err != nil {
		fmt.Printf("Error selecting child repo: %v\n", err)
		return
	}
	if child == "" {
		return // Cancelled
	}

	fmt.Println("Select Parent Repository:")
	parent, err := selectRepo(reader, rootPath, "Select parent")
	if err != nil {
		fmt.Printf("Error selecting parent repo: %v\n", err)
		return
	}
	if parent == "" {
		return // Cancelled
	}

	if child == parent {
		fmt.Println("Error: Child and Parent cannot be the same.")
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
	repoName, err := selectRepo(reader, rootPath, "Select repository to branch")
	if err != nil {
		fmt.Printf("Error selecting repo: %v\n", err)
		return
	}
	if repoName == "" {
		return
	}

	fmt.Print("Enter New Branch Name: ")
	branch, _ := reader.ReadString('\n')
	branch = strings.TrimSpace(branch)

	if branch == "" {
		fmt.Println("Error: Branch Name is required.")
		return
	}

	if err := core.CreateRepoBranch(rootPath, repoName, branch); err != nil {
		fmt.Printf("Error creating branch: %v\n", err)
	} else {
		fmt.Println("Branch created successfully.")
	}
}

func handleCheckout(reader *bufio.Reader, rootPath string) {
	repoName, err := selectRepo(reader, rootPath, "Select repository to checkout")
	if err != nil {
		fmt.Printf("Error selecting repo: %v\n", err)
		return
	}
	if repoName == "" {
		return
	}

	fmt.Print("Enter Branch Name: ")
	branch, _ := reader.ReadString('\n')
	branch = strings.TrimSpace(branch)

	if branch == "" {
		fmt.Println("Error: Branch Name is required.")
		return
	}

	if err := core.CheckoutRepo(rootPath, repoName, branch); err != nil {
		fmt.Printf("Error checking out branch: %v\n", err)
	} else {
		fmt.Println("Checked out branch successfully.")
	}
}

func handleAdd(reader *bufio.Reader, rootPath string) {
	fmt.Print("Enter file paths to add (comma separated, or '.' for all): ")
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

	if err := core.Add(rootPath, files); err != nil {
		fmt.Printf("Error adding files: %v\n", err)
	} else {
		fmt.Println("Files added successfully.")
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
	repoName, err := selectRepo(reader, rootPath, "Select repository to move")
	if err != nil {
		fmt.Printf("Error selecting repo: %v\n", err)
		return
	}
	if repoName == "" {
		return
	}

	fmt.Print("Enter New Relative Path: ")
	newPath, _ := reader.ReadString('\n')
	newPath = strings.TrimSpace(newPath)

	if newPath == "" {
		fmt.Println("Error: New Path is required.")
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

// selectRepo presents a hierarchical menu to select a repository.
// It returns the selected repo name, or empty string if cancelled.
func selectRepo(reader *bufio.Reader, rootPath, prompt string) (string, error) {
	repos, err := core.GetRepositories(rootPath)
	if err != nil {
		return "", err
	}

	if len(repos) == 0 {
		fmt.Println("No repositories registered.")
		return "", nil
	}

	// Build tree
	// parent -> []children
	tree := make(map[string][]string)
	// roots are repos with empty parent
	var roots []string
	repoMap := make(map[string]core.Repo)

	for _, r := range repos {
		repoMap[r.Name] = r
		if r.Parent == "" {
			roots = append(roots, r.Name)
		} else {
			tree[r.Parent] = append(tree[r.Parent], r.Name)
		}
	}

	// Navigation loop
	currentLevel := roots
	pathStack := []string{} // Stack of parent names

	for {
		fmt.Printf("\n--- %s ---\n", prompt)
		if len(pathStack) > 0 {
			fmt.Printf("Path: %s\n", strings.Join(pathStack, " > "))
		} else {
			fmt.Println("Path: (root)")
		}

		// List options
		for i, name := range currentLevel {
			hasChildren := len(tree[name]) > 0
			suffix := ""
			if hasChildren {
				suffix = " >"
			}
			fmt.Printf("%d. %s%s (%s)\n", i+1, name, suffix, repoMap[name].Path)
		}

		fmt.Println("b. Back / Cancel")
		fmt.Print("Select option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "b" {
			if len(pathStack) > 0 {
				// Go up one level
				pathStack = pathStack[:len(pathStack)-1]

				// Re-calculate current level based on new parent
				if len(pathStack) == 0 {
					currentLevel = roots
				} else {
					// Wait, we need to find siblings of 'parent'
					// Actually, simpler: just re-traverse or store levels in stack?
					// Storing levels is easier but consumes memory.
					// Re-traversing:
					if len(pathStack) == 0 {
						currentLevel = roots
					} else {
						p := pathStack[len(pathStack)-1]
						currentLevel = tree[p]
					}
				}
				continue
			} else {
				return "", nil // Cancel
			}
		}

		var index int
		_, err := fmt.Sscanf(input, "%d", &index)
		if err != nil || index < 1 || index > len(currentLevel) {
			fmt.Println("Invalid selection.")
			continue
		}

		selectedName := currentLevel[index-1]

		// If it has children, ask if user wants to select THIS repo or dive deeper
		if len(tree[selectedName]) > 0 {
			fmt.Printf("Selected '%s'.\n", selectedName)
			fmt.Println("1. Select this repository")
			fmt.Println("2. Navigate into children")
			fmt.Print("Choice: ")
			choice, _ := reader.ReadString('\n')
			choice = strings.TrimSpace(choice)

			if choice == "1" {
				return selectedName, nil
			} else if choice == "2" {
				pathStack = append(pathStack, selectedName)
				currentLevel = tree[selectedName]
				continue
			} else {
				fmt.Println("Invalid choice.")
				continue
			}
		} else {
			// Leaf node, select it
			return selectedName, nil
		}
	}
}

func init() {
	registerCommand(interactiveCommand{})
}
