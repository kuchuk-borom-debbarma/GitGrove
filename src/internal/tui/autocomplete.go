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
		if filepath.IsAbs(input) {
			searchDir = input
		} else {
			searchDir = filepath.Join(basePath, input)
		}
		prefix = ""
	} else {
		dir := filepath.Dir(input)
		if dir == "." && !filepath.IsAbs(input) {
			dir = ""
		}

		if filepath.IsAbs(input) {
			searchDir = dir
		} else {
			searchDir = filepath.Join(basePath, dir)
		}

		prefix = filepath.Base(input)
		if prefix == "." { // e.g. input "backend" -> Dir ".", Base "backend"
			prefix = input
			if filepath.IsAbs(input) {
				searchDir = input // Should be dir? filepath.Dir(input)
				// Wait, if input is "/Users", Dir is "/", Base is "Users".
				// searchDir should be "/".
				// IF input is "/Users/foo", Dir is "/Users", Base "foo".

				// Re-evaluating prefix logic for absolute paths.
				// filepath.Dir("/Users") -> "/"
				// filepath.Base("/Users") -> "Users"
				// logic above: "dir := filepath.Dir(input)".
				// if input="/Users", dir="/".
				// searchDir = "/"
				// prefix = "Users"
				// Correct.
			} else {
				// Revert to relative logic
				if prefix == "." {
					// wait, filepath.Base("backend") -> "backend".
					// previous logic: if prefix == "." { prefix = input; searchDir = basePath }
					// This was for "." case? Or "backend" case?
					// filepath.Dir("backend") -> "."
					// filepath.Base("backend") -> "backend".
					// So prefix != ".".
					// When does prefix == "."?
					// filepath.Base(".") -> "."
					// If input is ".", Dir is ".", Base is ".".
				}
				// Original logic:
				// if prefix == "." { ... }
				// was handling likely "." input specifically.
				// Let's keep it robust.
				searchDir = basePath // Default fallback for relative
			}
		}

		// Clean logic for both:
		// 1. Determine directory portion (searchDir)
		// 2. Determine file portion (prefix)

		if filepath.IsAbs(input) {
			dir := filepath.Dir(input)
			prefix = filepath.Base(input)
			// Handle case where input is exactly "/"
			if input == "/" {
				dir = "/"
				prefix = ""
			}
			searchDir = dir
		} else {
			// Relative logic
			dir := filepath.Dir(input)
			prefix = filepath.Base(input)

			if dir == "." {
				dir = ""
			}
			searchDir = filepath.Join(basePath, dir)

			// Special handling for clean names without slashes
			if !strings.Contains(input, string(filepath.Separator)) && input != "" {
				searchDir = basePath
				prefix = input
			}
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
