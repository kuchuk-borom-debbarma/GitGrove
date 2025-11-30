#!/bin/bash
set -e

# Setup
TEST_DIR="test_full_env"
rm -rf $TEST_DIR remote.git
mkdir -p $TEST_DIR
cd $TEST_DIR

# Helper to assert success
assert_success() {
    if [ $? -eq 0 ]; then
        echo "✅ $1 passed"
    else
        echo "❌ $1 failed"
        exit 1
    fi
}

# Helper to assert failure
assert_fail() {
    if [ $? -ne 0 ]; then
        echo "✅ $1 failed as expected"
    else
        echo "❌ $1 succeeded but should have failed"
        exit 1
    fi
}

echo "==============================================="
echo "      GitGrove Comprehensive Verification      "
echo "==============================================="

# 1. Initialization
echo "--- 1. Initialization ---"
git init
../gitgrove init
assert_success "gg init"
git checkout -b main
assert_success "checkout main"

# 2. Registration
echo "--- 2. Registration ---"
mkdir -p services/backend services/frontend libs/shared
touch services/backend/main.go services/frontend/index.html libs/shared/util.go
git add .
git commit -m "Initial project structure"

../gitgrove register 'backend;./services/backend' 'frontend;./services/frontend' 'shared;./libs/shared'
assert_success "gg register multiple"

git add .
git commit -m "Register repos"

# Test duplicate registration (should fail)
../gitgrove register 'backend;./services/backend' 2>/dev/null && assert_fail "Duplicate register" || echo "✅ Duplicate register failed as expected"

# 3. Linking
echo "--- 3. Linking ---"
../gitgrove link 'backend;shared' 'frontend;shared'
assert_success "gg link parent-child"

# Test cycle (should fail)
../gitgrove link 'shared;backend' 2>/dev/null && assert_fail "Cycle link" || echo "✅ Cycle link failed as expected"

# 4. Navigation & Switching
echo "--- 4. Navigation & Switching ---"
../gitgrove switch backend
assert_success "gg switch backend"

# Verify flattened view
if [ -f "main.go" ] && [ ! -d "services" ]; then
    echo "✅ Flattened view confirmed"
else
    echo "❌ Flattened view incorrect"
    ls -R
    exit 1
fi

# 5. Add & Commit (Scoped)
echo "--- 5. Add & Commit ---"
echo "// comment" >> main.go
../gitgrove add main.go
assert_success "gg add file"

# Test adding out of scope file (should fail/warn)
# In flattened view, we are at services/backend.
# Try to add frontend file: ../frontend/index.html (since backend and frontend are in services/)
../gitgrove add ../frontend/index.html 2>/dev/null
if git diff --cached --name-only | grep -q "frontend/index.html"; then
    echo "❌ Staged out of scope file (frontend)"
    exit 1
else
    echo "✅ Out of scope file skipped"
fi

# Test adding .gg metadata (should fail)
../gitgrove add .gg/repos/backend/path 2>/dev/null
if git diff --cached --name-only | grep -q ".gg/"; then
    echo "❌ Staged metadata file"
    exit 1
else
    echo "✅ Metadata file skipped"
fi

../gitgrove commit -m "Update backend main"
assert_success "gg commit"

# 6. Branching
echo "--- 6. Branching ---"
../gitgrove branch backend feature-auth
assert_success "gg branch"
../gitgrove checkout backend feature-auth
assert_success "gg checkout"

echo "auth logic" > auth.go
../gitgrove add auth.go
../gitgrove commit -m "Add auth"

# 7. Navigation (Up/Down/Ls)
echo "--- 7. Navigation ---"
../gitgrove up
assert_success "gg up (to shared)"

# Verify we are in shared
if [ -f "util.go" ]; then
    echo "✅ Navigated to parent (shared)"
else
    echo "❌ Failed to navigate to parent"
    exit 1
fi

../gitgrove ls > ls_output.txt
if grep -q "backend" ls_output.txt && grep -q "frontend" ls_output.txt; then
    echo "✅ gg ls listed children"
else
    echo "❌ gg ls failed"
    cat ls_output.txt
    exit 1
fi
rm ls_output.txt

../gitgrove down frontend
assert_success "gg down frontend"

# 8. Info
echo "--- 8. Info ---"
../gitgrove info > info_output.txt
assert_success "gg info"
cat info_output.txt
rm info_output.txt

# 9. Move
echo "--- 9. Move ---"
# Go back to root context (system root or just up until root?)
# Move requires clean state and physical directory presence.
# Since we are in sparse checkout, we need to restore full view.
git sparse-checkout disable
git checkout main

mkdir -p services/web
../gitgrove move --repo frontend --to services/web/frontend
assert_success "gg move"

# Commit the move on main branch
git add .
git commit -m "Move frontend to services/web"

# 10. Push
echo "--- 10. Push ---"
# Setup remote
cd ..
mkdir remote.git
cd remote.git
git init --bare
cd ../$TEST_DIR

git remote add origin ../remote.git

# Push backend (which has new branch feature-auth and main)
# We need to switch to backend first to push? No, gg push takes args.
# But we need to be in a GitGrove context.
# We are currently in 'shared' repo.

# Push backend only (pushing multiple repos to same origin/main causes conflict)
../gitgrove push backend
assert_success "gg push backend"

# Verify remote
cd ../remote.git
if git rev-parse --verify refs/heads/main >/dev/null 2>&1; then
    echo "✅ Remote has 'main' branch"
else
    echo "❌ Remote missing 'main' branch"
    exit 1
fi

echo "==============================================="
echo "      Verification Completed Successfully      "
echo "==============================================="
