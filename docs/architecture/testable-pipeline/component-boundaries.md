# Component Boundaries — testable-pipeline

**Feature**: testable-pipeline
**Wave**: DESIGN
**Date**: 2026-03-18

---

## Component 1: Makefile (new)

**File**: `Makefile` (project root)
**Type**: New file
**Responsibility**: Canonical command sequences for all CI jobs and local developer workflows. Every command that CI or the developer runs is defined here once.

### Target Definitions

| Target | Mirrors CI Job | Command Sequence | Known Gaps |
|--------|---------------|-----------------|------------|
| `validate` | `validate-and-build` | cicd/check-versions.sh → go mod download → go-arch-lint check → go vet ./... → golangci-lint run → go test ./internal/... → go build -o kanban ./cmd/kanban | Binary output to ./kanban (not temp path); no persist_to_workspace |
| `acceptance` | `acceptance` | go test ./tests/acceptance/... with KANBAN_BINARY set to ./kanban | No workspace attach (assumes make validate was run first in same directory) |
| `ci` | validate-and-build + acceptance | make validate, then make acceptance (sequential, stop on first failure) | None |
| `release-snapshot` | `release` (local safe variant) | goreleaser release --snapshot --config cicd/goreleaser.yml --clean | No GITHUB_TOKEN; no tag push; no Homebrew update; produces dist/ locally |
| `release` | `release` | goreleaser release --config cicd/goreleaser.yml --clean | Requires GITHUB_TOKEN and HOMEBREW_TAP_GITHUB_TOKEN; invoked by CI only |
| `tag-dry` | `tag` (dry-run variant) | git fetch --unshallow --tags (or git fetch --tags); go-semver-release release --dry-run ... | Requires GITHUB_TOKEN for read access; requires full git history |
| `help` | — | Print formatted target list (default target when no args given) | — |

### Behavioral Constraints

- `make validate` must exit 0 on success and exit 1 on first failure, printing which step failed
- `make acceptance` must exit non-zero with an actionable error message if `./kanban` binary does not exist or is stale
- `make ci` must stop at `make validate` failure; `make acceptance` must not run if `make validate` fails
- `make release-snapshot` must fail gracefully with installation instructions if goreleaser is not found in PATH
- Each target includes an inline comment identifying the CI job it mirrors and any known gaps
- `make help` output must fit within 80 columns

### Version Sourcing

The Makefile does not hardcode tool versions. `cicd/check-versions.sh` (invoked as the first step of `make validate`) reads all versions dynamically from `cicd/config.yml` pipeline parameters. No duplicate version state in the Makefile.

---

## Component 2: cicd/config.yml changes

**File**: `cicd/config.yml`
**Type**: Existing file — targeted modifications
**Responsibility**: CircleCI pipeline definition. After this feature, it delegates command sequences to `make` and adds goreleaser caching.

### Change 1: Add install-goreleaser command

Add an `install-goreleaser` reusable command following the exact same pattern as `install-golangci-lint` and `install-go-arch-lint`:

```
Structure:
  restore_cache:
    keys:
      - goreleaser-<< pipeline.parameters.goreleaser-version >>
  run:
    if [ ! -f "$(go env GOPATH)/bin/goreleaser" ]; then
      go install github.com/goreleaser/goreleaser/v2@<< pipeline.parameters.goreleaser-version >>
    fi
  save_cache:
    key: goreleaser-<< pipeline.parameters.goreleaser-version >>
    paths:
      - /home/circleci/go/bin/goreleaser
```

Cache key is version-pinned. A version bump in `goreleaser-version` parameter invalidates the cache and triggers fresh installation. Uses `go install` (not `curl | bash`) — consistent with the existing pattern for other tools.

### Change 2: validate-and-build job — no change to command invocations

The `validate-and-build` job retains its existing direct command invocations. It does NOT call `make validate`.

Rationale: `cicd/check-versions.sh` is a local-only tool. It checks local tool installations against pipeline parameter values. In CI, tools are installed by the pipeline itself (`install-go-arch-lint`, `install-golangci-lint`) and goreleaser is not installed in this job at all. If `make validate` ran `check-versions.sh` inside CI's `validate-and-build` job, `check-versions.sh` would exit 1 for goreleaser (not installed) and break the job.

Parity guarantee mechanism: the Makefile `validate` target defines the same command sequence as the CI `validate-and-build` job. Both are authoritative. The Makefile is the developer's local interface; CI runs the sequence directly. Inline comments in both the Makefile and the CI job reference each other as the source of truth.

validate-and-build job: unchanged from current state.

### Change 3: acceptance job delegates to make acceptance

```
acceptance:
  executor: go-executor
  steps:
    - checkout
    - attach_workspace: { at: . }   ← kept: provides kanban binary from validate-and-build
    - go/mod-download               ← kept: ensures module cache
    - run: make acceptance          ← replaces Set Environment + E2E Tests run steps
```

