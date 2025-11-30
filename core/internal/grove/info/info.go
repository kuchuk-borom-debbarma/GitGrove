package info

import (
	"fmt"
	"strings"
)

type Info struct {
	Basic *BasicInfo
	Repos *RepoInfo
	Links *LinkInfo
}

func GetInfo(rootAbsPath string) (*Info, error) {
	basic, err := GetBasicInfo(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic info: %w", err)
	}

	repos, err := GetRepoInfo(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo info: %w", err)
	}

	links := GetLinkInfo(repos)

	return &Info{
		Basic: basic,
		Repos: repos,
		Links: links,
	}, nil
}

func (d *Info) String() string {
	var sb strings.Builder

	sb.WriteString("GitGrove Info\n")
	sb.WriteString("=============\n\n")

	sb.WriteString(fmt.Sprintf("Root:   %s\n", d.Basic.RootPath))
	sb.WriteString(fmt.Sprintf("Branch: %s\n", d.Basic.CurrentBranch))
	cleanState := "Clean"
	if !d.Basic.IsClean {
		cleanState = "Dirty"
	}
	sb.WriteString(fmt.Sprintf("State:  %s\n", cleanState))
	sb.WriteString(fmt.Sprintf("System: %s\n\n", d.Basic.SystemCommit))

	sb.WriteString("Registered Repositories:\n")
	sb.WriteString("------------------------\n")
	if len(d.Repos.Repos) == 0 {
		sb.WriteString("(none)\n")
	} else {
		sb.WriteString(d.Links.String())
	}

	return sb.String()
}
