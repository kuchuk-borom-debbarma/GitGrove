package groveUtil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitUtil "github.com/kuchuk-borom-debbarma/GitGrove/src/internal/util/git"
	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

// IsGroveInitialized checks if the .gg directory and configuration file exist.
func IsGroveInitialized(path string) error {
	// 1. Check local file system
	configPath := filepath.Join(path, ".gg", "gg.json")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("gitgrove is already initialized in %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking grove initialization: %w", err)
	}

	// 2. Check if we are in an orphan branch (gg/<trunk>/<repoName>)
	// If so, check if .gg/gg.json exists in <trunk>
	if err := gitUtil.IsGitRepository(path); err != nil {
		// Not a git repo, so definitely not initialized
		return nil
	}

	currentBranch, err := gitUtil.CurrentBranch(path)
	if err != nil {
		// Ignore error, maybe no commits yet
		return nil
	}

	// Pattern: gg/<trunk>/<repoName>
	parts := strings.Split(currentBranch, "/")
	if len(parts) >= 3 && parts[0] == "gg" {
		// Assuming trunk is the second part.
		// NOTE: Trunk name might contain slashes? For now assuming no slashes in trunk or repo name,
		// or strict mapping. The standard is gg/<trunk>/<repoName>.
		// If trunk has slashes, this split might satisfy len >= 3 but be ambiguous.
		// However, based on our generation logic: fmt.Sprintf("gg/%s/%s", currentBranch, repo.Name)
		// We can try to reconstruct trunk.
		// But wait, if trunk has slashes `feature/x`, then `gg/feature/x/repo`.
		// Then parts will be ["gg", "feature", "x", "repo"].
		// We need to know where trunk ends.
		// Without knowing repo name, it is hard.
		// BUT, we just need to find *A* branch that has .gg/gg.json.
		// Let's assume the trunk logic is simple for now or try to match available branches.
		// Simpler approach: check if parts[1] is a valid ref that has .gg/gg.json?
		// Or assume no slashes in trunk for MVP or accept limitation.
		// Let's assume trunk might correspond to everything between gg/ and /<last_part>.
		// Reconstruct trunk candidate?
		// Actually, if we are in `gg/master/serviceA`, trunk is `master`.
		// If `gg/feature/x/serviceA`, trunk is `feature/x`.
		// We can iterate over possible split points?
		// Given `gg/A/B/C`, trunk could be `A`, `A/B`.

		// Let's try to detect if we can find .gg/gg.json in any prefix combination.
		// Start from parts[1] (index 1). End at len(parts)-2 (inclusive).
		// Because last part is repoName.
		for i := 1; i < len(parts)-1; i++ {
			trunkCandidate := strings.Join(parts[1:i+1], "/")
			exists, err := gitUtil.FileExistsInBranch(path, trunkCandidate, ".gg/gg.json")
			if err == nil && exists {
				return fmt.Errorf("gitgrove is already initialized (orphan branch of %s)", trunkCandidate)
			}
		}
	}

	return nil
}

// CreateGroveConfig creates the .gg directory and the gg.json file.
func CreateGroveConfig(path string, repoAwareContextMessage bool) error {
	ggDir := filepath.Join(path, ".gg")
	if err := os.MkdirAll(ggDir, 0755); err != nil {
		return fmt.Errorf("failed to create .gg directory: %w", err)
	}

	configPath := filepath.Join(ggDir, "gg.json")

	config := GGConfig{
		Repositories:            make(map[string]model.GGRepo),
		RepoAwareContextMessage: repoAwareContextMessage,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal gg.json: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create gg.json: %w", err)
	}

	return nil
}

// GGConfig represents the structure of gg.json
type GGConfig struct {
	Repositories            map[string]model.GGRepo `json:"repositories"`
	RepoAwareContextMessage bool                    `json:"repo_aware_context_message"`
}

// LoadConfig reads the gg.json configuration from the .gg directory.
func LoadConfig(ggRootPath string) (*GGConfig, error) {
	configPath := filepath.Join(ggRootPath, ".gg", "gg.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gg.json: %w", err)
	}

	var config GGConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse gg.json: %w", err)
	}

	if config.Repositories == nil {
		config.Repositories = make(map[string]model.GGRepo)
	}

	return &config, nil
}

// LoadConfigFromGitRef reads the gg.json configuration from a specific git reference (branch/commit).
func LoadConfigFromGitRef(ggRootPath string, ref string) (*GGConfig, error) {
	configPath := ".gg/gg.json"
	data, err := gitUtil.ReadFileFromBranch(ggRootPath, ref, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gg.json from ref %s: %w", ref, err)
	}

	var config GGConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse gg.json from ref %s: %w", ref, err)
	}

	if config.Repositories == nil {
		config.Repositories = make(map[string]model.GGRepo)
	}

	return &config, nil
}

