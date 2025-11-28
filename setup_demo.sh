#!/bin/bash
set -e

# Define paths
PROJECT_ROOT=$(pwd)
BUILD_DIR="$PROJECT_ROOT/build"
DEMO_DIR="$BUILD_DIR/demo"

echo "Building gitgrove..."
mkdir -p "$BUILD_DIR"

# Build from the cli directory
cd cli
go build -o "$BUILD_DIR/gitgrove" ./cmd/main.go
cd ..

echo "Creating demo repository structure..."
rm -rf "$DEMO_DIR" # Clean up previous run
mkdir -p "$DEMO_DIR"
# Create nested directories
mkdir -p "$DEMO_DIR/repo/backend/service-a"
mkdir -p "$DEMO_DIR/repo/backend/service-b"
mkdir -p "$DEMO_DIR/repo/frontend/web-app"

echo "Creating dummy files..."
echo "# Demo Monorepo" > "$DEMO_DIR/README.md"

# Backend Service A
echo "package main; func main() { println(\"Service A\") }" > "$DEMO_DIR/repo/backend/service-a/main.go"
echo "# Service A" > "$DEMO_DIR/repo/backend/service-a/README.md"

# Backend Service B
echo "package main; func main() { println(\"Service B\") }" > "$DEMO_DIR/repo/backend/service-b/main.go"
echo "# Service B" > "$DEMO_DIR/repo/backend/service-b/README.md"

# Frontend Web App
echo '{ "name": "web-app", "version": "1.0.0" }' > "$DEMO_DIR/repo/frontend/web-app/package.json"
echo "# Web App" > "$DEMO_DIR/repo/frontend/web-app/README.md"

echo "Initializing git in demo repository..."
cd "$DEMO_DIR"
git init
git add .
git commit -m "Initial commit"

echo "Copying gitgrove binary to demo root..."
cp "$BUILD_DIR/gitgrove" .
git add gitgrove
git commit -m "Add gitgrove binary"

echo "Setup complete!"
echo "You can now run the demo:"
echo "cd $DEMO_DIR"
echo "./gitgrove init"
