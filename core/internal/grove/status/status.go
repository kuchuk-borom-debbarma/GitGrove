package status

import (
	"fmt"
	"strings"
)

type Status struct {
	Basic *BasicStatus
	Repos *RepoStatus
	Links *LinkStatus
}

func GetStatus(rootAbsPath string) (*Status, error) {
	basic, err := GetBasicStatus(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic status: %w", err)
	}

	repos, err := GetRepoStatus(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo status: %w", err)
	}

	links := GetLinkStatus(repos)

	return &Status{
		Basic: basic,
		Repos: repos,
		Links: links,
	}, nil
}

func (s *Status) String() string {
	var sb strings.Builder

	sb.WriteString("GitGrove Status\n")
	sb.WriteString("===============\n\n")

	sb.WriteString(fmt.Sprintf("Root:   %s\n", s.Basic.RootPath))
	sb.WriteString(fmt.Sprintf("Branch: %s\n", s.Basic.CurrentBranch))
	cleanState := "Clean"
	if !s.Basic.IsClean {
		cleanState = "Dirty"
	}
	sb.WriteString(fmt.Sprintf("State:  %s\n", cleanState))
	sb.WriteString(fmt.Sprintf("System: %s\n\n", s.Basic.SystemCommit))

	sb.WriteString("Registered Repositories:\n")
	sb.WriteString("------------------------\n")
	if len(s.Repos.Repos) == 0 {
		sb.WriteString("(none)\n")
	} else {
		sb.WriteString(s.Links.String())
	}

	return sb.String()
}
