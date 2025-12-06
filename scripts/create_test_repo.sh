#!/bin/bash

# Target directory
TARGET_DIR="../build/test-repo"

# Ensure we are in the scripts directory
cd "$(dirname "$0")"

# Clean up previous run
if [ -d "$TARGET_DIR" ]; then
    echo "removing existing $TARGET_DIR..."
    rm -rf "$TARGET_DIR"
fi

# Create directory
mkdir -p "$TARGET_DIR"
cd "$TARGET_DIR"

# Initialize Git
git init
echo "Initialized empty Git repository in $(pwd)"

# Create README
echo "# Test Repo for GitGrove" > README.md
echo "This is a dummy monorepo for testing." >> README.md

# Create serviceA
mkdir -p serviceA
echo "package main" > serviceA/main.go
echo "func main() { println(\"Hello from Service A\") }" >> serviceA/main.go
echo "# Service A" > serviceA/README.md

# Create frontend
mkdir -p frontend
echo "<html><body><h1>Hello World</h1></body></html>" > frontend/index.html
echo "console.log('Frontend app');" > frontend/app.js
echo "# Frontend" > frontend/README.md

# Commit all
git add .
git commit -m "Initial commit for test repo"

echo "Test repo created successfully at $TARGET_DIR"
echo "Structure:"
find . -maxdepth 3 -not -path '*/.*'
