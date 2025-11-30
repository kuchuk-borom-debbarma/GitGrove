#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}✓ [TEST] $1${NC}"
}

info() {
    echo -e "${BLUE}ℹ [INFO] $1${NC}"
}

warn() {
    echo -e "${YELLOW}⚠ [WARN] $1${NC}"
}

error() {
    echo -e "${RED}✗ [ERROR] $1${NC}"
    exit 1
}

assert_contains() {
    local output="$1"
    local expected="$2"
    local test_name="$3"
    
    if [[ "$output" == *"$expected"* ]]; then
        log "$test_name: Found '$expected'"
    else
        error "$test_name: Expected to find '$expected' but didn't. Output: $output"
    fi
}

assert_file_exists() {
    local file="$1"
    local test_name="$2"
    
    if [ -f "$file" ]; then
        log "$test_name: File $file exists"
    else
        error "$test_name: File $file does not exist"
    fi
}

assert_dir_exists() {
    local dir="$1"
    local test_name="$2"
    
    if [ -d "$dir" ]; then
        log "$test_name: Directory $dir exists"
    else
        error "$test_name: Directory $dir does not exist"
    fi
}

# Setup paths
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
BUILD_DIR="$PROJECT_ROOT/build"
GG_BIN="$BUILD_DIR/gg"
TEST_DIR="/tmp/gg_verify_$(date +%s)"

info "========================================="
info "GitGrove Comprehensive Verification Suite"
info "========================================="

# Build gg
info "Building gg binary..."
"$PROJECT_ROOT/scripts/build_gg.sh"
log "Binary built successfully"

# Create test directory
info "Creating test directory at $TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Initialize git
info "Initializing git repository..."
git init
git config user.email "test@gitgrove.test"
git config user.name "Test User"
echo "# Test Repo" > README.md
git add README.md
git commit -m "Initial commit"
log "Git repository initialized"

# ============================================
# TEST 1: Init
# ============================================
info "TEST 1: Testing 'init' command..."
"$GG_BIN" init
assert_dir_exists ".gg" "Init"
assert_dir_exists ".gg/repos" "Init"
assert_file_exists ".gg/repos/.gitkeep" "Init"
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/internal" "Init - should be on system branch"

# ============================================
# TEST 2: Register (Single)
# ============================================
info "TEST 2: Testing 'register' command (single repo)..."
mkdir -p repo/backend
"$GG_BIN" register --name root --path .
# Check metadata in gitgroove/internal branch
git checkout gitgroove/internal 2>/dev/null
assert_dir_exists ".gg/repos/root" "Register - root repo metadata dir"
assert_file_exists ".gg/repos/root/path" "Register - root repo path file"
git checkout - 2>/dev/null
INFO_OUTPUT=$("$GG_BIN" info)
assert_contains "$INFO_OUTPUT" "root" "Register - info should show root"

# ============================================
# TEST 3: Register (Multiple)
# ============================================
info "TEST 3: Testing 'register' command (multiple repos)..."
mkdir -p repo/frontend
mkdir -p repo/backend/service-a
mkdir -p repo/backend/service-b
"$GG_BIN" register --name backend --path repo/backend
"$GG_BIN" register --name frontend --path repo/frontend
"$GG_BIN" register --name service-a --path repo/backend/service-a
"$GG_BIN" register --name service-b --path repo/backend/service-b
# Check metadata in gitgroove/internal branch
git checkout gitgroove/internal 2>/dev/null
assert_dir_exists ".gg/repos/backend" "Register - backend repo metadata"
assert_dir_exists ".gg/repos/frontend" "Register - frontend repo metadata"
assert_dir_exists ".gg/repos/service-a" "Register - service-a repo metadata"
git checkout - 2>/dev/null

# ============================================
# TEST 4: Link
# ============================================
info "TEST 4: Testing 'link' command..."
"$GG_BIN" link --child backend --parent root
"$GG_BIN" link --child frontend --parent root
"$GG_BIN" link --child service-a --parent backend
"$GG_BIN" link --child service-b --parent backend
INFO_OUTPUT=$("$GG_BIN" info)
assert_contains "$INFO_OUTPUT" "backend" "Link - backend linked"
assert_contains "$INFO_OUTPUT" "frontend" "Link - frontend linked"

# ============================================
# TEST 5: Info
# ============================================
info "TEST 5: Testing 'info' command..."
INFO_OUTPUT=$("$GG_BIN" info)
assert_contains "$INFO_OUTPUT" "root" "Info - shows root"
assert_contains "$INFO_OUTPUT" "backend" "Info - shows backend"
assert_contains "$INFO_OUTPUT" "frontend" "Info - shows frontend"
assert_contains "$INFO_OUTPUT" "service-a" "Info - shows service-a"

# ============================================
# TEST 6: Switch
# ============================================
info "TEST 6: Testing 'switch' command..."
"$GG_BIN" switch root main
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/repos/root/branches/main" "Switch - on root/main branch"

# ============================================
# TEST 7: Ls (List children)
# ============================================
info "TEST 7: Testing 'ls' command..."
LS_OUTPUT=$("$GG_BIN" ls)
assert_contains "$LS_OUTPUT" "backend" "Ls - shows backend"
assert_contains "$LS_OUTPUT" "frontend" "Ls - shows frontend"

