package status

import (
	"fmt"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/core/internal/util/git"
)

type BasicStatus struct {
	RootPath      string
	CurrentBranch string
	IsClean       bool
	SystemCommit  string
}

func GetBasicStatus(rootAbsPath string) (*BasicStatus, error) {
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

	systemRef := "refs/heads/gitgroove/system"
	systemCommit, err := gitUtil.ResolveRef(rootAbsPath, systemRef)
	if err != nil {
		systemCommit = "not initialized"
	}

	return &BasicStatus{
		RootPath:      rootAbsPath,
		CurrentBranch: branch,
		IsClean:       isClean,
		SystemCommit:  systemCommit,
	}, nil
}
