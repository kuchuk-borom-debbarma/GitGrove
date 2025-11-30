#!/bin/bash

# Script to aggregate all code files from specified folders into a single file
# Usage: ./aggregate_code.sh [output_file] [folder1] [folder2] [folder3] ...

# Default values
OUTPUT_FILE="${1:-codebase_dump.txt}"
shift  # Remove first argument (output file) from $@
CLI_FOLDERS=("$@")  # Folders specified via command line

# =============================================================================
# CONFIGURATION: Edit these arrays to set default folders and files to scan
# =============================================================================

# Default folders to scan (relative paths from where script is run)
# These will be used if no folders are specified on command line
DEFAULT_FOLDERS=(
    "core"
    # Add your project-specific folders here
    # "frontend/src"
    # "backend/api"
    # "shared"
)

# Specific files to always include when found within the scanned folders
# (these will be included even if they don't match code file extensions)
INCLUDE_FILES=(
    "README.md"
    "package.json"
    "requirements.txt"
    "Cargo.toml"
    "go.mod"
    "pom.xml"
    "build.gradle"
    "CMakeLists.txt"
    "Makefile"
    "Dockerfile"
    ".env.example"
    # Add your specific files here
    # "config/database.yml"
    # "docs/api.md"
)

# =============================================================================

# Determine which folders to scan
if [ ${#CLI_FOLDERS[@]} -gt 0 ]; then
    SCAN_FOLDERS=("${CLI_FOLDERS[@]}")
    echo "Using command-line specified folders"
else
    SCAN_FOLDERS=("${DEFAULT_FOLDERS[@]}")
    echo "Using default configured folders"
    echo "Tip: You can override by specifying folders: $0 [output_file] folder1 folder2 ..."
fi

# Common code file extensions
CODE_EXTENSIONS=(
    "*.py"      # Python
    "*.js"      # JavaScript
    "*.ts"      # TypeScript
    "*.jsx"     # React JSX
    "*.tsx"     # React TSX
    "*.java"    # Java
    "*.c"       # C
    "*.cpp"     # C++
    "*.cc"      # C++
    "*.cxx"     # C++
    "*.h"       # Header files
    "*.hpp"     # C++ headers
    "*.cs"      # C#
    "*.php"     # PHP
    "*.rb"      # Ruby
    "*.go"      # Go
    "*.rs"      # Rust
    "*.swift"   # Swift
    "*.kt"      # Kotlin
    "*.scala"   # Scala
    "*.sh"      # Shell scripts
    "*.bash"    # Bash scripts
    "*.zsh"     # Zsh scripts
    "*.fish"    # Fish scripts
    "*.ps1"     # PowerShell
    "*.r"       # R
    "*.R"       # R
    "*.m"       # Objective-C/MATLAB
    "*.mm"      # Objective-C++
    "*.pl"      # Perl
    "*.pm"      # Perl modules
    "*.lua"     # Lua
    "*.vim"     # Vim script
    "*.sql"     # SQL
    "*.html"    # HTML
    "*.htm"     # HTML
    "*.css"     # CSS
    "*.scss"    # Sass
    "*.sass"    # Sass
    "*.less"    # Less CSS
    "*.xml"     # XML
    "*.json"    # JSON
    "*.yaml"    # YAML
    "*.yml"     # YAML
    "*.toml"    # TOML
    "*.ini"     # INI files
    "*.conf"    # Configuration files
    "*.config"  # Configuration files
    "*.md"      # Markdown
    "*.txt"     # Text files
    "*.dockerfile" # Dockerfile
    "Dockerfile"   # Dockerfile (no extension)
    "Makefile"     # Makefile
    "*.mk"      # Makefile
    "*.cmake"   # CMake
    "*.gradle"  # Gradle
    "*.sbt"     # SBT
    "*.clj"     # Clojure
    "*.cljs"    # ClojureScript
    "*.ex"      # Elixir
    "*.exs"     # Elixir
    "*.erl"     # Erlang
    "*.hrl"     # Erlang headers
    "*.dart"    # Dart
    "*.f90"     # Fortran
    "*.f95"     # Fortran
    "*.asm"     # Assembly
    "*.s"       # Assembly
)

# Function to check if a folder exists and is accessible
check_folder() {
    local folder="$1"
    if [ ! -d "$folder" ]; then
        echo "Warning: Folder '$folder' does not exist, skipping..."
        return 1
    fi
    if [ ! -r "$folder" ]; then
        echo "Warning: Folder '$folder' is not readable, skipping..."
        return 1
    fi
    return 0
}

# Function to check if file is a code file
is_code_file() {
    local file="$1"
    local filename=$(basename "$file")

    # Check for exact matches first (like Dockerfile, Makefile)
    for ext in "${CODE_EXTENSIONS[@]}"; do
        if [[ "$filename" == "$ext" ]]; then
            return 0
        fi
    done

    # Check for pattern matches
    for ext in "${CODE_EXTENSIONS[@]}"; do
        if [[ "$filename" == $ext ]]; then
            return 0
        fi
    done

    return 1
}

# Function to check if file is binary
is_binary_file() {
    if file "$1" | grep -q "text"; then
        return 1  # Not binary
    else
        return 0  # Binary
    fi
}

echo "Starting code aggregation..."
echo "Folders to scan: ${SCAN_FOLDERS[*]}"
echo "Output file: $(realpath "$OUTPUT_FILE")"
echo "----------------------------------------"

# Clear the output file and add header
> "$OUTPUT_FILE"
echo "# Code Aggregation Output" >> "$OUTPUT_FILE"
echo "# Generated on: $(date)" >> "$OUTPUT_FILE"
echo "# Scanned folders: ${SCAN_FOLDERS[*]}" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Counter for processed files
file_count=0

# Process each specified folder
for folder in "${SCAN_FOLDERS[@]}"; do
    echo "Scanning folder: $folder"

    # Check if folder exists and is accessible
    if ! check_folder "$folder"; then
        continue
    fi

    # Find and process all code files in this folder
    while IFS= read -r -d '' file; do
        # Check if it's a code file or in the include files list
        filename=$(basename "$file")
        is_include_file=false

        # Check if this file matches any of the include files (just filename)
        for include_file in "${INCLUDE_FILES[@]}"; do
            include_filename=$(basename "$include_file")
            if [[ "$filename" == "$include_filename" ]]; then
                is_include_file=true
                break
            fi
        done

        if is_code_file "$file" || [[ "$is_include_file" == true ]]; then
            # Skip binary files
            if is_binary_file "$file"; then
                echo "Skipping binary file: $file"
                continue
            fi

            # Get relative path from current directory (where script is run)
            # Use a more reliable method to get relative path
            if command -v realpath >/dev/null 2>&1; then
                current_dir=$(realpath ".")
                file_path=$(realpath "$file")
                rel_path="${file_path#$current_dir/}"
                # If the file is in current directory, just use basename
                if [[ "$rel_path" == "$file_path" ]]; then
                    rel_path=$(basename "$file")
                fi
            else
                # Fallback: remove leading ./ if present, otherwise use as-is
                rel_path="${file#./}"
            fi

            # Safety check: if rel_path is empty, use the original file path
            if [[ -z "$rel_path" ]]; then
                rel_path="$file"
            fi

            echo "Processing: $rel_path"

            # Write the file header with clear formatting
            echo "================================================================================================" >> "$OUTPUT_FILE"
            echo "FILE: $rel_path" >> "$OUTPUT_FILE"
            echo "================================================================================================" >> "$OUTPUT_FILE"
            echo "" >> "$OUTPUT_FILE"

            # Write the file content
            cat "$file" >> "$OUTPUT_FILE"
            echo "" >> "$OUTPUT_FILE"
            echo "" >> "$OUTPUT_FILE"

            ((file_count++))
        fi
    done < <(find "$folder" -type f -print0)
done

echo "----------------------------------------"
echo "Code aggregation completed!"
echo "Processed $file_count files"
echo "Output written to: $(realpath "$OUTPUT_FILE")"