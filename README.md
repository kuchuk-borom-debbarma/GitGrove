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

---

## The Solution

GitGrove lets you work on **just one service** at a time, as if it were the entire repository:

```bash
gitgrove switch auth main
# Your terminal now shows ONLY auth/ files, flattened to root level!

git status           # Shows only auth changes
git log              # Shows only auth commits  
gitgrove cd ..       # Navigate back to parent
gitgrove ls          # See sibling services
```

---

## Key Features

- ✅ **Focused Context**: Only see what you're working on
- ✅ **Isolated History**: Each service has its own commit log
- ✅ **Filesystem Navigation**: Use `cd` and `ls` to move between services
- ✅ **100% Git Compatible**: Regular `git push`, `git pull`, `git merge` all work
- ✅ **No Submodules**: Everything stays in one repo

---

## Who Is This For?

GitGrove is ideal for:

- **Monorepo teams** who want service-level isolation without splitting into separate repositories
- **Developers** tired of filtering through irrelevant changes in `git status` and `git log`
- **Teams** managing microservices in a single repository
- **Projects** with nested components that need independent versioning
- **Anyone** who wants the benefits of polyrepos with the simplicity of a monorepo

---

## Quick Start

### Option 1: Interactive Mode (Recommended for Beginners)

```bash
# Build GitGrove
git clone https://github.com/kuchuk-borom-debbarma/GitGrove.git
cd GitGrove/cli
go build -o ../gitgrove ./cmd/main.go

# Run interactive mode
cd your-monorepo/
/path/to/GitGrove/gitgrove interactive
```

The interactive menu provides a guided experience for all GitGrove operations.

---

### Option 2: CLI Mode (For Advanced Users)

> **⚠️ IMPORTANT**: CLI mode requires adding `gitgrove` to your PATH first.

```bash
# Build and add to PATH
git clone https://github.com/kuchuk-borom-debbarma/GitGrove.git
cd GitGrove/cli
go build -o ../gitgrove ./cmd/main.go
export PATH=$PATH:$(pwd)/..

# Initialize and use
cd your-monorepo/
gitgrove init
gitgrove register --name auth --path services/auth
gitgrove switch auth main
```

**📖 [View Complete CLI Documentation](docs/cli.md)** - All commands, options, and advanced usage

---

## Documentation

### Getting Started
- **[Quick Start Guide](docs/quickstart.md)** - Step-by-step tutorial
- **[CLI Reference](docs/cli.md)** - Complete command documentation with examples
- **[Workflow Guide](docs/workflow.md)** - Daily usage patterns and best practices

### Technical Documentation
- **[Architecture](docs/architecture.md)** - Design overview and concepts
- **[Internals](docs/internals.md)** - Implementation details and how it works

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

---

