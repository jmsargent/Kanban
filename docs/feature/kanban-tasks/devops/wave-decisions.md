# DEVOPS Wave Decisions — kanban-tasks

**Wave**: DEVOPS
**Date**: 2026-03-16
**Feature**: kanban-tasks (Git-Native Kanban Task Management CLI)

---

## Decision Log

| # | Decision | Choice | Rationale |
|---|---------|--------|-----------|
| D1 | Deployment Target | CLI binary — no server deployment | Tool is distributed to developers, not deployed as a service |
| D2 | Container Orchestration | None | Binary distribution requires no containers; adds complexity with zero benefit |
| D3 | CI/CD Platform | CircleCI, config at `cicd/config.yml` | Project already on CircleCI; custom config path avoids `.circleci/` convention |
| D4 | Existing Infrastructure | Greenfield — design from scratch | No existing CI/CD workflows found in repo |
| D5 | Observability | Deferred — no server to instrument | CLI tool; guardrail KPIs measured via git history and local hook.log |
| D6 | Deployment Strategy | Continuous Deployment — every CI pass on main triggers release | Trunk-based development; binary distribution; no staged rollout needed |
| D7 | Continuous Learning | Disabled | Not applicable for a CLI binary tool |
| D8 | Git Branching Strategy | Trunk-Based Development | Small team, CD cadence, mature test suite; aligns with DORA elite performance |
| D9 | Mutation Testing | Disabled | See Mutation Testing section below |

---

## Mutation Testing Decision

**Status**: Disabled.

**Rationale**: The project uses code review and a comprehensive multi-layer test suite (domain unit, use-case mock, adapter integration, E2E acceptance) as the primary quality validation mechanism. The test suite already demonstrates high mutation kill rates (93% recorded in commit `751e829`). The overhead of mutation testing in CI does not provide additional safety relative to the existing gates for a project at this scale and cadence.

**Future trigger for re-evaluation**: if the codebase exceeds 50k LOC or if a regression escapes the test suite in a way mutation testing would have caught, re-evaluate as a nightly-delta strategy.

**Persisted in**: `CLAUDE.md` under `## Mutation Testing Strategy`.

---

## CI/CD Design Decisions

### CD Trigger: Every CI Pass (Not Tagged Commits)

Continuous Deployment was chosen over tag-triggered releases. The rationale:

- Trunk-Based Development means `main` is always releasable.
- Every commit that passes CI has passed 5 quality gates and full acceptance tests.
- Manual tagging is friction that contradicts the CD philosophy.
- goreleaser creates the semver tag as part of the release process.

**Rejected alternative**: tag-triggered releases (e.g., `git tag v1.2.3 && git push --tags`).
Why rejected: adds a manual step between CI pass and user availability; inconsistent with continuous deployment and trunk-based development.

### Config Location: `cicd/` (Not `.circleci/`)

All CI/CD config lives in `cicd/` at the repo root. CircleCI supports custom config paths via Project Settings. This decision:

- Keeps CI/CD artifacts co-located (config, goreleaser config, hooks, installer).
- Makes the CI/CD surface area explicit and discoverable.
- Avoids the "magic directory" convention of `.circleci/`.

**Constraint**: requires one-time setup in CircleCI Project Settings → Advanced → Config File Path = `cicd/config.yml`.

### Release Tool: goreleaser v2

goreleaser handles: cross-compilation for 6 targets, archive packaging, GitHub Release creation with auto-changelog, Homebrew formula update. It was already referenced in the DESIGN wave decisions and is the idiomatic Go release tool.

**Rejected alternative**: custom shell scripts with `GOOS`/`GOARCH` loops.
Why rejected: goreleaser provides checksums, archive packaging, GitHub Release API integration, and Homebrew tap automation that would require significant custom script maintenance.

### Cross-Compilation Targets: 6 Platforms

| OS | Architecture | Archive Format |
|----|-------------|----------------|
| linux | amd64 | tar.gz |
| linux | arm64 | tar.gz |
| darwin | amd64 | tar.gz |
| darwin | arm64 | tar.gz |
| windows | amd64 | zip |
| windows | arm64 | zip |

Rationale: these cover the primary developer workstation configurations. Windows arm64 is included for future-proofing (Qualcomm Snapdragon developer machines).

### Distribution Channels: Three

1. **GitHub Releases**: direct binary download; universal; no toolchain required.
2. **`go install`**: zero-config for Go developers; always up-to-date.
3. **Homebrew tap**: standard macOS/Linux developer workflow; upgradeable with `brew upgrade`.

**Rejected**: Snap, Flatpak, apt/deb, rpm packages. Insufficient demand for the developer-CLI target audience.

---

## DESIGN Wave Decisions That Informed Infrastructure

These decisions were made in the DESIGN wave and constrain the DEVOPS infrastructure:

| DESIGN Decision | Infrastructure Impact |
|----------------|-----------------------|
| Hexagonal architecture; go-arch-lint enforced | `go-arch-lint check` is step 1 of CI pipeline |
| Commit-msg hook always exits 0 (ADR-004) | Hook errors go to `.kanban/hook.log`, not to git stderr |
| `kanban ci-done` as explicit CI step (ADR-005) | CI step adds < 5s guardrail KPI to enforce |
| Atomic file writes; no race conditions | No concurrency or locking infrastructure needed |
| Static binary, no external runtime deps | Binary distribution model is viable; no container runtime needed |

---

## Open Questions (None)

All decisions have been made. No deferred decisions remain for the DEVOPS wave.

---

## Constraints Summary

1. CircleCI custom config path must be configured manually in Project Settings before first pipeline run.
2. `homebrew-kanban` GitHub repo must be created before first goreleaser release.
3. Two CircleCI context secrets required: `GITHUB_TOKEN` and `HOMEBREW_TAP_GITHUB_TOKEN`.
4. goreleaser uses `[skip ci]` in its automated Homebrew formula commit to prevent CI re-trigger.
5. Pre-commit hook is opt-in (install via `cicd/install-hooks.sh`); it is not enforced by git config automatically.
