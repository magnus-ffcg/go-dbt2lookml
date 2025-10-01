# Git Hooks

This directory contains Git hooks for the project.

## Setup

### Option 1: Configure Git to use this directory

```bash
git config core.hooksPath .githooks
```

This will make Git use all hooks in this directory automatically.

### Option 2: Symlink individual hooks

```bash
ln -s ../../.githooks/pre-commit .git/hooks/pre-commit
```

## Available Hooks

### pre-commit

Runs before each commit to ensure code quality:
- ✅ Go formatting check (`gofmt`)
- ✅ Go vet
- ✅ golangci-lint (if installed)
- ✅ Quick tests

If any check fails, the commit is aborted.

## Bypass Hook (Not Recommended)

If you need to bypass the pre-commit hook in an emergency:

```bash
git commit --no-verify
```

**Note:** This should be avoided as it skips quality checks.
