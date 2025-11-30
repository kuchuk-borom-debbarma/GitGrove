# GitGrove

*Manage monorepos like a filesystem, version them like Git.*

GitGrove lets you work on specific parts of a monorepo in **isolation**, while maintaining a unified Git history. Think of it as `cd` and `ls` for your repository structure, with Git's power underneath.

---

## Quick Example

```bash
# You have a monorepo with this structure:
# .
# ├── api/
# ├── web/
# └── services/
#     ├── auth/
#     └── payments/

# With GitGrove, you can:
gitgrove ls                    # List repos at current level
gitgrove cd services/auth      # Jump into auth service
gitgrove info                  # See where you are in the tree
gitgrove cd ..                 # Go back up
```

---

## Why GitGrove?

**The Problem:** Large monorepos are noisy. Running `git status` shows 100 changed files, but you only care about 3. Every commit touches multiple services. Finding histor is hard.

**The Solution:** GitGrove creates **virtual repositories** within your monorepo. Each service gets:
- ✅ Its own isolated workspace
- ✅ Its own commit history  
- ✅ Flattened view (work on `auth/` as if it's the repo root)
- ✅ 100% Git compatibility underneath

---

## Installation

```bash
git clone https://github.com/kuchuk-borom-debbarma/GitGrove.git
cd GitGrove/cli
go build -o ../gitgrove ./cmd/main.go

# Optionally add to PATH
export PATH=$PATH:$(pwd)/..
```

---

## Getting Started

### 1. Initialize in Your Repo

```bash
cd your-monorepo/
gitgrove init
```

This creates a `gitgroove/system` branch to track metadata.

### 2. Register Your Services

Let's say you have a microservices app:

```
ecommerce/
├── api/              # REST API
├── web/              # React frontend
└── services/
    ├── auth/         # Authentication
    ├── payments/     # Payment processing
    └── inventory/    # Stock management
```

Register them:

```bash
gitgrove register --name api --path api
gitgrove register --name web --path web
gitgrove register --name services --path services
gitgrove register --name auth --path services/auth
gitgrove register --name payments --path services/payments
gitgrove register --name inventory --path services/inventory
```

**Batch Registration (Semicolon Syntax):**
```bash
# Register multiple at once - note the quotes!
gitgrove register "auth;services/auth" "payments;services/payments" "inventory;services/inventory"
```

> **Note**: When using semicolon syntax, each argument must be quoted and in `name;path` format.

### 3. Define Hierarchy (Optional)

Link repos to create a tree:

```bash
gitgrove link --child auth --parent services
gitgrove link --child payments --parent services
gitgrove link --child inventory --parent services
```

Or batch link with semicolon syntax:
```bash
gitgrove link "auth;services" "payments;services" "inventory;services"
```

### 4. Start Working

**Switch to a service:**
```bash
gitgrove switch auth main
# Your working directory now shows ONLY auth/ files at the root!
```

**Navigate like a filesystem:**
```bash
gitgrove ls              # Shows: (no children) - auth is a leaf
gitgrove cd ..           # Go up to 'services'
gitgrove ls              # Shows: auth, payments, inventory
gitgrove cd payments     # Jump to payments service
```

**See where you are:**
```bash
gitgrove info
```
Output:
```
GitGrove Info
=============
...
Registered Repositories:
------------------------
└── services
    ├── auth
    ├── inventory
    └── * payments [CURRENT]  ← You are here
```

**Make changes:**
```bash
# Edit some files
echo "new code" >> processor.go

gitgrove add .
gitgrove commit "Add payment retry logic"
# This commit ONLY affects the payments service!
```

---

## Command Reference

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Initialize GitGrove | `gitgrove init` |
| `info` | Show repo tree and current location | `gitgrove info` |
| `register` | Register a directory as a repo | `gitgrove register --name auth --path services/auth` |
| `link` | Define parent-child relationship | `gitgrove link --child auth --parent services` |
| `switch` | Switch to a repo branch | `gitgrove switch auth main` |

### Navigation

| Command | Description | Example |
|---------|-------------|---------|
| `ls` | List child repositories | `gitgrove ls` |
| `cd <repo>` | Navigate to a child repo | `gitgrove cd auth` |
| `cd ..` | Navigate to parent repo | `gitgrove cd ..` |

### Development

| Command | Description | Example |
|---------|-------------|---------|
| `add` | Stage files (scoped to current repo) | `gitgrove add .` |
| `commit` | Commit changes | `gitgrove commit "Fix bug"` |
| `branch` | Create a repo-specific branch | `gitgrove branch auth feature-x` |
| `move` | Relocate a repo | `gitgrove move auth services/auth-v2` |

### Interactive Mode

```bash
gitgrove interactive
```

Launches a menu-driven interface with hierarchical navigation for all operations.

---

## Real-World Workflow

**Scenario:** You need to fix a bug in the auth service, then update the API that uses it.

```bash
# Start at project root
cd ecommerce/

# Navigate to auth service
gitgrove switch auth main
# (Working directory now contains only auth/ files)

# Make your changes
vim token_validator.go
gitgrove add .
gitgrove commit "Fix token expiry check"

# Move to API
gitgrove cd ..               # Go up to 'services'
gitgrove cd ..               # Go up to root
gitgrove cd api              # Go to API
# (Working directory now shows only api/ files)

# Update API code
vim auth_middleware.go
gitgrove add .
gitgrove commit "Update to use new token validator"

# See your isolated histories
gitgrove info
```

---

## How It Works

GitGrove uses Git's plumbing to create "flattened" branches where a subdirectory becomes the root. When you `gitgrove switch payments main`, you're checking out `gitgroove/repos/payments/branches/main` — a special branch where `services/payments/` is the tree root.

For technical details, see [docs/internals.md](docs/internals.md).

---

## FAQ

**Q: Is this just Git submodules?**  
A: No. Submodules are separate Git repos. GitGrove keeps everything in one repo with one unified history, but gives you isolated views.

**Q: Can I still use regular Git commands?**  
A: Yes! `git push`, `git pull`, `git merge` all work normally. GitGrove is a layer on top.

**Q: What happens to my existing history?**  
A: Nothing. GitGrove only creates new branch refs. Your main branch is untouched.

---

## License

MIT
