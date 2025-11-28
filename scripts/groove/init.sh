#!/bin/bash
set -e

# Get the root of the repo
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Run the setup script to build and reset the demo environment
./scripts/setup_demo.sh

# Navigate to the demo directory
cd build/demo

# Run the init command
./gitgrove init
