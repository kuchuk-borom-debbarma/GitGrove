#!/bin/bash
set -e

# Setup test environment
TEST_DIR="test_stage_env"
rm -rf "$TEST_DIR"
mkdir "$TEST_DIR"
cd "$TEST_DIR"

# Build gitgrove
# Build from cli directory where go.mod exists
cd ../cli
go build -o ../test_stage_env/gitgrove cmd/main.go
cd ../test_stage_env
CLI="$(pwd)/gitgrove"

echo "=== 1. Initialize Root Repo ==="
mkdir root
cd root
git init
$CLI init
echo "Root repo initialized."

echo "=== 2. Register Repos ==="
# Register root repo
$CLI register --name root --path .
echo "child/" > .gitignore
git add .gitignore
git commit -m "Ignore child repo"
git add .gitgroverepo
git commit -m "Add root marker"

# Create and register child repo
mkdir child
# Register will create .gitgroverepo, so we just need to commit the dir first if we want clean state?
# Actually, Register checks clean state. So we must have clean state BEFORE calling it.
# But we need the directory 'child' to exist.
touch child/.gitkeep
git add -f child
git commit -m "Add child dir"

$CLI register --name child --path child
git add -f child/.gitgroverepo
git commit -m "Add child marker"

echo "=== 3. Link Repos ==="
# Link child -> root
$CLI link --child child --parent root

echo "=== 4. Test Staging in Child Branch ==="
# Switch to child repo branch
$CLI switch child

touch child/child_file
echo "Staging child_file in child branch..."
$CLI stage child/child_file
echo "PASS: Staged child_file in child branch"
git commit -m "Staged child file"

echo "Attempting to stage root_file (outside child scope)..."
touch root_file
if $CLI stage root_file; then
    echo "FAIL: Should not have staged file outside child repo scope"
    exit 1
else
    echo "PASS: Correctly rejected file outside child repo scope"
fi
rm root_file # Cleanup

echo "=== 5. Test Staging in Root Branch ==="
# Switch to root repo branch
$CLI switch root

echo "Staging root_file in root branch..."
touch root_file
$CLI stage root_file
echo "PASS: Staged root_file in root branch"

echo "Attempting to stage child_file (nested repo)..."
touch child/child_file
if $CLI stage child/child_file; then
    echo "FAIL: Should not have staged file in nested repo"
    exit 1
else
    echo "PASS: Correctly rejected file in nested repo"
fi

echo "Attempting to stage .gg/ file..."
if $CLI stage .gg/repos/root/path; then
    echo "FAIL: Should not have staged .gg file"
    exit 1
else
    echo "PASS: Correctly rejected .gg file"
fi

echo "=== 6. Test Commit in Root Branch ==="
# Commit the staged root_file
$CLI commit -m "Commit root file"
echo "PASS: Committed root_file"

echo "=== 7. Test Commit with Invalid Staged File (Manual Add) ==="
# Manually stage a file that violates rules (e.g. nested repo file)
# Note: git add -f allows adding ignored files, but we want to test our commit validation.
# We are in root repo. Add child file manually.
touch child/manual_file
git add -f child/manual_file

if $CLI commit -m "Commit invalid file"; then
    echo "FAIL: Should not have committed invalid file"
    exit 1
else
    echo "PASS: Correctly rejected commit with invalid staged file"
fi
git restore --staged child/manual_file # Cleanup staging
rm child/manual_file # Cleanup file

echo "Debug: Checking status before switch"
git status

echo "=== 8. Test Commit with Identity Mismatch (Tampering) ==="
# Switch to child
$CLI switch child
# Tamper with marker
echo "fake-identity" > child/.gitgroverepo
touch child/valid_file
$CLI stage child/valid_file

if $CLI commit -m "Commit with bad marker"; then
    echo "FAIL: Should not have committed with bad marker"
    exit 1
else
    echo "PASS: Correctly rejected commit with identity mismatch"
fi
# Restore marker
echo "child" > child/.gitgroverepo

echo "All tests passed!"