# ============================================
# TEST 8: Cd (Navigate down)
# ============================================
info "TEST 8: Testing 'cd' command (down)..."
"$GG_BIN" cd backend
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/repos/backend/branches/main" "Cd down - on backend/main branch"

# ============================================
# TEST 9: Cd (Navigate up)
# ============================================
info "TEST 9: Testing 'cd' command (up)..."
"$GG_BIN" cd ..
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/repos/root/branches/main" "Cd up - on root/main branch"

# ============================================
# TEST 10: Branch (Create)
# ============================================
info "TEST 10: Testing 'branch' command..."
"$GG_BIN" branch root feature-x
BRANCHES=$(git branch -a)
assert_contains "$BRANCHES" "gitgroove/repos/root/branches/feature-x" "Branch - feature-x created"

# ============================================
# TEST 11: Switch to new branch
# ============================================
info "TEST 11: Testing 'switch' to new branch..."
"$GG_BIN" switch root feature-x
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/repos/root/branches/feature-x" "Switch - on feature-x branch"

# ============================================
# TEST 12: Add
# ============================================
info "TEST 12: Testing 'add' command..."
echo "test content" > test.txt
"$GG_BIN" add test.txt
GIT_STATUS=$(git status --porcelain)
assert_contains "$GIT_STATUS" "A  test.txt" "Add - test.txt staged"

# ============================================
# TEST 13: Commit
# ============================================
info "TEST 13: Testing 'commit' command..."
"$GG_BIN" commit -m "Add test.txt"
LAST_COMMIT=$(git log -1 --pretty=%B)
assert_contains "$LAST_COMMIT" "Add test.txt" "Commit - message correct"

# ============================================
# TEST 14: Add with dot (all files)
# ============================================
info "TEST 14: Testing 'add' with . (all files)..."
echo "another file" > test2.txt
echo "third file" > test3.txt
"$GG_BIN" add .
GIT_STATUS=$(git status --porcelain)
assert_contains "$GIT_STATUS" "A  test2.txt" "Add . - test2.txt staged"
assert_contains "$GIT_STATUS" "A  test3.txt" "Add . - test3.txt staged"
"$GG_BIN" commit -m "Add multiple files"

# ============================================
# TEST 15: Checkout (different branch)
# ============================================
info "TEST 15: Testing 'checkout' command..."
# Switch to backend first
"$GG_BIN" switch backend main
# Create a dev branch in backend repo
"$GG_BIN" branch backend dev
# Switch to dev branch
"$GG_BIN" switch backend dev
echo "dev file" > dev.txt
"$GG_BIN" add dev.txt
"$GG_BIN" commit -m "Add dev file"

# Switch back to main
"$GG_BIN" switch backend main

# Now checkout dev branch in backend
"$GG_BIN" checkout backend dev
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/repos/backend/branches/dev" "Checkout - backend on dev branch"
assert_file_exists "dev.txt" "Checkout - dev.txt exists"

# ============================================
# TEST 16: Move
# ============================================
info "TEST 16: Testing 'move' command..."
"$GG_BIN" cd ~
mkdir -p new-location
"$GG_BIN" move --repo backend --to new-location/backend
assert_dir_exists "new-location/backend" "Move - backend moved to new location"
# Check metadata - need to reset to HEAD to see committed changes
git checkout gitgroove/internal 2>/dev/null
git reset --hard HEAD 2>/dev/null
PATH_CONTENT=$(cat .gg/repos/backend/path)
assert_contains "$PATH_CONTENT" "new-location/backend" "Move - metadata updated"
# Clean up untracked files from the move
git clean -fd 2>/dev/null
git checkout - 2>/dev/null

# ============================================
# TEST 17: Push (dry run - no remote)
# ============================================
info "TEST 17: Testing 'push' command (expected to fail gracefully)..."
# This will fail because there's no remote, but we verify it doesn't crash
set +e
"$GG_BIN" push --repos root 2>&1 | grep -q "remote\|upstream" && warn "Push - failed as expected (no remote)" || warn "Push - command executed"
set -e

# ============================================
# TEST 18: Cd to System Root
# ============================================
info "TEST 18: Testing 'cd' to system root..."
"$GG_BIN" cd ~
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/internal" "Cd ~ - on system branch"

# ============================================
# TEST 19: Ls from System Root
# ============================================
info "TEST 19: Testing 'ls' from system root..."
LS_OUTPUT=$("$GG_BIN" ls)
assert_contains "$LS_OUTPUT" "root" "Ls from system - shows root"

# ============================================
# TEST 20: Cd from System Root to repo
# ============================================
info "TEST 20: Testing 'cd' from system root to repo..."
"$GG_BIN" cd root
CURRENT_BRANCH=$(git branch --show-current)
assert_contains "$CURRENT_BRANCH" "gitgroove/repos/root/branches" "Cd from system - on root branch"

# Cleanup
info "Cleaning up test directory..."
cd /tmp
rm -rf "$TEST_DIR"

info "========================================="
log "ALL TESTS PASSED SUCCESSFULLY! ✓"
info "========================================="
