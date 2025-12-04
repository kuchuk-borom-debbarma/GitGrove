#!/bin/bash
set -e

# Define paths
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
BUILD_DIR="$PROJECT_ROOT/build"

echo "Building gitgrove as gg..."
mkdir -p "$BUILD_DIR"

# Build from the src directory
cd "$PROJECT_ROOT/src"
go build -o "$BUILD_DIR/gg" core/cli/main.go

echo "Build complete: $BUILD_DIR/gg"
