package grove

import (
	"fmt"
	"strings"
)

// ParseRepoFromBranch extracts the target repository name from a GitGrove branch name.
//
// Expected format: gitgroove/repos/<hierarchy>/branches/<branchName>
// Hierarchy is a sequence of repo names separated by "children".
// Example: gitgroove/repos/root/children/backend/branches/main -> backend
func ParseRepoFromBranch(branchName string) (string, error) {
	prefix := "gitgroove/repos/"
	if !strings.HasPrefix(branchName, prefix) {
		return "", fmt.Errorf("not on a valid GitGrove repository branch (current: %s)", branchName)
	}

	// Remove prefix
	trimmed := strings.TrimPrefix(branchName, prefix)

	// Find the hierarchy part.
	// The hierarchy ends where "/branches/" begins.
	// Since repo names cannot contain "/", the hierarchy is a sequence of names separated by "/children/".

	// Split by "/"
	segments := strings.Split(trimmed, "/")

	// We expect: [repo1, "children", repo2, "children", ..., "branches", branchName...]
	// We need to find the "branches" segment that terminates the hierarchy.

	var repoNames []string
	var branchNameIdx int = -1

	// State machine: 0=ExpectRepoName, 1=ExpectChildrenOrBranches
	state := 0
	for i, seg := range segments {
		if state == 0 {
			// Expect repo name
			repoNames = append(repoNames, seg)
			state = 1
		} else {
			// Expect "children" or "branches"
			if seg == "children" {
				state = 0 // Next is repo name
			} else if seg == "branches" {
				branchNameIdx = i + 1
				break
			} else {
				// Invalid structure
				return "", fmt.Errorf("invalid branch format at segment '%s': expected 'children' or 'branches'", seg)
			}
		}
	}

	if branchNameIdx == -1 || branchNameIdx >= len(segments) {
		return "", fmt.Errorf("invalid GitGrove branch format: missing branch name")
	}

	// The target repo is the last one in the hierarchy
	if len(repoNames) == 0 {
		return "", fmt.Errorf("invalid GitGrove branch format: no repo found")
	}

	return repoNames[len(repoNames)-1], nil
}
