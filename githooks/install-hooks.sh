#!/bin/sh
# install-hooks.sh — install developer quality gate hooks for kanban-tasks
#
# Usage: githook/install-hooks.sh
#
# Installs:
#   .git/hooks/pre-commit  — runs all CI quality gates before each commit
#   .git/hooks/commit-msg  — validates commit message follows conventional commits
#
# This script is idempotent. Running it multiple times is safe.

set -e

REPO_ROOT="$(git rev-parse --show-toplevel)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

# Install pre-commit hook
echo "Installing pre-commit hook..."
cp "$REPO_ROOT/githooks/pre-commit" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"
echo "Installed: .git/hooks/pre-commit"

# Install commit-msg hook
echo "Installing commit-msg hook..."
cp "$REPO_ROOT/githooks/commit-msg" "$HOOKS_DIR/commit-msg"
chmod +x "$HOOKS_DIR/commit-msg"
echo "Installed: .git/hooks/commit-msg"