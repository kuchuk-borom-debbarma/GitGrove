#!/bin/bash
set -e

# Define paths
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
BUILD_DIR="$PROJECT_ROOT/build"
DEMO_DIR="$BUILD_DIR/demo"
GG_BIN="$BUILD_DIR/gg"

if [ ! -f "$GG_BIN" ]; then
    echo "Error: gg binary not found at $GG_BIN. Run build_gg.sh first."
    exit 1
fi

if [ ! -d "$DEMO_DIR" ]; then
    echo "Error: Demo dir not found at $DEMO_DIR. Run create_demo_repo.sh first."
    exit 1
fi

echo "Initializing git in demo repository..."
cd "$DEMO_DIR"
# Initialize git if not already initialized
if [ ! -d ".git" ]; then
    git init
    git add .
    git commit -m "Initial commit"
else
    echo "Git already initialized."
fi

echo "Initializing GitGroove..."
"$GG_BIN" init

echo "Registering repositories..."
"$GG_BIN" register --name root --path .
"$GG_BIN" register --name backend --path repo/backend
"$GG_BIN" register --name frontend --path repo/frontend
"$GG_BIN" register --name service-a --path repo/backend/service-a
"$GG_BIN" register --name service-b --path repo/backend/service-b
"$GG_BIN" register --name web-app --path repo/frontend/web-app

echo "Linking repositories..."
"$GG_BIN" link --child backend --parent root
"$GG_BIN" link --child frontend --parent root
"$GG_BIN" link --child service-a --parent backend
"$GG_BIN" link --child service-b --parent backend
"$GG_BIN" link --child web-app --parent frontend

echo "Demo initialization complete!"
