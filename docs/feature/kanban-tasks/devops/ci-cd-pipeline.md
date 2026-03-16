# CI/CD Pipeline Design — kanban-tasks

**Wave**: DEVOPS
**Date**: 2026-03-16
**Feature**: kanban-tasks (Git-Native Kanban Task Management CLI)

---

## Overview

- **CI/CD Platform**: CircleCI 2.1
- **Source Host**: GitHub
- **Config Location**: `cicd/config.yml` (custom path — see Setup section)
- **Release Tool**: goreleaser v2 (`cicd/goreleaser.yml`)
- **Branching Model**: Trunk-Based Development (all work on `main`)
- **CD Model**: Continuous Deployment — every successful CI build on `main` triggers a release

## CircleCI Custom Config Path Setup

CircleCI defaults to `.circleci/config.yml`. This project stores config in `cicd/config.yml`.

**One-time setup required**: In the CircleCI project settings:
1. Navigate to Project Settings → Advanced
2. Set "Config File Path" to `cicd/config.yml`

This must be done before the first pipeline run. Document this in the project onboarding runbook.

---

## Trigger Policy

| Event | Workflow Triggered |
|-------|--------------------|
| Push to `main` | `ci` workflow |
| `ci` workflow passes on `main` | `cd` workflow (via pipeline continuation) |
| Push to any other branch | No workflow (trunk-based: no branch CI) |

There are no scheduled triggers, no manual approval gates, and no PR-based CI (trunk-based development means all validation happens on `main`).

---

## CI Workflow — Stage Design

Target total duration: under 15 minutes.

### Stage 1: Architecture Gate (< 5 seconds)

**Step**: `go-arch-lint check`
**Runs**: First — fast gate, fails fast on architecture violations before spending time on slower checks.
**Quality gate**: Zero architecture rule violations. Violations block the pipeline.

### Stage 2: Static Analysis (< 2 minutes, parallel)

**Steps** (can run in parallel after arch gate):
- `go vet ./...` — standard Go static analysis
- `golangci-lint run` — composite linter (staticcheck, errcheck, gosec, etc.)

**Quality gate**: Zero lint errors, zero vet errors. Both must pass to proceed.

### Stage 3: Test (< 5 minutes)

**Step**: `go test ./...`
**Covers**: domain unit tests, use case tests (with in-memory mocks), adapter integration tests (using `t.TempDir()` and real `git init`).
**Quality gate**: 100% test pass rate. No coverage threshold configured (mutation testing disabled; code review is the quality gate for test quality).

### Stage 4: Build (< 1 minute)

**Step**: `go build -o kanban ./cmd/kanban`
**Produces**: `kanban` binary for the CI executor platform (linux/amd64).
**Artifact**: binary persisted to CircleCI workspace for acceptance tests.

### Stage 5: Acceptance Tests (< 5 minutes)

**Step**: `go test ./tests/acceptance/...` (or the acceptance test runner in `tests/acceptance/kanban-tasks/bin/`)
**Mode**: E2E — invokes the compiled `kanban` binary as a subprocess. Asserts stdout, stderr, exit codes, and file state.
**Dependency**: requires the binary from Stage 4 to be present on PATH.
**Quality gate**: 100% acceptance test pass rate. Blocks release.

### Stage 6: Persist Workspace

**Step**: CircleCI `persist_to_workspace` — saves the compiled binary and git metadata.
**Purpose**: CD workflow attaches this workspace to avoid rebuilding from source.

---

## CD Workflow — Release

Triggered only after `ci` workflow passes on `main`. Runs goreleaser to produce and publish the release.

### Stage 1: Attach Workspace

Attaches the workspace persisted by CI. The compiled binary is available for reference, though goreleaser will cross-compile from source for all 6 targets.

### Stage 2: goreleaser release

**Command**: `goreleaser release --config cicd/goreleaser.yml --clean`
**Reads**: `cicd/goreleaser.yml`
**Produces**:
- GitHub Release with auto-generated changelog (commits since last tag)
- 6 cross-compiled archives (tar.gz for unix, zip for windows)
- `checksums.txt` (sha256)
- Homebrew formula pushed to `homebrew-kanban` repo

**Secrets required** (via CircleCI context `release-context`):
- `GITHUB_TOKEN` — create GitHub Release, upload assets
- `HOMEBREW_TAP_GITHUB_TOKEN` — push formula to homebrew-kanban repo

**Loop prevention**: goreleaser's automated formula commit to the homebrew-kanban repo uses `[skip ci]` in the commit message to prevent a CI re-trigger loop in that repo.

---

## Artifact Promotion Strategy

The design principle is **single-source-of-truth build**: CI and CD operate on the same git commit.

1. CI validates the commit on `linux/amd64` (the CircleCI executor).
2. CD runs goreleaser against the same commit, cross-compiling for all 6 targets using the Go toolchain.
3. The Go toolchain is deterministic: same source + same toolchain version = reproducible binary.
4. No separate artifact registry (e.g., S3) is needed. The workspace carries only the CI binary (used for acceptance tests), not the release artifacts.

This avoids: artifact versioning drift, separate artifact storage costs, and the complexity of a multi-stage artifact promotion pipeline that would be over-engineered for a single-binary CLI tool.

---

## Quality Gate Summary

| Gate | Type | Blocks |
|------|------|--------|
| go-arch-lint | Blocking | Pipeline |
| go vet | Blocking | Pipeline |
| golangci-lint | Blocking | Pipeline |
| go test ./... | Blocking | Pipeline |
| go build | Blocking | Pipeline |
| Acceptance tests | Blocking | Release |
| goreleaser | Blocking | Published release |

Local quality gates (pre-commit hook) mirror CI stages 1-5. See `cicd/pre-commit`.

---

## DORA Metrics Alignment

| Metric | Target | Mechanism |
|--------|--------|-----------|
| Deployment Frequency | Every push to main that passes CI | CD triggers on every CI pass — continuous deployment |
| Lead Time for Changes | < 15 minutes (commit to release) | CI pipeline target; goreleaser adds ~2 minutes |
| Change Failure Rate | Minimized by quality gates | 6-step CI gate before any release |
| Time to Restore | Rollback = tag prior release + re-run goreleaser | goreleaser idempotent; prior release always available on GitHub Releases |

---

## Rollback Procedure

Rollback for a CLI binary distribution is re-release of a prior version.

1. **Identify prior good version**: check GitHub Releases for the last known-good tag.
2. **Re-run goreleaser**: `goreleaser release --config cicd/goreleaser.yml --clean` against the prior commit (by checking out that commit and creating a new tag).
3. **Homebrew**: goreleaser pushes the updated formula pointing to the prior version's archive URL.
4. **GitHub Releases**: prior release assets remain available for direct download at all times.

Users on `go install` will pick up the new release on next install. Homebrew users get the corrected formula on next `brew update`. Direct download users can fetch the prior release from GitHub Releases at any time — prior releases are never deleted.

No automated rollback trigger is needed: there is no live service to monitor.
