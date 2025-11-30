# GitGrove

**Navigate your monorepo like a filesystem. Version it like Git.**

GitGrove transforms a single Git repository into a collection of virtual repositories. Each service or component gets its own isolated workspace, commit history, and navigation—all while sharing one unified Git history underneath.

---

## The Problem

You have a monorepo with multiple services:
```
ecommerce/
├── api/              # Your REST API
├── web/              # React frontend  
└── services/
    ├── auth/         # Authentication service
    ├── payments/     # Payment processor
    └── inventory/    # Stock management
```

**Working in this structure is painful:**
- `git status` shows 100 files, but you only care about auth
- Your commit history mixes changes from 5 different teams
- You can't easily see what changed in just the payments service
- Switching contexts means mentally filtering irrelevant changes

## The Solution

GitGrove lets you work on **just one service** at a time, as if it were the entire repository:

```bash
gitgrove switch auth main
# Your terminal now shows ONLY auth/ files, flattened to root level!
# The working directory looks like:
# .
# ├── login.go
# ├── token.go
# └── .gitgroverepo

git status           # Shows only auth changes
git log              # Shows only auth commits  
gitgrove cd ..       # Navigate back to parent
gitgrove ls          # See sibling services
```

**Benefits:**
- ✅ **Focused Context**: Only see what you're working on
- ✅ **Isolated History**: Each service has its own commit log
- ✅ **Filesystem Navigation**: Use `cd` and `ls` to move between services
- ✅ **100% Git Compatible**: Regular `git push`, `git pull`, `git merge` all work
- ✅ **No Submodules**: Everything stays in one repo

---

## Quick Start (5 Minutes)

### 1. Install

```bash
git clone https://github.com/kuchuk-borom-debbarma/GitGrove.git
cd GitGrove/cli
go build -o ../gitgrove ./cmd/main.go

# Optional: Add to PATH 
export PATH=$PATH:$(pwd)/..
```

### 2. Set Up Your First Virtual Repository

```bash
cd your-monorepo/

# Assumes that you added gitgrove to path
# Initialize GitGrove (one-time setup)
gitgrove init

# Register your first service
gitgrove register --name auth --path services/auth

# Switch to it
gitgrove switch auth main
```

**What just happened?**
- `init` created a hidden `gitgroove/system` branch to store metadata
- `register` told GitGrove "treat `services/auth` as its own repo"
- `switch` checked you out to a special branch where `services/auth` appears at root level

### 3. Work Normally

```bash
# You're now "inside" the auth service
ls                           # Shows: login.go, token.go, etc.
echo "// fix" >> login.go    
gitgrove add .               # Stage files (only from auth)
gitgrove commit -m "Fix bug" # Commit (only affects auth)
```

**Your commit only touched `services/auth`. The rest of the monorepo is untouched.**

---

## Complete Setup Example

Let's set up the full e-commerce monorepo:

### Step 1: Initialize
```bash
cd ecommerce/
gitgrove init
```
This creates the internal system branch. You only do this once per repository.

### Step 2: Register All Services

**Option A: One at a time (verbose but clear)**
```bash
gitgrove register --name api --path api
gitgrove register --name web --path web
gitgrove register --name services --path services
gitgrove register --name auth --path services/auth
gitgrove register --name payments --path services/payments
gitgrove register --name inventory --path services/inventory
```

**Option B: Batch register (faster)**
```bash
# Register top-level services
gitgrove register --name api --path api
gitgrove register --name web --path web
gitgrove register --name services --path services

# Batch register nested services (note the quotes!)
gitgrove register "auth;services/auth" "payments;services/payments" "inventory;services/inventory"
```

> **Semicolon Syntax**: When using `name;path` format, you must quote each argument. Each one creates a separate virtual repo.

### Step 3: Create Hierarchy (Optional but Recommended)

Tell GitGrove which services are related:

```bash
gitgrove link --child auth --parent services
gitgrove link --child payments --parent services  
gitgrove link --child inventory --parent services
```

Or batch:
```bash
gitgrove link "auth;services" "payments;services" "inventory;services"
```

**Why link?** It enables hierarchical navigation with `cd` and `ls`.

### Step 4: Navigate and Work

**See where you are:**
```bash
gitgrove info
```
Output shows your repo tree:
```
└── services
    ├── * auth [CURRENT]
    ├── payments
    └── inventory
```

