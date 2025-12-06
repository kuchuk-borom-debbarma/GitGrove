package tui

import (
	"os"
	"path/filepath"
	"strings"
)

// getSuggestions returns a list of subdirectories matching the input path.
// input: relative path entered by user so far.
// basePath: root directory to search from (the repo root).
func getSuggestions(basePath, input string) []string {
	// If input is empty, list root dirs?
	// If input ends with /, list children of that dir.
	// If input does not end with /, list siblings matching prefix.

	var searchDir string
	var prefix string

	// Clean input to handle ./ or multiple slashes?
	// Only valid relative paths inside basePath should be allowed?
	// For simplicity, assume input is clean enough or just string matching.

	if strings.HasSuffix(input, "/") {
		searchDir = filepath.Join(basePath, input)
		prefix = ""
	} else {
		dir := filepath.Dir(input)
		if dir == "." {
			dir = ""
		}
		searchDir = filepath.Join(basePath, dir)
		prefix = filepath.Base(input)
		if prefix == "." { // e.g. input "backend" -> Dir ".", Base "backend"
			prefix = input
			searchDir = basePath
		}
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil
	}

	var suggestions []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") { // Ignore hidden dirs like .git, .gg
			if strings.HasPrefix(entry.Name(), prefix) {
				// We want to return the full relative path to append to input?
				// Or just the completion?
				// User wants to tab complete.
				// If input is "back", suggestion "backend".
				// If input is "backend/", suggestion "backend/serviceA".

				// Let's return the full path relative to basePath so we can just replace textInput value.

				var suggestion string
				if strings.HasSuffix(input, "/") {
					suggestion = input + entry.Name()
				} else {
					// Replace base with full name
					dir := filepath.Dir(input)
					if dir == "." {
						suggestion = entry.Name()
					} else {
						suggestion = filepath.Join(dir, entry.Name())
					}
				}
				// Append slash for convenience?
				suggestion += "/"
				suggestions = append(suggestions, suggestion)
			}
		}
	}
	return suggestions
}
