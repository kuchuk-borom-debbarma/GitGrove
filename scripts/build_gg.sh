#!/bin/bash
set -e

# Define paths
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
BUILD_DIR="$PROJECT_ROOT/build"

echo "Building gitgrove as gg..."
mkdir -p "$BUILD_DIR"

# Build from the cli directory
cd "$PROJECT_ROOT/cli"
go build -o "$BUILD_DIR/gg" ./cmd/main.go

echo "Build complete: $BUILD_DIR/gg"
