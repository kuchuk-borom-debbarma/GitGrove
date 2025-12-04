package groveUtil

import (
	"fmt"
	"os"
	"path/filepath"
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
func CreateGroveConfig(path string) error {
	ggDir := filepath.Join(path, ".gg")
	if err := os.MkdirAll(ggDir, 0755); err != nil {
		return fmt.Errorf("failed to create .gg directory: %w", err)
	}

	configPath := filepath.Join(ggDir, "gg.json")
	// Create an empty or default config file
	if err := os.WriteFile(configPath, []byte("{\n  \"repositories\": {}\n}\n"), 0644); err != nil {
		return fmt.Errorf("failed to create gg.json: %w", err)
	}

	return nil
}
