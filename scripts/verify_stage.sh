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
git add .gitgroverepo
git commit -m "Add root marker"

# Create and register child repo
mkdir child
# Register will create .gitgroverepo, so we just need to commit the dir first if we want clean state?
# Actually, Register checks clean state. So we must have clean state BEFORE calling it.
# But we need the directory 'child' to exist.
touch child/.gitkeep
git add child
git commit -m "Add child dir"

$CLI register --name child --path child
git add child/.gitgroverepo
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

echo "All tests passed!"
