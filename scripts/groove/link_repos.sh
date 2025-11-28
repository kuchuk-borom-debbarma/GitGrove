#!/bin/bash
set -e

# Get the root of the repo
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT/build/demo"

echo "Linking repositories..."

# Link service-a -> backend
# Link service-b -> backend
# Link web-app -> frontend
# But wait, 'backend' and 'frontend' are not registered repos?
# The demo setup created:
# repo/backend/service-a
# repo/backend/service-b
# repo/frontend/web-app
# And registered:
# service-a -> repo/backend/service-a
# service-b -> repo/backend/service-b
# web-app -> repo/frontend/web-app

# To link them, I need parent repos.
# I should register 'backend' and 'frontend' first?
# Or maybe the user intends to link service-a to something else?
# The demo structure implies:
# root -> backend -> service-a
# root -> backend -> service-b
# root -> frontend -> web-app

# So I need to register 'backend' and 'frontend' as repos too?
# But they are directories.
# Let's register them.

echo "Registering intermediate repos..."
./gitgrove register --name backend --path repo/backend
./gitgrove register --name frontend --path repo/frontend

echo "Linking..."
./gitgrove link --child service-a --parent backend
./gitgrove link --child service-b --parent backend
./gitgrove link --child web-app --parent frontend

echo "Linking complete."
