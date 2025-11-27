package file

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath cleans a path and converts "\" → "/".
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	return strings.ReplaceAll(clean, "\\", "/")
}

func Exists(path string) bool {
	path = NormalizePath(path)
	_, err := os.Stat(path)
	return err == nil
}

func EnsureNotExists(path string) error {
	path = NormalizePath(path)
	if Exists(path) {
		return errors.New("path already exists: " + path)
	}
	return nil
}

func CreateDir(path string) error {
	path = NormalizePath(path)
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.New("failed to create directory " + path + ": " + err.Error())
	}
	return nil
}

func CreateEmptyFile(path string) error {
	path = NormalizePath(path)
	if err := CreateDir(filepath.Dir(path)); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return errors.New("failed to create file " + path + ": " + err.Error())
	}
	return f.Close()
}

func WriteTextFile(path string, content string) error {
	path = NormalizePath(path)
	if err := CreateDir(filepath.Dir(path)); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

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
