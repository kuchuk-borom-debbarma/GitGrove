package grove

// Cd changes the current repository context.
// Use ".." to navigate to the parent repository.
// Use a repository name to navigate to a child repository.
func Cd(rootAbsPath, target string) error {
	if target == ".." {
		// Navigate up to parent
		return Up(rootAbsPath)
	}
	// Navigate down to child
	return Down(rootAbsPath, target)
}
