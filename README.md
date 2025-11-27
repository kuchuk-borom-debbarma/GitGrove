
# **GitGrove**

### *Multi-Repository Version Control on Top of Git*

GitGrove is a lightweight layer built on top of Git that lets you treat a single Git repository as **many independent nested repositories** — each with its own commit history, branches, and lifecycle.

To Git, everything remains a **normal repository**.
To GitGrove, your project becomes a **structured forest**, where each directory can behave like its own repo with isolated changes, logs, and branches — all while sharing the same Git backend.

---

## What GitGrove Gives You

### **✔ Nested Repositories Inside One Git Repo**

Define multiple logical repos (e.g., `backend`, `frontend`, `docs`) inside the same working tree.

### **✔ Independent Commit Histories**

Commit only the files belonging to a specific nested repo — GitGrove keeps the histories separate using lightweight branch references.

### **✔ Branch Hierarchy & Repo Structure**

Create branches per nested repo that reflect your architecture, dependencies, and workspace layout.

### **✔ Metadata Synchronization**

GitGrove uses a special metadata branch to automatically sync configuration and repo definitions across all Git branches.

### **✔ Multi-Repo Snapshots**

Capture consistent states of multiple nested repos at once, using Git’s reliable merge and branching engine.

### **✔ 100% Compatible With Git**

Nothing in your repository breaks normal Git commands.
You can `git pull`, `git push`, `git checkout`, and everything will behave as expected.

---

## Why GitGrove?

Managing many small services, modules, or components inside a single Git repository can be messy:

* changes mix together
* histories become unreadable
* branch structures don’t reflect your architecture

GitGrove solves this by giving each folder **its own identity** — without submodules, subtrees, or monorepo tooling complexity.

You get:

* the clarity of multiple repos
* the simplicity of a single Git repository
* and the power of Git behind everything you do

---

## Who Is GitGrove For?

* Multi-service or multi-package projects
* Modular applications
* Monorepos that need independent histories
* Developers who want structure without extra tooling
* Anyone who wants repo-within-repo behavior without fighting Git

---

