#!/bin/bash
set -e

# Get the root of the repo
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT/build/demo"

echo "Registering repositories..."

# Register backend services
# Assuming CLI syntax: gitgrove register --name <name> --path <path>
# Or: gitgrove register <name> <path>
# Based on common CLI patterns, I'll use flags if possible, or positional args.
# Since I haven't implemented it yet, I'll define the expected usage here.

./gitgrove register --name service-a --path repo/backend/service-a
./gitgrove register --name service-b --path repo/backend/service-b
./gitgrove register --name web-app --path repo/frontend/web-app

echo "Registration complete."
