package groveUtil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/src/model"
)

// IsGroveInitialized checks if the .gg directory and configuration file exist.
func IsGroveInitialized(path string) error {
	configPath := filepath.Join(path, ".gg", "gg.json")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("gitgrove is already initialized in %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking grove initialization: %w", err)
	}
	return nil
}

// CreateGroveConfig creates the .gg directory and the gg.json file.
func CreateGroveConfig(path string, atomicCommit bool) error {
	ggDir := filepath.Join(path, ".gg")
	if err := os.MkdirAll(ggDir, 0755); err != nil {
		return fmt.Errorf("failed to create .gg directory: %w", err)
	}

	configPath := filepath.Join(ggDir, "gg.json")

	config := GGConfig{
		Repositories: make(map[string]model.GGRepo),
		AtomicCommit: atomicCommit,
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
	Repositories map[string]model.GGRepo `json:"repositories"`
	AtomicCommit bool                    `json:"atomic_commit"`
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

// RegisterRepoInConfig adds new repositories to the gg.json configuration.
// It performs validation to ensure no name/path conflicts or nested repositories.
func RegisterRepoInConfig(ggRootPath string, newRepos []model.GGRepo) error {
	configPath := filepath.Join(ggRootPath, ".gg", "gg.json")

	// Read existing config
	config, err := LoadConfig(ggRootPath)
	if err != nil {
		return err
	}

	// Validate and add new repos
	for _, newRepo := range newRepos {
		// Check for name conflict
		if _, exists := config.Repositories[newRepo.Name]; exists {
			return fmt.Errorf("repository with name '%s' already exists", newRepo.Name)
		}

		// Check for path conflict and nested repositories
		for _, existingRepo := range config.Repositories {
			if existingRepo.Path == newRepo.Path {
				return fmt.Errorf("repository with path '%s' already exists (name: %s)", newRepo.Path, existingRepo.Name)
			}

			// Check for nesting
			rel, err := filepath.Rel(existingRepo.Path, newRepo.Path)
			if err == nil && !strings.HasPrefix(rel, "..") {
				return fmt.Errorf("cannot register '%s' inside existing repo '%s'", newRepo.Path, existingRepo.Path)
			}

			rel, err = filepath.Rel(newRepo.Path, existingRepo.Path)
			if err == nil && !strings.HasPrefix(rel, "..") {
				return fmt.Errorf("cannot register '%s' which contains existing repo '%s'", newRepo.Path, existingRepo.Path)
			}
		}

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