**Navigate like a filesystem:**
```bash
gitgrove ls                  # List children: auth, payments, inventory
gitgrove cd auth             # Move into auth
gitgrove cd ..               # Move up to services
gitgrove cd payments         # Move to sibling
```

**Switch explicitly:**
```bash
gitgrove switch payments main      # Jump directly to payments/main
gitgrove switch auth feature-x     # Jump to a specific branch
```

---

## Daily Workflow

### Making Changes in a Service

```bash
# Navigate to the service
gitgrove switch auth main

# Make changes
vim login.go

# Stage and commit (scoped to auth only)
gitgrove add .
gitgrove commit -m "Add OAuth support"

# The commit ONLY affects services/auth
# Other services are completely untouched
```

### Working Across Multiple Services

```bash
# Fix a bug in auth
gitgrove switch auth main
vim token.go
gitgrove add .
gitgrove commit -m "Fix token expiry"

# Update the API that uses auth
gitgrove cd ..        # Go up
gitgrove cd ..        # Go to root
gitgrove cd api       # Go to API
vim middleware.go
gitgrove add .
gitgrove commit -m "Use new auth tokens"
```

**Each service maintains its own clean commit history.**

### Creating Feature Branches

```bash
# Create a feature branch for payments
gitgrove branch payments feature-stripe-integration

# Switch to it
gitgrove switch payments feature-stripe-integration

# Work on it
# ...commits happen only in payments...
```

---

## Command Reference

### Setup Commands

| Command | What It Does | Example |
|---------|--------------|---------|
| `init` | Set up GitGrove in your repo (one time) | `gitgrove init` |
| `register` | Turn a folder into a virtual repo | `gitgrove register --name auth --path services/auth` |
| `link` | Define parent-child relationships | `gitgrove link --child auth --parent services` |

### Navigation Commands

| Command | What It Does | Example |
|---------|--------------|---------|
| `info` | Show repo tree + your location | `gitgrove info` |
| `ls` | List children of current repo | `gitgrove ls` |
| `cd <name>` | Navigate to a child service | `gitgrove cd auth` |
| `cd ..` | Navigate to parent | `gitgrove cd ..` |
| `switch <repo> <branch>` | Jump directly to any repo/branch | `gitgrove switch auth main` |

### Development Commands

| Command | What It Does | Example |
|---------|--------------|---------|
| `add <files>` | Stage files (scoped to current repo) | `gitgrove add .` |
| `commit -m "msg"` | Commit staged changes | `gitgrove commit -m "Fix"` |
| `branch <repo> <name>` | Create a new branch | `gitgrove branch auth feature-x` |

### Other Commands

| Command | What It Does |
|---------|--------------|
| `interactive` | Menu-driven interface for all operations |
| `move --repo <name> --to <path>` | Relocate a virtual repo |

---

## How It Works (Under the Hood)

GitGrove uses Git's plumbing commands to create special "flattened" branches:

1. **Normal branch**: `gitgroove/system` stores metadata
2. **Virtual repo branch**: `gitgroove/repos/auth/branches/main` 
   - The tree at `services/auth` becomes the **root** tree
   - When you check this out, you see `services/auth` contents at your working directory root

**Key insight**: GitGrove doesn't move files. It creates Git branch references where the subdirectory IS the root. Git's checkout mechanism handles the rest.

For deep technical details, see [docs/internals.md](docs/internals.md).

---

## FAQ

**Q: Is this better than Git submodules?**  
A: Different use case. Submodules are separate repos with separate histories. GitGrove keeps one unified history but gives you isolated *views* of different parts.

**Q: Can I still use regular Git commands?**  
A: Yes! `git push`, `git pull`, `git merge` all work normally. GitGrove is strictly additive.

**Q: What if I make a mistake?**  
A: Your original `main` branch is never touched. GitGrove only creates new branch refs. You can always `git checkout main` to see your original structure.

**Q: Does it work with large monorepos?**  
A: Yes. GitGrove is just Git branches, so it scales as well as Git does.

**Q: Can I mix GitGrove commands and git commands?**  
A: Yes, but use `gitgrove add` and `gitgrove commit` for safety. They enforce boundaries so you don't accidentally commit files from the wrong service.

---

## Next Steps

- **New to GitGrove?** Try the [Quick Start Guide](docs/quickstart.md)
- **Want to understand internals?** Read [Technical Documentation](docs/internals.md)
- **Ready to integrate?** Start with `gitgrove interactive` for a guided experience

---

