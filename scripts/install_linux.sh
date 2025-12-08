#!/bin/bash
set -e

# GitGrove Installer for Linux
# Builds the binary & Adds to PATH

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
BUILD_DIR="$PROJECT_ROOT/build"
OLD_RELEASE_DIR="$PROJECT_ROOT/build/release"
SRC_DIR="$PROJECT_ROOT/src"

echo "🚀 Installing GitGrove for Linux..."

if [ -d "$OLD_RELEASE_DIR" ]; then
    rm -rf "$OLD_RELEASE_DIR"
fi

echo "📦 Building binary..."
mkdir -p "$BUILD_DIR"
cd "$SRC_DIR"
GOOS=linux go build -o "$BUILD_DIR/gg" ./cmd/gitgrove/main.go
chmod +x "$BUILD_DIR/gg"
echo "✅ Build complete: $BUILD_DIR/gg"

# 3. Instructions
echo ""
echo "🎉 Build Complete!"
echo "To use 'gg', you need to add it to your PATH."
echo ""
echo "Option 1: Create a symlink (Recommended)"
echo "  sudo ln -sf $BUILD_DIR/gg /usr/local/bin/gg"
echo ""
echo "Option 2: Add build directory to PATH"
echo "  export PATH=\$PATH:$BUILD_DIR"
echo ""

