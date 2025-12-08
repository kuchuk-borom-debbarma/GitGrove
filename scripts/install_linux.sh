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

echo "🔗 Configuring PATH..."
RC_FILE=""
if [ -f "$HOME/.bashrc" ]; then
    RC_FILE="$HOME/.bashrc"
elif [ -f "$HOME/.zshrc" ]; then
    RC_FILE="$HOME/.zshrc"
else
    echo "⚠️  No .bashrc or .zshrc found. Please add '$BUILD_DIR' to PATH manually."
    exit 0
fi

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
