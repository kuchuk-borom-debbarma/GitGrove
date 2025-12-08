#!/bin/bash

# GitGrove Installation Helper
# Adds 'gg' to your system PATH

echo "Welcome to GitGrove Installer!"
echo "This script will help you add the 'gg' binary to your PATH."
echo ""

# 1. Ask for location
echo "Please enter the absolute path to the directory containing your 'gg' binary:"
read -r BIN_DIR

# Validate path
if [ ! -d "$BIN_DIR" ]; then
    echo "Error: Directory '$BIN_DIR' does not exist."
    exit 1
fi

if [ ! -f "$BIN_DIR/gg" ]; then
    echo "Warning: 'gg' binary not found in '$BIN_DIR'. Proceeding anyway, but ensure it exists later."
fi

# 2. Detect Shell
SHELL_NAME=$(basename "$SHELL")
RC_FILE=""

if [ "$SHELL_NAME" = "zsh" ]; then
    RC_FILE="$HOME/.zshrc"
elif [ "$SHELL_NAME" = "bash" ]; then
    if [ -f "$HOME/.bashrc" ]; then
        RC_FILE="$HOME/.bashrc"
    else
        RC_FILE="$HOME/.bash_profile"
    fi
else
    echo "Unknown shell: $SHELL_NAME"
    echo "Please manually add the following line to your shell configuration:"
    echo "export PATH=\$PATH:$BIN_DIR"
    exit 0
fi

# 3. Add to PATH
echo "Detected shell: $SHELL_NAME"
echo "Adding to: $RC_FILE"

# Check if already present
if grep -q "export PATH=\$PATH:$BIN_DIR" "$RC_FILE"; then
    echo "GitGrove is already in your PATH!"
else
    echo "" >> "$RC_FILE"
    echo "# GitGrove" >> "$RC_FILE"
    echo "export PATH=\$PATH:$BIN_DIR" >> "$RC_FILE"
    echo "Successfully updated $RC_FILE"
fi

echo ""
echo "Installation complete!"
echo "Please restart your terminal or run: source $RC_FILE"
