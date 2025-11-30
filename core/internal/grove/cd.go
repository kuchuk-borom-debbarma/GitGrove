package grove

import "os"

// Cd changes the current repository context.
// Use ".." to navigate to the parent repository.
// Use a repository name to navigate to a child repository.
func Cd(rootAbsPath, target string) error {
	if target == ".." {
		// Navigate up to parent
		return Up(rootAbsPath)
	}
	if target == "~" {
		// Navigate to System Root
		return SwitchToSystem(rootAbsPath)
	}

	// Check if target matches user's home directory (shell expansion of ~)
	homeDir, err := os.UserHomeDir()
	if err == nil && target == homeDir {
		return SwitchToSystem(rootAbsPath)
	}

	// Navigate down to child
	return Down(rootAbsPath, target)
}
