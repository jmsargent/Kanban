# Evolution Document — testable-pipeline

**Date**: 2026-03-18
**Feature ID**: testable-pipeline
**Status**: Delivered — all 4 steps complete with COMMIT/PASS

---

## Feature Summary

The testable-pipeline feature makes the CI pipeline locally verifiable. Prior to
this feature, the developer (Jon) was using live CI pushes as the primary test
environment for pipeline configuration changes — a 5–15 minute feedback loop
that produced spurious GitHub Releases on every failed config attempt.

Four implementation steps were delivered:

1. **01-01: Pre-commit gate parity** — added `cicd/check-versions.sh` as step
   `[0/5]` (prerequisite version gate) and uncommented the `go-arch-lint` block
   at step `[3/5]`, which had been commented out since initial project setup.

2. **01-02: Walking skeleton Makefile** — created the Makefile with the `validate`
   target spanning steps `[0/5]` through `[4/5]`, establishing the parity
   contract between local execution and CI's `validate-and-build` job.

3. **01-03: Complete Makefile** — extended to all 7 targets: `validate`,
   `acceptance`, `ci`, `release-snapshot`, `release`, `tag-dry`, and `help`.
   Standardized the `KANBAN_BIN` environment variable convention across the
   Makefile, `cicd/config.yml`, and `cicd/pre-commit` (was `KANBAN_BINARY` in
   pre-commit).

4. **01-04: CI config goreleaser caching and `[skip release]` guard** — added
   an `install-goreleaser` reusable command to `cicd/config.yml` using
   `go install` + CircleCI cache (replacing `curl | bash`), and added a shell
   guard to the `tag` and `release` jobs that exits 0 early when the commit
   message contains `[skip release]`.

---

## Key Decisions

| ADR | Title | Outcome |
|-----|-------|---------|
| ADR-008 | Makefile as CI command interface | Accepted |
| ADR-009 | `[skip release]` shell guard implementation | Accepted |
| ADR-010 | goreleaser caching via `go install` | Accepted |

### ADR-008: validate-and-build retains direct commands

The `validate-and-build` CI job does NOT call `make validate`. The Makefile's
`validate` target runs `cicd/check-versions.sh` as its first step — but in CI,
`check-versions.sh` would exit 1 for goreleaser (not installed in that job).
The Makefile is the developer's local interface; CI runs the same command
sequence directly. Inline comments in both reference each other.

### ADR-009: Shell guard over pipeline parameters

CircleCI's `<< pipeline.git.commit_message >>` parameter is populated only for
API-triggered pipelines — not for push events. The shell guard reads
`git log -1 --pretty=%B` at job runtime. Jobs still appear green in the CI UI
(exit 0 early), providing an audit trail that they ran but skipped intentionally.

### ADR-010: go install replaces curl | bash

goreleaser is now installed via `go install github.com/goreleaser/goreleaser/v2@<version>`
with a version-pinned cache key. A version bump in the `goreleaser-version`
pipeline parameter automatically invalidates the cache.

### Additional decisions made during delivery

- **goreleaser-version updated to v2.14.3** — updated to match the locally
  installed version discovered during step 01-04.
- **go-semver-release check made optional** in `check-versions.sh` — this tool
  is CI-only (not installed locally); treating its absence as a warning rather
  than a hard failure prevents false negatives in the pre-commit gate.
- **KANBAN_VALIDATE_DEPTH recursion guard** — `make validate` runs acceptance
  tests which include the walking skeleton test which invokes `make validate`,
  creating infinite recursion. An environment variable depth guard breaks the
  cycle cleanly.
- **go test ./internal/... in validate** — `go test ./...` inside the Makefile
  recurses into acceptance tests. The `validate` target uses `./internal/...`
  to scope unit tests only; acceptance tests are a separate `make acceptance`
  invocation.
- **Absolute KANBAN_BIN path** — relative path (`./kanban`) fails from the Go
  test working directory. The Makefile uses `$(CURDIR)/kanban` for portability.

---

## Lessons Learned

### Test recursion via Makefile

`make validate` invokes acceptance tests; the walking skeleton acceptance test
invokes `make validate`; this creates an infinite recursive subprocess loop.
The solution is an environment variable depth guard (`KANBAN_VALIDATE_DEPTH`):
the Makefile sets it before invoking acceptance tests; if it is already set,
`make validate` skips the acceptance step. Clean and dependency-free.

### go test scope in Makefile

Running `go test ./...` from inside a Makefile target recurses into
`tests/acceptance/...`, which requires the binary to exist and triggers
acceptance-specific setup. The `validate` target must use `go test ./internal/...`
to run only unit tests. The `acceptance` target handles the full acceptance run.

### Relative binary paths and working directory

Go test processes set their working directory to the package being tested, not
the project root. A relative `./kanban` path resolves to a different directory.
Use `$(CURDIR)/kanban` (absolute, set at Makefile parse time) to guarantee the
binary path is consistent regardless of which package's test process is running.

### go-semver-release is CI-only

`go-semver-release` requires access to GitHub's API and full git history with
tags — it is only meaningful in a CI context with secrets. Installing it locally
for version-check purposes only adds developer friction. `check-versions.sh`
treats its absence as a warning, not a hard failure.

---

## Testability Model

The acceptance tests implement a three-tier testability model defined in
`docs/scenarios/testable-pipeline/test-scenarios.md`:

- **Tier 1 — Fully local**: Invoke real shell commands, real Makefile targets,
  and real scripts. No secrets required. 15 scenarios.
- **Tier 2 — Requires goreleaser**: Tests for `make release-snapshot`. goreleaser
  must be installed locally. 3 scenarios.
- **Tier 3 — CI-only permanent skips**: Observable outcomes only exist inside a
  live CircleCI run (cache hit/miss, actual GitHub Release creation). 4 scenarios
  document expected CI behavior as executable specifications.

---

## Artifacts

### Permanent artifacts (migrated)

- `docs/architecture/testable-pipeline/architecture-design.md`
- `docs/architecture/testable-pipeline/component-boundaries.md`
- `docs/architecture/testable-pipeline/technology-stack.md`
- `docs/scenarios/testable-pipeline/test-scenarios.md`
- `docs/scenarios/testable-pipeline/walking-skeleton.md`
- `docs/adrs/ADR-008-makefile-as-ci-interface.md`
- `docs/adrs/ADR-009-skip-release-shell-guard.md`
- `docs/adrs/ADR-010-goreleaser-go-install-cache.md`

### Implementation artifacts

- `Makefile` — 7 targets; single source of truth for local commands
- `cicd/pre-commit` — step 0 added; go-arch-lint restored
- `cicd/config.yml` — `install-goreleaser` command; `[skip release]` guards;
  acceptance job delegates to `make acceptance`
- `tests/acceptance/pipeline_test.go` — pipeline acceptance test suite
