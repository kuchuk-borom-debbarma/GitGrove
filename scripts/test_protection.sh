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

error() {
    echo -e "${RED}✗ [ERROR] $1${NC}"
    exit 1
}

# Setup paths
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
BUILD_DIR="$PROJECT_ROOT/build"
GG_BIN="$BUILD_DIR/gg"
TEST_DIR="/tmp/gg_protection_test_$(date +%s)"

info "========================================="
info "GitGrove Protection Mechanism Test Suite"
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

# Initialize GitGrove
info "TEST 1: Initializing GitGrove..."
"$GG_BIN" init
log "GitGrove initialized"

# Verify we're on gitgroove/internal
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "gitgroove/internal" ]]; then
    error "Expected to be on gitgroove/internal, but on $CURRENT_BRANCH"
fi
log "Confirmed on gitgroove/internal branch"

# TEST 2: Try to stage a file on internal branch (should fail)
info "TEST 2: Testing staging protection on internal branch..."
echo "test content" > test.txt
set +e
OUTPUT=$("$GG_BIN" add test.txt 2>&1)
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -eq 0 ]; then
    error "Staging should have been rejected on internal branch, but succeeded"
fi

if [[ "$OUTPUT" == *"cannot stage files on gitgroove/internal branch"* ]]; then
    log "Staging correctly rejected with proper error message"
else
    error "Staging rejected but with wrong error message: $OUTPUT"
fi

# TEST 3: Try to commit on internal branch (should fail)
info "TEST 3: Testing commit protection on internal branch..."
# First stage with git directly to bypass our protection
git add test.txt
set +e
OUTPUT=$("$GG_BIN" commit -m "test commit" 2>&1)
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -eq 0 ]; then
    error "Commit should have been rejected on internal branch, but succeeded"
fi

if [[ "$OUTPUT" == *"cannot commit on gitgroove/internal branch"* ]]; then
    log "Commit correctly rejected with proper error message"
else
    error "Commit rejected but with wrong error message: $OUTPUT"
fi

# Clean up staged file
git reset HEAD test.txt
rm test.txt  # Remove the test file

# TEST 4: Register a repo and switch to it
info "TEST 4: Testing that add/commit work on repo branches..."
mkdir -p repo1
"$GG_BIN" register --name repo1 --path repo1
"$GG_BIN" switch repo1 main

# Verify we're on a repo branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "gitgroove/repos/repo1/branches/main" ]]; then
    error "Expected to be on repo branch, but on $CURRENT_BRANCH"
fi
log "Successfully switched to repo branch"

# TEST 5: Verify add works on repo branch
info "TEST 5: Testing that add works on repo branch..."
echo "repo content" > file1.txt
"$GG_BIN" add file1.txt
GIT_STATUS=$(git status --porcelain)
if [[ "$GIT_STATUS" == *"A  file1.txt"* ]]; then
    log "Add works correctly on repo branch"
else
    error "Add failed on repo branch: $GIT_STATUS"
fi

# TEST 6: Verify commit works on repo branch
info "TEST 6: Testing that commit works on repo branch..."
"$GG_BIN" commit -m "Add file1"
LAST_COMMIT=$(git log -1 --pretty=%B)
if [[ "$LAST_COMMIT" == *"Add file1"* ]]; then
    log "Commit works correctly on repo branch"
else
    error "Commit failed on repo branch"
fi

# TEST 7: Switch back to internal and verify protection still works
info "TEST 7: Verifying protection after switching back to internal..."
"$GG_BIN" cd ~

CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "gitgroove/internal" ]]; then
    error "Expected to be on gitgroove/internal after cd ~"
fi

echo "another test" > test2.txt
set +e
OUTPUT=$("$GG_BIN" add test2.txt 2>&1)
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -eq 0 ]; then
    error "Staging should still be rejected on internal branch"
fi

if [[ "$OUTPUT" == *"cannot stage files on gitgroove/internal branch"* ]]; then
    log "Protection still works after switching branches"
else
    error "Protection not working after branch switch: $OUTPUT"
fi

# Cleanup
info "Cleaning up test directory..."
cd /tmp
rm -rf "$TEST_DIR"

info "========================================="
log "ALL PROTECTION TESTS PASSED! ✓"
info "========================================="
