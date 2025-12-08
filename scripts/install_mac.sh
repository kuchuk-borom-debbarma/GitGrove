#!/bin/bash
set -e

# GitGrove Installer for MacOS
# Builds the binary & Adds to PATH

# Define paths
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
BUILD_DIR="$PROJECT_ROOT/build"
OLD_RELEASE_DIR="$PROJECT_ROOT/build/release"
SRC_DIR="$PROJECT_ROOT/src"

echo "🚀 Installing GitGrove for MacOS..."

# 1. Clean old artifacts to prevent PATH confusion
if [ -d "$OLD_RELEASE_DIR" ]; then
    echo "Cleaning up old release directory..."
    rm -rf "$OLD_RELEASE_DIR"
fi

# 2. Build
echo "📦 Building binary..."
mkdir -p "$BUILD_DIR"
cd "$SRC_DIR"
go build -o "$BUILD_DIR/gg" ./cmd/gitgrove/main.go
chmod +x "$BUILD_DIR/gg"
echo "✅ Build complete: $BUILD_DIR/gg"

# 3. Add to PATH
echo "🔗 Configuring PATH..."
SHELL_NAME=$(basename "$SHELL")
RC_FILE=""

if [ "$SHELL_NAME" = "zsh" ]; then
    RC_FILE="$HOME/.zshrc"
elif [ "$SHELL_NAME" = "bash" ]; then
    RC_FILE="$HOME/.bash_profile"
else
    echo "⚠️  Unsupported shell: $SHELL_NAME. Please add '$BUILD_DIR' to PATH manually."
    exit 0
fi

# Check if PATH is already set correctly using strict regex
if grep -Eq ":$BUILD_DIR(:|\"|$)" "$RC_FILE"; then
    echo "GitGrove is already in your PATH in $RC_FILE"
else
    echo "" >> "$RC_FILE"
    echo "# GitGrove" >> "$RC_FILE"
    echo "export PATH=\$PATH:$BUILD_DIR" >> "$RC_FILE"
    echo "✅ Added to $RC_FILE"
fi

echo ""
echo "🎉 Installation Complete! Version v1.0"
echo "👉 Please restart your terminal or run: source $RC_FILE"
