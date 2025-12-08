#!/bin/bash

# Target directory
TARGET_DIR="./test-repo"

# Ensure we are in the scripts directory
cd "$(dirname "$0")"

# Clean up previous run
if [ -d "$TARGET_DIR" ]; then
    echo "Removing existing $TARGET_DIR..."
    rm -rf "$TARGET_DIR"
fi

# Create directory
echo "Creating target directory..."
mkdir -p "$TARGET_DIR"
cd "$TARGET_DIR"

# Initialize Git
echo "Initializing Git..."
git init

# Create .gitignore
echo "Creating .gitignore..."
cat <<EOF > .gitignore
# Binaries
/bin
/build
*.exe
*.test

# Dependency directories
node_modules/
vendor/

# Temporary files
*.swp
*.log
.DS_Store
EOF

# Create Root Files
echo "Creating root files..."
echo "# GitGrove Test Monorepo (Isolated Projects)" > README.md
echo "This repo contains multiple independent projects to test GitGrove's isolation features." >> README.md

# --- Create Services (Isolated Go Modules) ---
echo "Creating isolated services..."

# 1. Order Service
mkdir -p services/order-service
echo "Initializing order-service..."
(
    cd services/order-service
    go mod init github.com/example/order-service
    
    cat <<EOF > main.go
package main

import "fmt"

func main() {
    fmt.Println("Starting Order Service...")
    // Business logic would go here
}
EOF

    cat <<EOF > Dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o service .
CMD ["./service"]
EOF
)

# 2. User Service
mkdir -p services/user-service
echo "Initializing user-service..."
(
    cd services/user-service
    go mod init github.com/example/user-service
    
    cat <<EOF > main.go
package main

import "fmt"

func main() {
    fmt.Println("Starting User Service...")
    // User logic would go here
}
EOF

    cat <<EOF > Dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o service .
CMD ["./service"]
EOF
)

# --- Create Frontend (Isolated Node Project) ---
echo "Creating isolated frontend..."

mkdir -p frontend/admin-panel
echo "Initializing admin-panel..."
(
    cd frontend/admin-panel
    
    cat <<EOF > package.json
{
  "name": "admin-panel",
  "version": "1.0.0",
  "scripts": {
    "start": "node src/index.js"
  }
}
EOF

    mkdir -p src
    echo "console.log('Starting Admin Panel');" > src/index.js
    echo "<html><body><h1>Admin Panel</h1></body></html>" > public_index.html
    echo "# Admin Panel" > README.md
)

# Initial Commit
echo "Committing files..."
git add .
git commit -m "Initial commit of isolated projects"

echo "------------------------------------------------"
echo "✅ Test repo created successfully at: $TARGET_DIR"
echo "------------------------------------------------"
echo "Structure:"
find . -maxdepth 3 -not -path '*/.*'