`KANBAN_BINARY` (or `KANBAN_BIN` — the acceptance test suite accepts both; Makefile uses `KANBAN_BINARY`) is set inside the Makefile's `acceptance` target, pointing to `./kanban`. The CI job's explicit `Set Environment` step is replaced by the Makefile target. No version-check issue — the `acceptance` target does not invoke `check-versions.sh`.

### Change 4: release job uses install-goreleaser and make release

```
release:
  executor: go-executor
  steps:
    - checkout
    - run: git fetch --tags
    - go/mod-download
    - install-goreleaser           ← new: replaces curl | bash inline install
    - run:
        command: |
          export GITHUB_REPOSITORY_OWNER=$CIRCLE_PROJECT_USERNAME
          make release             ← replaces direct goreleaser invocation
```

### Change 5: [skip release] shell guards on tag and release jobs

Add a shell guard at the start of the `Compute and push semver tag` run step in the `tag` job and the goreleaser run step in the `release` job:

```
# Guard: skip if commit message contains [skip release] (case-insensitive)
COMMIT_MSG=$(git log -1 --pretty=%B)
if echo "$COMMIT_MSG" | grep -qi '\[skip release\]'; then
  echo "[skip release] detected — skipping tag/release"
  exit 0
fi
```

This guard is inserted after the existing "already tagged" guard in the `tag` job (so the two guards are additive, not conflicting). The guard is identical in both `tag` and `release` jobs.

Both jobs still appear in the CircleCI UI (they run to completion with exit 0), providing an audit trail that they ran but skipped intentionally.

An inline comment in `cicd/config.yml` explains the `[skip release]` convention. One sentence in `README.md` or `CONTRIBUTING.md` documents the convention for the developer.

---

## Component 3: cicd/pre-commit changes

**File**: `cicd/pre-commit`
**Type**: Existing file — targeted modifications
**Responsibility**: Quality gate run on every git commit. After this feature, it matches the CI validate-and-build job exactly (modulo test scope difference: `./...` vs `./internal/...`).

### Change 1: Add step 0 — check-versions.sh

Insert before step 1 (go test):

```
# Step 0: Tool version check (prerequisite gate)
echo "[0/5] check-versions"
if ! cicd/check-versions.sh; then
    echo "FAIL [0/5] check-versions — update local tools to match cicd/config.yml before committing"
    exit 1
fi
echo "PASS [0/5] check-versions"
```

Step 0 runs first. If it exits non-zero, the hook exits 1 and no subsequent steps run. The exit is actionable — the error output from `check-versions.sh` identifies which tool is mismatched and what the expected version is.

### Change 2: Uncomment go-arch-lint block (lines 43–49)

Remove the `#` comment characters from the 6 lines comprising the go-arch-lint step. The code is correct; only the comment markers are removed.

After uncommenting, the step reads:
```sh
# Step 3: Architecture enforcement
echo "[3/5] go-arch-lint check"
if ! go-arch-lint check; then
    echo "FAIL [3/5] go-arch-lint — fix architecture violations before committing"
    exit 1
fi
echo "PASS [3/5] go-arch-lint"
```

### Step numbering after changes

| Step label | Command | Notes |
|------------|---------|-------|
| [0/5] | cicd/check-versions.sh | New — prerequisite gate |
| [1/5] | go test ./... | Unchanged |
| [2/5] | golangci-lint run | Unchanged |
| [3/5] | go-arch-lint check | Restored (was commented out) |
| [4/5] | go build | Unchanged |
| [5/5] | acceptance tests | Unchanged |

Total: 6 steps labeled [0/5] through [5/5]. The `/5` denominator refers to the five quality gates (not counting the prerequisite check), preserving the existing visual convention while adding the step 0 label.

### Behavioral constraint

If `go-arch-lint` is not installed, `check-versions.sh` (step 0) detects this and exits 1 with a message directing the developer to install it. Step 3 is never reached in this case — no silent skip is possible.

---

## Component 4: [skip release] convention documentation

**Scope**: Inline comment in `cicd/config.yml` (added in Component 2) + one-sentence note in project README or CONTRIBUTING.

**Convention**: When a commit message on `main` contains `[skip release]` (case-insensitive, position-independent), the `tag` and `release` CI jobs run but exit 0 early without creating a tag, pushing a release, or updating the Homebrew formula.

**Not skipped**: `validate-and-build` and `acceptance` always run on every push to main.

**Recovery when accidentally included**: Push an empty commit without the flag: `git commit --allow-empty -m "ci: trigger release"`.

---

## Unchanged Components

| Component | Reason |
|-----------|--------|
| cicd/check-versions.sh | Already fully implemented; reads versions from cicd/config.yml dynamically; exits 1 on mismatch. No change required. |
| cicd/goreleaser.yml | No changes to release configuration required by this feature. |
| cicd/install-hooks.sh | Copies cicd/pre-commit to .git/hooks/pre-commit; this mechanism is unchanged. Developers re-run install-hooks.sh after US-TP-03 is shipped to pick up the updated pre-commit. |
| internal/ (application code) | This feature is entirely CI/CD infrastructure. No application code changes. |
