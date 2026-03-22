#!/bin/sh
# install-hooks.sh — install developer quality gate hooks for kanban-tasks
#
# Usage: cicd/install-hooks.sh
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
cp "$REPO_ROOT/cicd/pre-commit" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"
echo "Installed: .git/hooks/pre-commit"

# Install commit-msg hook
echo "Installing commit-msg hook..."
cp "$REPO_ROOT/cicd/commit-msg" "$HOOKS_DIR/commit-msg"
chmod +x "$HOOKS_DIR/commit-msg"
echo "Installed: .git/hooks/commit-msg"

echo ""
echo "Hooks installed:"
echo "  pre-commit  — gotestsum | golangci-lint | go-arch-lint | go build | acceptance tests"
echo "  commit-msg  — validates conventional commits format (feat|fix|chore|...)"
echo ""
echo "NOTE: 'kanban install-hook' installs a separate commit-msg hook in user"
echo "repos for task auto-transitions. Do not run it in the kanban-tasks repo —"
echo "it would overwrite the conventional commits validator above."
