#!/bin/bash
set -e

# Setup
mkdir -p test_env
cd test_env
rm -rf .git .gg .gitgroverepo backend frontend
git init
../gitgrove init

# Create dummy dirs
mkdir backend frontend
touch backend/file.txt frontend/file.txt
git add .
git commit -m "Initial structure"

echo "--- Testing Register with Quotes ---"
../gitgrove register 'backend;./backend' 'frontend;./frontend'
git add .
git commit -m "Add repo markers"

echo "--- Testing Link with Quotes ---"
../gitgrove link 'frontend;backend'

echo "--- Testing Stage . ---"
../gitgrove switch backend main
cd backend
echo "new content" > newfile.txt
../../gitgrove stage .

echo "--- Verification Complete ---"
