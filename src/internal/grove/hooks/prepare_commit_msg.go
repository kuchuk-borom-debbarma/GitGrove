package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	groveUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/grove"
)

// PrepareCommitMsg modifies the commit message to prepend [RepoName] if strict context is established.
func PrepareCommitMsg(msgFile, source, sha string) error {
	// Only run on normal commits (skip rebase, merge, amend if preferred, but usually we want it on message creation)
	// source can be: message, template, merge, squash, commit (when -c/-C/-F is used).
	// We mainly want to catch when user types a new message or uses -m.
	// If source is empty, it's usually `git commit` (no -m) opening editor?
	// Actually git docs say:
	// - message (if -m or -F)
	// - template (if -t)
	// - merge (if merging)
	// - squash (if squashing)
	// - commit (if -c, -C, --amend)
	// If source is missing, it's standard commit.

	// If it's a merge, we probably shouldn't mess with it? Or maybe we should?
	// Let's stick to standard commits for now.

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get cwd: %w", err)
	}

	config, err := groveUtil.LoadConfig(cwd)
	if err != nil {
		// Not a grove repo or error loading config, skip
		// fmt.Printf("Debug: Error loading config: %v\n", err)
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			return nil
		}
		// Double check existence to be sure
		if _, statErr := os.Stat(filepath.Join(cwd, ".gg", "gg.json")); os.IsNotExist(statErr) {
			return nil
		}
		return err
	}

	// fmt.Printf("Debug: Config loaded. AtomicCommit: %v\n", config.AtomicCommit)

	if !config.RepoAwareContextMessage {
		return nil
	}

	// Logic from pre_commit.go to find affected repos
	stagedFiles, err := gitUtil.GetStagedFiles(cwd)
	// fmt.Printf("Debug: Staged files: %v\n", stagedFiles)
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(stagedFiles) == 0 {
		return nil
	}

	affectedRepos := make(map[string]bool)
	affectedRoot := false

	for _, file := range stagedFiles {
		matched := false
		for _, repo := range config.Repositories {
			relPath, err := filepath.Rel(repo.Path, file)
			if err == nil && !strings.HasPrefix(relPath, "..") {
				affectedRepos[repo.Name] = true
				matched = true
				break
			}
		}
		if !matched {
			affectedRoot = true
		}
	}

	// Logic:
	// If 1 repo affected AND no root files affected -> Prepend [RepoName]
	if len(affectedRepos) == 1 && !affectedRoot {
		var repoName string
		for r := range affectedRepos {
			repoName = r
		}

		// Read existing message
		msgBytes, err := os.ReadFile(msgFile)
		if err != nil {
			return fmt.Errorf("failed to read commit message file: %w", err)
		}
		msg := string(msgBytes)

		prefix := fmt.Sprintf("[%s] ", repoName)
		if !strings.HasPrefix(msg, prefix) {
			newMsg := prefix + msg
			if err := os.WriteFile(msgFile, []byte(newMsg), 0644); err != nil {
				return fmt.Errorf("failed to write commit message file: %w", err)
			}
		}
	}

	return nil
}
