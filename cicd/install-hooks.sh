#!/bin/sh
# install-hooks.sh — install developer quality gate hooks for kanban-tasks
#
# Usage: cicd/install-hooks.sh
#
# Installs:
#   .git/hooks/pre-commit — runs all CI quality gates before each commit
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

echo ""
echo "Pre-commit hook installed. It will run on every git commit:"
echo "  go test ./... | golangci-lint | go-arch-lint | go build | acceptance tests"
echo ""
echo "NOTE: To also install the kanban commit-msg hook in this repo (for"
echo "task auto-transitions), run:"
echo "  kanban install-hook"
echo ""
echo "The commit-msg hook (kanban install-hook) and the pre-commit hook"
echo "(this script) are independent. Both are recommended for development."
