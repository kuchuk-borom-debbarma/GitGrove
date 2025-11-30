package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// FindRepoRoot returns the absolute path to the root of the git repository.
func FindRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find git root (are you in a git repo?): %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