// RegisterRepoInConfig adds new repositories to the gg.json configuration.
// It performs validation to ensure no name/path conflicts or nested repositories.
func RegisterRepoInConfig(ggRootPath string, newRepos []model.GGRepo) error {
	// Read existing config
	config, err := LoadConfig(ggRootPath)
	if err != nil {
		return err
	}

	if err := ValidateRepoRegistration(ggRootPath, config, newRepos); err != nil {
		return err
	}

	return AddReposToConfig(ggRootPath, newRepos)
}

// ValidateRepoRegistration checks if the new repos can be safely added to the config.
func ValidateRepoRegistration(ggRootPath string, config *GGConfig, newRepos []model.GGRepo) error {
	for _, newRepo := range newRepos {
		// Normalize path (note: we do this check on a copy or assuming caller cleans it?
		// RegisterRepoInConfig does verify, but caller of Validate might modify newRepo.Path in place?
		// Strings are immutable, struct fields are not.
		// Let's clean it here for validation purposes.
		cleanedPath := filepath.Clean(newRepo.Path)

		// Validation: Check if path is within root
		absPath := filepath.Join(ggRootPath, cleanedPath)
		absPath = filepath.Clean(absPath)

		relCheck, err := filepath.Rel(ggRootPath, absPath)
		if err != nil {
			return fmt.Errorf("invalid path '%s': %w", newRepo.Path, err)
		}
		if strings.HasPrefix(relCheck, "..") {
			return fmt.Errorf("path '%s' must be within repository root", newRepo.Path)
		}

		// Check for name conflict
		if _, exists := config.Repositories[newRepo.Name]; exists {
			return fmt.Errorf("repository with name '%s' already exists", newRepo.Name)
		}

		// Check for path conflict and nested repositories
		for _, existingRepo := range config.Repositories {
			if existingRepo.Path == cleanedPath {
				return fmt.Errorf("repository with path '%s' already exists (name: %s)", cleanedPath, existingRepo.Name)
			}

			// Check for nesting
			rel, err := filepath.Rel(existingRepo.Path, cleanedPath)
			if err == nil && !strings.HasPrefix(rel, "..") {
				return fmt.Errorf("cannot register '%s' inside existing repo '%s'", cleanedPath, existingRepo.Path)
			}

			rel, err = filepath.Rel(cleanedPath, existingRepo.Path)
			if err == nil && !strings.HasPrefix(rel, "..") {
				return fmt.Errorf("cannot register '%s' which contains existing repo '%s'", cleanedPath, existingRepo.Path)
			}
		}
	}
	return nil
}

// AddReposToConfig adds the repositories to gg.json without further validation (assumes validation passed).
func AddReposToConfig(ggRootPath string, newRepos []model.GGRepo) error {
	configPath := filepath.Join(ggRootPath, ".gg", "gg.json")

	// Read existing config
	config, err := LoadConfig(ggRootPath)
	if err != nil {
		return err
	}

	for _, newRepo := range newRepos {
		// Ensure path is cleaned before saving
		newRepo.Path = filepath.Clean(newRepo.Path)
		config.Repositories[newRepo.Name] = newRepo
	}

	// Write updated config
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write gg.json: %w", err)
	}

	return nil
}

// SetContextRepo sets the gitgrove.context.repo config to the specified repository name.
func SetContextRepo(ggRepoPath string, repoName string) error {
	return gitUtil.SetLocalConfig(ggRepoPath, "gitgrove.context.repo", repoName)
}

// GetContextRepo gets the repository name from gitgrove.context.repo config.
func GetContextRepo(ggRepoPath string) (string, error) {
	return gitUtil.GetLocalConfig(ggRepoPath, "gitgrove.context.repo")
}

// ClearContextRepo removes the gitgrove.context.repo config.
func ClearContextRepo(ggRepoPath string) error {
	return gitUtil.UnsetLocalConfig(ggRepoPath, "gitgrove.context.repo")
}

// SetContextTrunk sets the gitgrove.context.trunk config to the specified branch name.
func SetContextTrunk(ggRepoPath string, trunkName string) error {
	return gitUtil.SetLocalConfig(ggRepoPath, "gitgrove.context.trunk", trunkName)
}

// GetContextTrunk gets the trunk branch name from gitgrove.context.trunk config.
func GetContextTrunk(ggRepoPath string) (string, error) {
	return gitUtil.GetLocalConfig(ggRepoPath, "gitgrove.context.trunk")
}

// ClearContextTrunk removes the gitgrove.context.trunk config.
func ClearContextTrunk(ggRepoPath string) error {
	return gitUtil.UnsetLocalConfig(ggRepoPath, "gitgrove.context.trunk")
}
