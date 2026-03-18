# Technology Stack — testable-pipeline

**Feature**: testable-pipeline
**Wave**: DESIGN
**Date**: 2026-03-18

---

## Stack Overview

This feature adds no new runtime dependencies. It introduces one new file (Makefile) and modifies two existing files (cicd/config.yml, cicd/pre-commit). All tools are already in use or are free OSS.

---

## Technology Decisions

### GNU Make (Makefile)

| Property | Value |
|----------|-------|
| Version | Any POSIX-compatible make (GNU Make 3.81+ sufficient) |
| License | GPL-3.0 |
| Purpose | Single source of truth for command sequences |
| Availability | Pre-installed on macOS (via Xcode Command Line Tools) and all CircleCI Go images |
| Alternatives considered | Shell script (run-ci.sh) — rejected in DISCOVER wave for lower discoverability; no standard interface for `make help`-equivalent |

GNU Make is the correct tool for this purpose. It provides a standard, discoverable interface (`make help`, `make <target>`), dependency ordering between targets, and is universally available in the project's two execution environments (macOS developer machine, CircleCI Go Docker image).

### goreleaser v2

| Property | Value |
|----------|-------|
| Version | v2.3.2 (pinned in cicd/config.yml — upgradeable) |
| License | MIT |
| GitHub | github.com/goreleaser/goreleaser |
| Maintenance | Active — regular releases, large community |
| Install (CI) | go install github.com/goreleaser/goreleaser/v2@<version> (replaces curl \| bash) |
| Install (local) | Same: go install github.com/goreleaser/goreleaser/v2@<version> |

The change from `curl | bash` to `go install` aligns goreleaser with the installation pattern already used for golangci-lint (via curl script), go-arch-lint, and go-semver-release (via go install). The `go install` path is more reproducible and version-pinnable.

### Existing tools (no change)

| Tool | Version (pinned in cicd/config.yml) | License | Role |
|------|--------------------------------------|---------|------|
| Go | 1.26.1 | BSD-3-Clause | Language runtime |
| golangci-lint | v2.11.3 | MIT | Composite linter |
| go-arch-lint | v1.14.0 | MIT | Architecture rule enforcement |
| go-semver-release | v6.1.0 | MIT | Semver tag computation |
| CircleCI orb circleci/go | 3.0.2 | Apache 2.0 | Go module download |

---

## No New Runtime Dependencies

This feature is developer tooling and CI configuration only. The application binary (`kanban`) is unchanged. No new Go packages, no new services, no new infrastructure.

---

## Proprietary Technology Assessment

No proprietary technology is used or introduced. CircleCI is the existing CI platform (not changed by this feature). The `[skip release]` implementation uses a shell guard (POSIX sh) — no CircleCI-specific feature is required.
