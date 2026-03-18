# ADR-008: Makefile as CI Command Interface

**Status**: Accepted
**Date**: 2026-03-18
**Feature**: testable-pipeline
**Resolves**: OD-02 (Makefile version sourcing), Walking Skeleton design decision

---

## Context

The kanban-tasks project has a trunk-based CI/CD pipeline on CircleCI. Every push to `main` triggers a four-job workflow: validate-and-build, acceptance, tag, release. The pipeline has produced five consecutive `ci: fix` commits in recent history — all caused by tool install and version steps that cannot be tested locally before pushing.

The root cause is structural: the CI pipeline is the only test environment for CI configuration changes. There is no local equivalent for the commands the pipeline runs. When the developer changes `cicd/config.yml`, the only way to validate the change is to push to `main` — which triggers a live release.

The quality attribute priority for this feature is testability: the developer must be able to validate any change to the pipeline before it produces a live release.

Two constraints apply:

1. CircleCI jobs run on `linux/amd64`; the developer's machine is `darwin/arm64`. The commands themselves are portable; the environment is not fully replicable. This is an accepted gap.
2. The release and tag jobs require CircleCI secrets (`GITHUB_TOKEN`, `HOMEBREW_TAP_GITHUB_TOKEN`) that cannot be used locally. This is an accepted gap, addressed separately by ADR-010 (`--snapshot` mode) and ADR-009 (`[skip release]` convention).

---

## Decision

The Makefile at the project root is the single source of truth for all command sequences. CircleCI jobs call `make <target>` rather than running commands directly. If `make validate` passes locally, `make validate` called by CI produces the same result (assuming same tool versions, enforced by `cicd/check-versions.sh`).

### Targets

| Target | Role | Mirrors |
|--------|------|---------|
| `validate` | Canonical validate sequence | CI `validate-and-build` job |
| `acceptance` | E2E test run | CI `acceptance` job |
| `ci` | validate + acceptance in sequence | CI validate-and-build + acceptance combined |
| `release-snapshot` | Local goreleaser dry-run | CI `release` job (secrets-free variant) |
| `release` | Full goreleaser release | CI `release` job (CI invocation only) |
| `tag-dry` | Semver computation without push | CI `tag` job (dry-run variant) |
| `help` | Target discovery | — |

### Command sequence for `make validate` (must match CI validate-and-build exactly)

```
cicd/check-versions.sh
go mod download
go-arch-lint check
go vet ./...
golangci-lint run
go test ./internal/...
go build -o kanban ./cmd/kanban
```

### Version drift prevention

`cicd/check-versions.sh` already exists and reads all tool versions from `cicd/config.yml` pipeline parameters at runtime. The `make validate` target invokes this script as its first step. No version values are duplicated in the Makefile.

### CI job structure

CI jobs retain their job-level scaffolding (checkout, workspace, tool installation, secrets injection) but replace multi-line command sequences with `make <target>`. Tool installation commands (`install-golangci-lint`, `install-go-arch-lint`, `install-goreleaser`) are retained as pre-steps that ensure tools are in PATH before `make <target>` runs.

---

## Alternatives Considered

### Alternative 1: Shell script (run-ci.sh or similar)

A single shell script containing all job commands, invoked both locally and by CI.

Rejection rationale:
- No standard interface for discovery (`make help` vs memorizing script names and flags)
- No dependency ordering between targets (make tracks `validate` as a prerequisite for `acceptance`)
- Lower community familiarity — Makefiles are the de-facto standard for Go project task runners
- Does not support independent target invocation (`./run-ci.sh validate` is ad-hoc; `make validate` is idiomatic)

### Alternative 2: Keep CI running commands directly; provide a separate local script

Maintain the current approach (CI runs commands directly) and add a separate `run-locally.sh` that duplicates the same commands for local use.

Rejection rationale:
- Recreates the synchronisation problem this feature is designed to eliminate: two places where the command sequence lives, two places to update when it changes
- The root cause of the divergence problem is having command sequences in more than one place. Two places is the minimum that produces divergence. One place (the Makefile) is the correct solution.

### Alternative 3: Use a CircleCI local runner (circleci local execute)

Run CircleCI jobs locally using the CircleCI CLI.

Rejection rationale:
- Explicitly evaluated and rejected in DISCOVER wave: CircleCI CLI has known limitations with orbs, contexts, and workspace attachments that apply directly to this project's pipeline
- Adds a new tooling dependency (CircleCI CLI) for partial coverage
- The Makefile approach provides equivalent coverage for the validate-and-build and acceptance jobs (the highest-frequency failure class) without the limitations

---

## Consequences

**Positive**:
- Command sequences defined in exactly one place — the Makefile
- Local and CI parity guaranteed by construction: `make validate` is the same invocation in both environments
- Standard interface: `make help` is discoverable; `make <target>` is idiomatic for Go projects
- Each target documents its CI mirror and known gaps inline — no hidden differences
- `cicd/check-versions.sh` is already implemented and already reads versions dynamically; the Makefile benefits from this without any version duplication

**Negative**:
- CI job structure changes: existing multi-step CircleCI jobs become single `make <target>` invocations. This reduces the visibility of individual step names in the CircleCI UI (all steps appear as one `make validate` run). Mitigation: Makefile targets print step labels (`[1/6] go test ./...`) that are visible in CI logs.
- `make validate` builds the binary to `./kanban` in the working directory; the pre-commit hook builds to a temp path. This is an accepted minor divergence (pre-commit's temp-path pattern protects working directory cleanliness during commit).
- Developers must have `make` installed. On macOS this requires Xcode Command Line Tools; on CircleCI Go images it is pre-installed.
