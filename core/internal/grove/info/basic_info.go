package info

import (
	"fmt"
	"strings"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

type BasicInfo struct {
	RootPath      string
	CurrentBranch string
	IsClean       bool
	InternalCommit  string
}

func GetBasicInfo(rootAbsPath string) (*BasicInfo, error) {
	if !gitUtil.IsInsideGitRepo(rootAbsPath) {
		return nil, fmt.Errorf("not a git repository: %s", rootAbsPath)
	}

	branch, err := gitUtil.GetCurrentBranch(rootAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	isClean := true
	if err := gitUtil.VerifyCleanState(rootAbsPath); err != nil {
		isClean = false
	}

	internalRef := "refs/heads/gitgroove/internal"
	systemCommit, err := gitUtil.ResolveRef(rootAbsPath, internalRef)
	if err != nil {
		systemCommit = "not initialized"
	}

	return &BasicInfo{
		RootPath:      rootAbsPath,
		CurrentBranch: strings.TrimSpace(branch), // Changed from 'branch' to 'strings.TrimSpace(branch)' and fixed variable name
		IsClean:       isClean,
		InternalCommit:  systemCommit,
	}, nil
}
