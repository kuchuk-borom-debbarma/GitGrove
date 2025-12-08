#!/bin/bash
set -e

# Package GitGrove Releases
# Builds binaries for multiple platforms, includes setup scripts, and zips them.

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DIST_DIR="$PROJECT_ROOT/build/dist"
SRC_DIR="$PROJECT_ROOT/src"
TEMP_DIR="$PROJECT_ROOT/build/temp"

echo "📦 Packaging GitGrove Releases..."

# Cleanup
rm -rf "$DIST_DIR"
rm -rf "$TEMP_DIR"
mkdir -p "$DIST_DIR"

# Helper function
package_platform() {
    local GOOS=$1
    local GOARCH=$2
    local SETUP_SCRIPT=$3
    local ARCHIVE_NAME="GitGrove_${GOOS}_${GOARCH}"
    local TARGET_DIR="$TEMP_DIR/$ARCHIVE_NAME"
    local BIN_NAME="gg"

    if [ "$GOOS" == "windows" ]; then
        BIN_NAME="gg.exe"
    fi

    echo "🔹 Packaging $ARCHIVE_NAME..."
    mkdir -p "$TARGET_DIR"

    # Build
    pushd "$SRC_DIR" > /dev/null
    GOOS=$GOOS GOARCH=$GOARCH go build -o "$TARGET_DIR/$BIN_NAME" ./cmd/gitgrove/main.go
    popd > /dev/null

    # Add Setup Script
    if [ "$GOOS" == "windows" ]; then
        cp "$PROJECT_ROOT/scripts/$SETUP_SCRIPT" "$TARGET_DIR/setup.ps1"
    else
        cp "$PROJECT_ROOT/scripts/$SETUP_SCRIPT" "$TARGET_DIR/setup.sh"
        chmod +x "$TARGET_DIR/setup.sh"
    fi

    # Zip
    pushd "$TEMP_DIR" > /dev/null
    if [ "$GOOS" == "windows" ]; then
         # Use zip if available, else warn (or just tar? user accepted rar/zip)
         # Using zip command
         zip -r "$DIST_DIR/$ARCHIVE_NAME.zip" "$ARCHIVE_NAME" > /dev/null
    else
         zip -r "$DIST_DIR/$ARCHIVE_NAME.zip" "$ARCHIVE_NAME" > /dev/null
    fi
    popd > /dev/null

    echo "✅ Created $DIST_DIR/$ARCHIVE_NAME.zip"
}

# --- Mac ---
package_platform darwin amd64 setup_mac.sh
package_platform darwin arm64 setup_mac.sh

# --- Linux ---
package_platform linux amd64 setup_linux.sh
package_platform linux arm64 setup_linux.sh

# --- Windows ---
package_platform windows amd64 setup_windows.ps1
package_platform windows arm64 setup_windows.ps1

echo ""
echo "🎉 All releases packaged in $DIST_DIR"
rm -rf "$TEMP_DIR"
