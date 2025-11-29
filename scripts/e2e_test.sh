#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[TEST] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

# Cleanup previous run
rm -rf e2e_temp
mkdir e2e_temp
cd e2e_temp

# Path to gitgrove binary (assuming it's built in the parent dir)
GITGROVE=../gitgrove

log "Initializing Git Repo..."
git init
git config user.email "you@example.com"
git config user.name "Your Name"
git commit --allow-empty -m "Root commit"
INITIAL_BRANCH=$(git symbolic-ref --short HEAD)

log "Initializing GitGrove..."
$GITGROVE init

# Verify system branch exists
if git show-ref --verify --quiet refs/heads/gitgroove/system; then
    log "System branch created."
else
    error "System branch not created."
fi

log "Registering 'backend' repo..."
mkdir backend
$GITGROVE register --name backend --path ./backend

# Verify marker
if [ -f backend/.gitgroverepo ]; then
    log "Backend marker created."
else
    error "Backend marker missing."
fi

# Verify orphan branch
if git show-ref --verify --quiet refs/heads/gitgroove/repos/backend/branches/main; then
    log "Backend orphan branch created."
else
    error "Backend orphan branch missing."
fi

log "Committing backend marker..."
git add .
git commit -m "Add backend marker"

log "Registering 'frontend' repo..."
mkdir frontend
$GITGROVE register --name frontend --path ./frontend

log "Committing frontend marker..."
git add .
git commit -m "Add frontend marker"

log "Linking frontend -> backend..."
$GITGROVE link --child frontend --parent backend

# Verify metadata
git checkout gitgroove/system
if grep -q "backend" .gg/repos/frontend/parent; then
    log "Link metadata verified."
else
    error "Link metadata missing."
fi
git checkout $INITIAL_BRANCH

log "Switching to backend..."
$GITGROVE switch backend

# Verify HEAD
CURRENT_BRANCH=$(git symbolic-ref --short HEAD)
if [ "$CURRENT_BRANCH" == "gitgroove/repos/backend/branches/main" ]; then
    log "Switched to backend branch."
else
    error "Failed to switch. Current branch: $CURRENT_BRANCH"
fi

log "Creating and Committing file in backend..."
# Directory and marker should persist because they are seeded in the orphan branch.
echo "package main" > backend/main.go
$GITGROVE stage backend/main.go
$GITGROVE commit -m "Initial backend commit"

# Verify commit
if git log -1 --pretty=%B | grep -q "Initial backend commit"; then
    log "Commit verified."
else
    error "Commit message mismatch."
fi

log "Moving backend to services/backend..."
# Move requires us to be in a clean state or handle the move carefully.
# The move command updates metadata.
$GITGROVE move --repo backend --to ./services/backend

# Verify physical move
if [ -d services/backend ] && [ ! -d backend ]; then
    log "Directory moved."
else
    error "Directory move failed."
fi

# Verify metadata update
git checkout gitgroove/system
if grep -q "services/backend" .gg/repos/backend/path; then
    log "Move metadata verified."
else
    error "Move metadata mismatch."
fi

log "All tests passed!"
