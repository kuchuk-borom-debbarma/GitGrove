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

# Create README.txt
cat <<EOF > "$TEMP_DIR/README.txt"
GitGrove (gg) Installation

1. Move the binary to a location in your PATH (e.g., /usr/local/bin).
   sudo mv gg /usr/local/bin/  (Linux/Mac)
   Move gg.exe to a folder in your PATH (Windows)

2. OR add the current directory to your PATH.
   export PATH=\$PATH:/path/to/extracted/folder
EOF

# Helper function
package_platform() {
    local GOOS=$1
    local GOARCH=$2
    # Removed SETUP_SCRIPT argument
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
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-X main.BuildTime=$(date +'%d-%m-%y-%H-%M-%S')" -o "$TARGET_DIR/$BIN_NAME" ./cmd/gitgrove/main.go
    popd > /dev/null

    # Add README
    cp "$TEMP_DIR/README.txt" "$TARGET_DIR/README.txt"

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
package_platform darwin amd64
package_platform darwin arm64

# --- Linux ---
package_platform linux amd64
package_platform linux arm64

# --- Windows ---
package_platform windows amd64
package_platform windows arm64

echo ""
echo "🎉 All releases packaged in $DIST_DIR"
rm -rf "$TEMP_DIR"
