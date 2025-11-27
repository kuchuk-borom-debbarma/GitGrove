package file

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath cleans a path and converts "\" → "/" for consistency.
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	return strings.ReplaceAll(clean, "\\", "/")
}

// Exists returns true if the file/directory exists.
func Exists(path string) bool {
	path = NormalizePath(path)
	_, err := os.Stat(path)
	return err == nil
}

// EnsureNotExists returns error if the path already exists.
func EnsureNotExists(path string) error {
	path = NormalizePath(path)
	if Exists(path) {
		return errors.New("path already exists: " + path)
	}
	return nil
}

// CreateDir creates a directory and its parents, if missing.
func CreateDir(path string) error {
	path = NormalizePath(path)
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.New("failed to create directory " + path + ": " + err.Error())
	}
	return nil
}

// CreateEmptyFile creates a file, also creating parent directories if needed.
func CreateEmptyFile(path string) error {
	path = NormalizePath(path)
	dir := filepath.Dir(path)

	if err := CreateDir(dir); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return errors.New("failed to create file " + path + ": " + err.Error())
	}
	return f.Close()
}

// WriteTextFile writes a text file (creates dirs if needed).
func WriteTextFile(path string, content string) error {
	path = NormalizePath(path)

	if err := CreateDir(filepath.Dir(path)); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// WriteJSONFile writes JSON (pretty-printed) to a file.
func WriteJSONFile(path string, v any) error {
	path = NormalizePath(path)

	if err := CreateDir(filepath.Dir(path)); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errors.New("failed to serialize JSON: " + err.Error())
	}

	return os.WriteFile(path, data, 0644)
}
