#!/bin/bash
set -e

# Define paths
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
RELEASE_DIR="$PROJECT_ROOT/build/release"
SRC_DIR="$PROJECT_ROOT/src"

echo "Building GitGrove releases..."
mkdir -p "$RELEASE_DIR"

# Navigate to source
cd "$SRC_DIR"

# Mac (Darwin)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o "$RELEASE_DIR/gg-darwin-amd64" cmd/gitgrove/main.go
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o "$RELEASE_DIR/gg-darwin-arm64" cmd/gitgrove/main.go

# Linux
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$RELEASE_DIR/gg-linux-amd64" cmd/gitgrove/main.go
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -o "$RELEASE_DIR/gg-linux-arm64" cmd/gitgrove/main.go

# Windows
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$RELEASE_DIR/gg-windows-amd64.exe" cmd/gitgrove/main.go

echo "Builds complete! Artifacts are in $RELEASE_DIR"
ls -lh "$RELEASE_DIR"
