# CLAUDE.md — Development Paradigm

**Project**: kanban-tasks (Git-Native Kanban Task Management CLI)
**Date**: 2026-03-15

---

## Language and Paradigm

**Language**: Go 1.22+
**Paradigm**: Object-Oriented with interfaces as ports (Go idiomatic style)

Go uses implicit interface satisfaction. A type satisfies a port interface by implementing all required methods — no explicit `implements` declaration. This is the idiomatic Go expression of hexagonal architecture.

---

## Architectural Style

**Hexagonal Architecture (Ports and Adapters)**

The dependency rule is absolute: all dependencies point inward toward the domain core.

```
Primary Adapters (CLI, Hook, CI)
        |
        v
   Use Cases  <-->  Port Interfaces
        |
        v
   Domain Core
        ^
        |
Secondary Adapters (FileSystem, Git)
(implement port interfaces; injected at wiring point)
```

### Package Layout

```
cmd/kanban/          # Binary entry point — wiring only
internal/domain/     # Pure types and business rules — zero external imports
internal/ports/      # Port interfaces — imported by use cases and adapters
internal/usecases/   # Application logic — imports domain + ports only
internal/adapters/
  cli/               # cobra commands (primary adapter)
  filesystem/        # Task and config file I/O (secondary adapter)
  git/               # git process wrapper (secondary adapter)
```

### Enforced Rules

Architecture rules are enforced by `go-arch-lint` in CI. Violations fail the build.

- `internal/domain` has zero imports from `internal/ports`, `internal/usecases`, or `internal/adapters`
- `internal/usecases` has zero imports from `internal/adapters`
- No adapter package imports another adapter package
- All secondary port dependencies cross via interfaces in `internal/ports`

---

## Key Technical Decisions

| Decision | Choice | ADR |
|---------|--------|-----|
| Language | Go 1.22+ | ADR-003 |
| Architecture | Hexagonal | ADR-001 |
| Task file format | Markdown + YAML front matter | ADR-002 |
| CLI framework | `cobra` (Apache 2.0) | ADR-003 |
| Git hook type | `commit-msg`, exit-0 guarantee | ADR-004 |
| CI step | `kanban ci-done` subcommand | ADR-005 |

---

## Testing Strategy

| Layer | Approach |
|-------|---------|
| Domain | Pure unit tests — no mocks, no I/O |
| Use cases | In-memory mock adapters satisfying port interfaces |
| Adapters | Integration tests using `t.TempDir()` and real `git init` |
| End-to-end | Compiled binary invoked as subprocess; assert stdout/stderr/exit code/file state |

---

## Build Commands

```sh
# Build
go build -o kanban ./cmd/kanban

# Test
go test ./...

# Lint
golangci-lint run

# Architecture check
go-arch-lint check

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o kanban-linux-amd64 ./cmd/kanban
```

---

## Non-Negotiable Constraints

1. The commit-msg hook **always exits 0**. No exception. Wrap all hook execution in recover.
2. All task file writes are **atomic** (write to `.tmp`, then `os.Rename`).
3. `internal/domain` has **zero imports** from any non-stdlib package.
4. The `kanban` binary **never auto-commits** on behalf of the developer (except `kanban ci-done` which is explicitly a CI step).
5. Exit codes are **consistent**: 0=success, 1=runtime error, 2=usage error.

---

## Mutation Testing Strategy

Mutation testing is **disabled**. Test quality validated through code review and CI coverage.

---

## CI/CD

- **Platform**: CircleCI, config at `cicd/config.yml`
  - CircleCI must be configured to use the custom config path: Project Settings → Advanced → Config File Path → `cicd/config.yml`
- **Pre-commit hook**: `cicd/pre-commit` — install via `cicd/install-hooks.sh`
  - Runs: `go test ./...` | `golangci-lint run` | `go-arch-lint check` | `go build` | acceptance tests
- **Release**: goreleaser v2, config at `cicd/goreleaser.yml`
  - Every CI pass on `main` triggers goreleaser (continuous deployment)
  - Publishes: GitHub Releases (6 cross-compiled targets) + Homebrew tap (`homebrew-kanban` repo) + `go install`
- **Branching**: Trunk-Based Development — all work on `main`, CI on every push to `main`
