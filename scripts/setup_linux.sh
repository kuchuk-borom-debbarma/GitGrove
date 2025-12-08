#!/bin/bash
# Setup script to add this directory to PATH
# Intended to be run after unzipping the release.

BIN_DIR="$(pwd)"
echo "GitGrove Linux Setup"
echo "Adding $BIN_DIR to PATH..."

RC_FILE=""
if [ -f "$HOME/.bashrc" ]; then
    RC_FILE="$HOME/.bashrc"
elif [ -f "$HOME/.zshrc" ]; then
    RC_FILE="$HOME/.zshrc"
else
    echo "⚠️  No .bashrc or .zshrc found. Please add '$BIN_DIR' to PATH manually."
    exit 0
fi

if grep -Eq ":$BIN_DIR(:|\"|$)" "$RC_FILE"; then
    echo "Already in PATH."
else
    echo "" >> "$RC_FILE"
    echo "# GitGrove" >> "$RC_FILE"
    echo "export PATH=\$PATH:\"$BIN_DIR\"" >> "$RC_FILE"
    echo "✅ Added to $RC_FILE"
fi

echo "👉 Please restart your terminal or run: source $RC_FILE"
