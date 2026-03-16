# Branching Strategy — kanban-tasks

**Wave**: DEVOPS
**Date**: 2026-03-16
**Feature**: kanban-tasks (Git-Native Kanban Task Management CLI)

---

## Strategy: Trunk-Based Development

All production code lives on `main`. There are no long-lived branches. Feature work is done in short-lived branches (< 1 day) or directly on `main` for trivial changes.

## Rationale

| Factor | Evidence |
|--------|----------|
| Team size | Small team (indicated by single-developer commit history and tooling context) |
| Release cadence | Continuous — every CI pass on main triggers a release |
| Test suite maturity | Full coverage: unit, integration, acceptance (E2E subprocess) |
| Tool complexity | Single static binary — no multi-environment promotion needed |

Trunk-Based Development is the canonical branching model for Continuous Deployment. It eliminates merge debt, keeps the main branch always releasable, and aligns with the DORA metrics goal of high deployment frequency.

GitFlow was rejected: it is designed for scheduled releases and multi-version support, both inapplicable here. GitHub Flow (PR-required) adds friction inconsistent with a single-developer or pair-programming context and trunk-based discipline.

## Rules

### The main Branch

- Always releasable. Every commit that reaches `main` must pass the pre-commit hook.
- Every push to `main` triggers CI. If CI passes, CD runs automatically.
- Force pushes to `main` are prohibited.
- History is linear (rebase, not merge commits, for short-lived branches).

### Short-Lived Feature Branches

- Naming: `<initials>/<topic>` (e.g., `js/add-edit-command`) or `fix/<description>`.
- Lifetime: merged or abandoned within 1 working day. Branches open longer than 1 day are a signal of a decomposition problem.
- Merging: rebase onto `main`, then fast-forward merge. No merge commits.
- No CI runs on feature branches (by design). The pre-commit hook provides local validation.

### Tags and Releases

- goreleaser creates a semver tag (`v<MAJOR>.<MINOR>.<PATCH>`) on each release. The tag is created by goreleaser as part of the CD step, not by a developer.
- Tag format: `v0.1.0`, `v0.2.0`, etc.
- goreleaser determines the next version automatically from commit messages using conventional commits (or increments patch by default if no conventional commit prefix is present).

## Developer Workflow

```
# Start work
git checkout -b js/add-board-filter    # optional: branch for larger changes
# ... make changes ...
git add -p                             # selective staging
git commit                             # pre-commit hook runs all quality gates
# if hook passes, commit lands
git checkout main
git rebase js/add-board-filter         # rebase onto main
git push origin main                   # triggers CI → CD → release
```

For trivial changes (typos, doc updates):
```
git commit                             # directly on main, hook runs
git push origin main
```

## Pre-Commit Hook

The pre-commit hook (`cicd/pre-commit`) runs all CI-equivalent quality gates locally before the commit lands. This is the primary local quality gate in trunk-based development where there is no PR review gate.

Install once: `cicd/install-hooks.sh`

Hook runs: `go test ./...` → `golangci-lint run` → `go-arch-lint check` → `go build` → acceptance tests.

A failing hook blocks the commit. The developer fixes the issue before pushing. This keeps `main` always green.

## What Trunk-Based Development Does NOT Use

- No `develop` branch
- No `release/*` branches
- No `hotfix/*` branches
- No required PR reviews (small team, pre-commit hook is the gate)
- No branch-based CI (CI runs on `main` only)
- No manual deployment gates
