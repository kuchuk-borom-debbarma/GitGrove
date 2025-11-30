# GitGrove Workflow Guide

This guide explains the standard development workflow using GitGrove, detailing how to manage repositories, branches, and commits within your monorepo.

## 1. Initialization

Start by initializing GitGrove in your project root. This sets up the `gitgroove/system` branch to track metadata.

```bash
# Initialize GitGrove (must be inside a git repo)
git init
gitgrove init
```

---

## 2. Registering a Repository

To add a new logical repository (e.g., `backend`) to GitGrove:

```bash
# 1. Register the repo
gitgrove register --name backend --path ./backend
```

### What happens behind the scenes?
1.  **Marker Creation**: GitGrove creates the `./backend` directory (if missing) and writes a `.gitgroverepo` file containing the name "backend".
2.  **Metadata Update**: The repo's path is recorded in the `gitgroove/system` branch.
3.  **Branch Seeding**: An orphan branch `gitgroove/repos/backend/branches/main` is created. Crucially, **this branch is seeded with the `.gitgroverepo` file** so that switching to it doesn't delete the directory.

### Important: Commit the Marker!
After registering, you will see an untracked file: `backend/.gitgroverepo`. **You must commit this to your current branch (e.g., `main`)**.

```bash
git add backend/.gitgroverepo
git commit -m "Register backend repo"
```

> **Why?** If you don't commit it, `main` doesn't know about the file. When you switch to the `backend` branch (which *does* have the file), Git might complain or handle it unpredictably. Keeping it tracked in both branches ensures smooth switching.

---

## 3. Working on a Repository

To work on the `backend` repo, you switch to its branch.

```bash
gitgrove switch backend
```

### The "Magic" Switch
When you switch:
- Git checks out `gitgroove/repos/backend/branches/main`.
- **Your working directory changes**:
    - Files belonging to `backend` appear (initially just `.gitgroverepo`).
    - Files from `main` (like `README.md` or other repos) are removed *from the working tree* (don't worry, they are safe in the `main` branch).
- The `backend/` directory persists because it exists in both your previous branch (you committed it!) and the new branch (GitGrove seeded it!).

### Development Cycle
Now you are "inside" the backend repo context.

```bash
# Create files
echo "package main" > backend/main.go

# Stage files (GitGrove validates they are inside 'backend/')
gitgrove stage backend/main.go

# Commit (GitGrove ensures you are on the correct branch)
gitgrove commit -m "Initial backend logic"
```

---

## 4. Linking Repositories

If `frontend` depends on `backend`, you link them.

```bash
gitgrove link --child frontend --parent backend
```

- This updates metadata in `gitgroove/system`.
- It does **not** affect your source code or branches. It's purely for tooling (e.g., CI/CD pipelines that need to know build order).

---

## 5. Moving a Repository

One of GitGrove's superpowers is moving repos without breaking history.

```bash
gitgrove move --repo backend --to ./services/backend
```

### What happens?
1.  **Physical Move**: The directory `./backend` is renamed to `./services/backend`.
2.  **Metadata Update**: `gitgroove/system` is updated with the new path.
3.  **History Preserved**: The `backend` branch is **untouched**. It still contains the same files. Because GitGrove tracks "repo identity" via the marker and metadata, the physical location is just an implementation detail.

---

## Summary of Commands

| Command | Action | Key Side Effect |
| :--- | :--- | :--- |
| `init` | Setup GitGrove | Creates `gitgroove/system` branch |
| `register` | Add new repo | Creates marker & seeded orphan branch |
| `switch` | Checkout repo branch | Updates working tree to repo content |
| `stage` | `git add` wrapper | Enforces repo boundaries |
| `commit` | `git commit` wrapper | Commits to repo branch |
| `link` | Define dependency | Updates system metadata |
| `move` | Rename directory | Updates system metadata & moves files |
