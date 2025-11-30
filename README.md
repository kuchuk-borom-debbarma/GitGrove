# GitGrove

**GitGrove** is a Git-based monorepo management tool that enables hierarchical repository organization within a single Git repository. It allows you to work with nested logical repositories while maintaining a unified version control history.

## What is GitGrove?

GitGrove transforms a standard Git repository into a structured workspace where you can:

- **Organize code hierarchically**: Define parent-child relationships between logical repositories
- **Work in isolation**: Switch between repositories and see only relevant files
- **Maintain unified history**: All changes are tracked in a single Git repository
- **Preserve Git workflows**: Use familiar Git commands with GitGrove-specific enhancements

## Key Concepts

- **Repositories**: Logical subdivisions of your project (e.g., `backend`, `frontend`, `shared`)
- **Hierarchy**: Parent-child relationships between repositories (e.g., `shared` → `backend` → `feature-service`)
- **Flattened View**: When you switch to a repository, you see only its contents at the root level
- **System Branch**: A special `gitgroove/system` branch stores GitGrove metadata

## Quick Example

```bash
# Initialize GitGrove in your Git repository
gg init

# Register repositories
gg register backend=./services/backend
gg register frontend=./services/frontend
gg register shared=./libs/shared

# Create hierarchy (shared is parent of backend and frontend)
gg link backend=shared
gg link frontend=shared

# Switch to backend repository
gg switch backend

# Work normally - you see only backend files at root level
gg add .
gg commit -m "Update backend"

# Navigate to parent
gg up

# List children
gg ls
```

## Documentation

- **[Quick Start Guide](docs/QUICKSTART.md)**: Get up and running in minutes
- **[Full Documentation](docs/FULL_DOCUMENTATION.md)**: Comprehensive guide covering all features
- **[Architecture Documentation](docs/architecture.md)**: Detailed system design and implementation patterns
- **[Technical Documentation](docs/TECHNICAL.md)**: Low-level implementation details

## Installation

```bash
# Build from source (requires Go 1.25.4+)
git clone https://github.com/kuchuk-borom-debbarma/GitGrove
cd GitGrove
go build -o gg ./cmd/gg
```

## Prerequisites

- Git installed and configured
- A Git repository (initialize with `git init` if needed)
- Clean working tree (no uncommitted changes)

## Core Commands

| Command | Description |
|---------|-------------|
| `gg init` | Initialize GitGrove in current repository |
| `gg register <name>=<path>` | Register a repository |
| `gg link <child>=<parent>` | Create parent-child relationship |
| `gg switch <repo> [branch]` | Switch to a repository |
| `gg add <files>` | Stage changes with validation |
| `gg commit -m <msg>` | Commit with validation |
| `gg up` | Navigate to parent repository (or System Root) |
| `gg down <child>` | Navigate to child repository |
| `gg cd <target>` | Navigate to repo, parent (`..`), or System Root (`~`) |
| `gg ls` | List child repositories |
| `gg info` | Show project status |
| `gg push <repo\|*>` | Push repositories to remote |

## When to Use GitGrove

**Good fit:**
- Monorepos with logical subdivisions
- Projects with shared libraries and dependent services
- Teams wanting isolation without multiple repositories

**Not ideal for:**
- Simple single-service projects
- Repositories with many large binary files

