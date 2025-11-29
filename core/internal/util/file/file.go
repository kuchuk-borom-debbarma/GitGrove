package file

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NormalizePath cleans a path and converts "\" → "/".
//
// It uses filepath.Clean to resolve dot-dots and remove extraneous separators,
// then forces forward slashes to ensure consistency across platforms (Windows/Linux/macOS).
// This is crucial for GitGroove metadata which should be platform-agnostic.
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	return strings.ReplaceAll(clean, "\\", "/")
}

// Exists checks if a file or directory exists at the given path.
//
// It returns true if the path exists (regardless of type), and false otherwise.
// Note: It returns false for any error (e.g., permission denied), not just "not found".
func Exists(path string) bool {
	path = NormalizePath(path)
	_, err := os.Stat(path)
	return err == nil
}

// EnsureNotExists verifies that the given path does not exist.
//
// It returns an error if the path exists, or nil if it does not.
// Used for idempotency checks (e.g., ensuring we don't overwrite an existing .gg).
func EnsureNotExists(path string) error {
	path = NormalizePath(path)
	if Exists(path) {
		return errors.New("path already exists: " + path)
	}
	return nil
}

// CreateDir creates a directory and all necessary parents (mkdir -p).
//
// It uses permission 0755 (rwxr-xr-x).
// Returns an error if creation fails.
func CreateDir(path string) error {
	path = NormalizePath(path)
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.New("failed to create directory " + path + ": " + err.Error())
	}
	return nil
}

// CreateEmptyFile creates an empty file at the specified path.
//
// It automatically creates any missing parent directories.
// Returns an error if directory creation or file creation fails.
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

// WriteTextFile writes a string content to a file.
//
// It automatically creates any missing parent directories.
// The file is written with permission 0644 (rw-r--r--).
// If the file exists, it is overwritten.
func WriteTextFile(path string, content string) error {
	path = NormalizePath(path)
	if err := CreateDir(filepath.Dir(path)); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// WriteJSONFile serializes a Go value to a JSON file with indentation.
//
// It automatically creates any missing parent directories.
// The JSON is marshaled with 2-space indentation for readability.
// The file is written with permission 0644.
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

func ReadTextFile(path string) (string, error) {
	path = NormalizePath(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func AppendTextFile(path, content string) error {
	path = NormalizePath(path)
	if err := CreateDir(filepath.Dir(path)); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return nil
}

// RandomString generates a random alphanumeric string of length n.
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
