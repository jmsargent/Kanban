#!/usr/bin/env bash
# cicd/check-versions.sh — verify local tool versions match the CI pipeline.
# Usage: ./cicd/check-versions.sh

set -euo pipefail

CONFIG="$(dirname "$0")/config.yml"
PASS=0
FAIL=0

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }

check() {
  local tool="$1"
  local expected="$2"
  local actual="$3"

  # Strip leading 'v' for comparison
  local e="${expected#v}"
  local a="${actual#v}"

  if [[ "$a" == "$e"* ]]; then
    green "  OK  $tool: $actual"
    PASS=$((PASS + 1))
  else
    red  " FAIL $tool: local=$actual  pipeline=$expected"
    FAIL=$((FAIL + 1))
  fi
}

# Extract a pipeline parameter default value from the CircleCI config.
param() {
  # Matches:  <param-name>:         (possibly with spaces, on one line)
  #             type: string
  #             default: "value"
  awk "/^  $1:/{found=1} found && /default:/{gsub(/[\" ]/, \"\"); split(\$0, a, \":\"); print a[2]; exit}" "$CONFIG"
}

echo "Checking local tool versions against $CONFIG …"
echo

# Go
EXPECTED_GO="$(param go-image-version)"
ACTUAL_GO="$(go version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1 || true)"
check "go" "$EXPECTED_GO" "${ACTUAL_GO:-not installed}"

# golangci-lint
EXPECTED_GOLANGCI="$(param golangci-lint-version)"
ACTUAL_GOLANGCI="$(golangci-lint --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)"
check "golangci-lint" "$EXPECTED_GOLANGCI" "${ACTUAL_GOLANGCI:-not installed}"

# go-arch-lint (uses 'version' subcommand, not '--version')
EXPECTED_ARCHLINT="$(param go-arch-lint-version)"
ACTUAL_ARCHLINT="$(go-arch-lint version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)"
check "go-arch-lint" "$EXPECTED_ARCHLINT" "${ACTUAL_ARCHLINT:-not installed}"

# go-semver-release (CI-only — used in tag job; skip check if not installed locally)
if command -v go-semver-release >/dev/null 2>&1; then
  EXPECTED_SEMVER="$(param go-semver-release-version)"
  ACTUAL_SEMVER="$(go-semver-release --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)"
  check "go-semver-release" "$EXPECTED_SEMVER" "${ACTUAL_SEMVER:-not installed}"
fi

# goreleaser
EXPECTED_GORELEASER="$(param goreleaser-version)"
ACTUAL_GORELEASER="$(goreleaser --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)"
check "goreleaser" "$EXPECTED_GORELEASER" "${ACTUAL_GORELEASER:-not installed}"

echo
if [[ $FAIL -eq 0 ]]; then
  green "All $PASS tools match the pipeline. ✓"
else
  red "$FAIL mismatch(es) — update local tools or the parameters in $CONFIG."
  exit 1
fi
