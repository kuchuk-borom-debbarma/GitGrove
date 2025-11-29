package doctor

import (
	"fmt"
	"strings"
)

type Doctor struct {
	Basic *BasicDoctor
	Repos *RepoDoctor
	Links *LinkDoctor
}

func GetDoctor(rootAbsPath string) (*Doctor, error) {
	basic, err := GetBasicDoctor(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic doctor: %w", err)
	}

	repos, err := GetRepoDoctor(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo doctor: %w", err)
	}

	links := GetLinkDoctor(repos)

	return &Doctor{
		Basic: basic,
		Repos: repos,
		Links: links,
	}, nil
}

func (d *Doctor) String() string {
	var sb strings.Builder

	sb.WriteString("GitGrove Doctor\n")
	sb.WriteString("===============\n\n")

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
