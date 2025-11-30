package grove

import (
	"fmt"
	"strings"
)

const branchPrefix = "refs/heads/gitgroove/repos/"

// RepoBranchRef constructs the full ref for a repo branch.
// Format: refs/heads/gitgroove/repos/<repoName>/branches/<branchName>
func RepoBranchRef(repoName, branchName string) string {
	return fmt.Sprintf("%s%s/branches/%s", branchPrefix, repoName, branchName)
}

// RepoBranchShortFromRef returns the short branch name (without refs/heads/) from a full ref.
func RepoBranchShortFromRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

// ParseRepoFromBranch extracts the repo name from a GitGrove branch name (short or full).
func ParseRepoFromBranch(branchName string) (string, error) {
	// Accept short or full ref. Normalize.
	short := strings.TrimPrefix(branchName, "refs/heads/")
	if !strings.HasPrefix(short, "gitgroove/repos/") {
		return "", fmt.Errorf("not a valid GitGrove repo branch: %s", branchName)
	}
	// short form: gitgroove/repos/<repoName>/branches/<branch>
	// Split by "/"
	// [gitgroove, repos, <repoName>, branches, <branchName>...]
	parts := strings.Split(short, "/")
	if len(parts) < 5 || parts[3] != "branches" {
		return "", fmt.Errorf("invalid GitGrove branch format: %s", branchName)
	}
	return parts[2], nil
}
