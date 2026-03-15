# ADR-003: Language and Runtime — Go

**Status**: Accepted
**Date**: 2026-03-15
**Feature**: kanban-tasks
**Resolves**: OD-04

---

## Context

The kanban CLI is a developer tool that must:

1. Run as an interactive terminal command (`kanban board`, `kanban add`, etc.)
2. Execute as a git commit-msg hook (fast, exits 0, logs errors)
3. Execute as a CI/CD pipeline step (non-TTY, scriptable, commits files back to repo)
4. Distribute as a single installable binary to any developer machine or CI environment

Team context: the team has a stated preference for Go as their primary development language.

Key selection criteria (in priority order):
1. Developer experience and team familiarity
2. Distribution simplicity (single static binary, no runtime dependency)
3. Performance (hook must complete within 500ms; board must produce first output within 100ms)
4. CI/CD integration ergonomics
5. YAML front matter parsing library availability

---

## Decision

**Go** (latest stable release, currently 1.22) with compilation to a single static binary via `go build`.

Primary rationale:

- **Team preference**: the team develops in Go; familiarity is the largest practical driver of velocity and code quality.
- **Single static binary**: `go build` produces a self-contained binary with zero runtime dependencies. Distribution via `brew`, `curl | sh`, or GitHub Releases requires no Node.js or Python on developer machines or CI agents.
- **Performance**: Go startup time is sub-millisecond. The 100ms first-output NFR and 500ms hook NFR have substantial headroom with Go.
- **CLI ecosystem**: `cobra` (MIT) is the industry-standard Go CLI framework, used by kubectl, Hugo, and GitHub CLI. Shell completion is built in.
- **Hexagonal architecture enforcement**: `go-arch-lint` (MIT) provides YAML-configured dependency rules enforced in CI.
- **Git operations**: standard library (`os/exec` wrapping `git`) is sufficient for the hook and CI step use cases; no git library binding needed.

---

## Alternatives Considered

### Alternative 1: TypeScript (Node.js)

TypeScript has a mature CLI ecosystem (`commander`, `chalk`) and native GitHub Actions integration. `gray-matter` is the industry standard for YAML front matter parsing.

Rejection rationale: the team has explicitly stated a preference for Go as a better development language for this project. While TypeScript has advantages for CI/CD platform integration, those advantages do not outweigh the team's Go familiarity and the single-binary distribution benefit. The GitHub Actions CI step (R3) can be implemented as a composite action wrapping the Go binary.

### Alternative 2: Python

Python is readable and has excellent YAML libraries (`python-frontmatter`). Well-supported in CI environments.

Rejection rationale: Python startup time (~100-200ms) is borderline against the 100ms first-output NFR. Distribution requires a Python runtime on every target machine unless bundled with PyInstaller, which produces large and fragile binaries. Go binary distribution is significantly cleaner.

### Alternative 3: Shell Script

Zero dependencies, runs anywhere with a POSIX shell.

Rejection rationale: the complexity of the full feature set (atomic ID generation, YAML parsing, status transition logic, error code contracts, shell completion) exceeds what shell can handle cleanly and testably. Shell is appropriate for the CI step delivery wrapper (R3) but not for the primary implementation.

---

## Technology Selections

| Component | Library | License | Version | Rationale |
|-----------|---------|---------|---------|-----------|
| CLI framework | `cobra` | Apache 2.0 | ^1.8 | Industry standard for Go CLIs; built-in shell completion |
| YAML parsing | `gopkg.in/yaml.v3` | MIT | v3 | Standard Go YAML library; used for .kanban/config and front matter block |
| Front matter parsing | `github.com/adrg/frontmatter` | MIT | ^0.2 | Lightweight YAML front matter extraction for task files |
| Terminal color | `github.com/fatih/color` | MIT | ^1.16 | Respects NO_COLOR; well-maintained |
| Architecture enforcement | `github.com/fe3dback/go-arch-lint` | MIT | ^1.0 | YAML-configured import rules for hexagonal boundary enforcement in CI |
| Testing | `testing` (stdlib) + `github.com/stretchr/testify` | MIT | ^1.9 | Standard Go test runner + assertion library |

**Runtime**: Go 1.22+ (current stable). Single static binary output via `go build -o kanban ./cmd/kanban`.

---

## Consequences

**Positive**:
- Single static binary: zero runtime dependency on developer machines or CI agents
- Sub-millisecond startup: substantial headroom against 100ms and 500ms performance NFRs
- Team familiarity: fastest path to a correct, idiomatic implementation
- `cobra` provides shell completion (bash, zsh, fish) with minimal configuration — directly satisfies NFR-5
- `go build` cross-compiles for macOS, Linux, and Windows from a single machine

**Negative**:
- YAML front matter parsing requires a dedicated library (`adrg/frontmatter`) rather than the all-in-one `gray-matter` available in the Node ecosystem — mitigated by the library being small and well-maintained
- GitHub Actions native TypeScript SDK is not applicable; R3 CI step will be a composite action wrapping the Go binary — this is a standard pattern and not a material constraint
- `go-arch-lint` is less widely adopted than `dependency-cruiser`; architecture rule configuration requires YAML rather than JavaScript — acceptable tradeoff

**Distribution strategy**: Pre-built binaries published via GitHub Releases for macOS (amd64/arm64), Linux (amd64/arm64), and Windows (amd64). Homebrew formula for macOS. `go install` for teams with Go toolchain present.
