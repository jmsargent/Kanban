# Wave Decisions: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Date**: 2026-03-15
**Wave**: DESIGN (solution-architect)

This document records decisions made during the DESIGN wave and the handoff package for the DISTILL wave (acceptance-designer) and DEVOPS wave (platform-architect).

---

## DISCUSS Wave Artifacts Consumed

| Artifact | Status |
|---------|--------|
| `discuss/wave-decisions.md` | Read — all DISCUSS decisions carried forward |
| `discuss/requirements.md` | Read — all FRs, NFRs, business rules traced to components |
| `discuss/acceptance-criteria.md` | Read — all ACs are behavioral and implementation-neutral |
| `discuss/user-stories.md` | Read — 7 stories, all DoR-passed |
| `discuss/story-map.md` | Read — walking skeleton and release slices inform component scope |
| `discuss/outcome-kpis.md` | Read — auto-transition KPIs inform reliability requirements |
| `discover/wave-decisions.md` | Not found (discover wave not produced) |

---

## Open Decisions Resolved in DESIGN Wave

| ID | Decision | Resolution | ADR |
|----|---------|-----------|-----|
| OD-01 | Task file format | Markdown with YAML front matter | ADR-002 |
| OD-02 | Task ID generation strategy | Sequential TASK-NNN with `O_CREATE\|O_EXCL` collision prevention | data-models.md |
| OD-03 | Column configuration | Fixed defaults (todo/in-progress/done) with configurable override in `.kanban/config` | data-models.md |
| OD-04 | Implementation language | Go 1.22+, single static binary | ADR-003 |
| OD-05 | Hook delivery mechanism | `.kanban/hooks/commit-msg` (version-controlled), installed to `.git/hooks/` by `kanban install-hook` | ADR-004 |
| OD-06 | CI step distribution | `kanban ci-done` subcommand + shell wrapper Tier 1; GitHub Actions composite action R3 | ADR-005 |

---

## Architectural Decisions Made

| Decision | ADR | Summary |
|---------|-----|---------|
| Hexagonal architecture | ADR-001 | Ports-and-adapters; domain core isolated from all I/O |
| Task file format | ADR-002 | Markdown + YAML front matter; 1-line diffs for status updates |
| Language: Go | ADR-003 | Team preference; single binary; sub-ms startup |
| Git hook strategy | ADR-004 | commit-msg hook; exit-0 guarantee; delegation to Go binary |
| CI/CD integration | ADR-005 | `kanban ci-done` subcommand; `[skip ci]` annotation |

---

## Architecture Summary for Handoff

**Style**: Hexagonal (Ports and Adapters)
**Language**: Go 1.22+
**Paradigm**: OOP with interfaces as ports (Go idiom: implicit interface satisfaction)

**Three execution contexts sharing one domain core**:
1. Interactive CLI (`kanban add`, `kanban board`, etc.) — CLIAdapter (cobra)
2. Git commit-msg hook (`kanban _hook commit-msg`) — GitHookAdapter
3. CI pipeline step (`kanban ci-done`) — CIPipelineAdapter

**Package layout**:
- `internal/domain` — pure domain types and business rules
- `internal/ports` — port interfaces (TaskRepository, ConfigRepository, GitPort)
- `internal/usecases` — application logic (InitRepo, AddTask, GetBoard, Transition, EditTask, DeleteTask)
- `internal/adapters/cli` — cobra commands
- `internal/adapters/filesystem` — task and config file I/O
- `internal/adapters/git` — git process wrapper

**Architecture enforcement**: `go-arch-lint` in CI enforces that `internal/domain` has zero adapter imports.

---

## Risks Carried Forward

| ID | Risk | Status | Owner |
|----|-----|--------|-------|
| R-01 | Commit message discipline | Design mitigates via `kanban add` commit tip and `kanban install-hook` reminder | Platform-architect (CI hooks) |
| R-02 | Concurrent ID collision | Mitigated by `O_CREATE\|O_EXCL` atomic file creation | Resolved in design |
| R-03 | CI step merge conflicts | Mitigated by single-field status update (1-line diff) and `[skip ci]` | Platform-architect |
| R-04 | Hook/CI config divergence | Resolved: both use same `ConfigRepository.Read` call | Resolved in design |

---

## Handoff Package for DISTILL Wave (acceptance-designer)

### Component Boundaries Ready for AC Elaboration

| Use Case | Primary Port | Secondary Ports Used |
|---------|-------------|---------------------|
| `InitRepo` | `InitRepoUseCase` | `ConfigRepository.Write`, `GitPort.AppendToGitignore` |
| `AddTask` | `AddTaskUseCase` | `TaskRepository.NextID`, `TaskRepository.Save`, `ConfigRepository.Read` |
| `GetBoard` | `GetBoardUseCase` | `TaskRepository.ListAll`, `ConfigRepository.Read` |
| `TransitionToInProgress` | `TransitionUseCase.ToInProgress` | `TaskRepository.FindByID`, `TaskRepository.Save`, `ConfigRepository.Read` |
| `TransitionToDone` | `TransitionUseCase.ToDone` | `TaskRepository.FindByID`, `TaskRepository.Save`, `ConfigRepository.Read`, `GitPort.CommitFiles` |
| `EditTask` | `EditTaskUseCase` | `TaskRepository.FindByID`, `TaskRepository.Save` |
| `DeleteTask` | `DeleteTaskUseCase` | `TaskRepository.FindByID`, `TaskRepository.Delete` |

### Exit Code Contract (for AC writers)

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success |
| `1` | Runtime error (task not found, not a git repo, file I/O error) |
| `2` | Usage error (empty title, past due date, invalid flag) |

---

## Handoff Package for DEVOPS Wave (platform-architect)

### Development Paradigm

**OOP with interfaces as ports.** Go uses implicit interface satisfaction (no `implements` keyword). Port interfaces are defined in `internal/ports`. Adapters satisfy them by implementing the required method signatures. The domain core has zero imports outside the standard library.

### Build and Distribution

| Target | Command |
|--------|---------|
| Local build | `go build -o kanban ./cmd/kanban` |
| Cross-compile macOS arm64 | `GOOS=darwin GOARCH=arm64 go build -o kanban-darwin-arm64 ./cmd/kanban` |
| Cross-compile Linux amd64 | `GOOS=linux GOARCH=amd64 go build -o kanban-linux-amd64 ./cmd/kanban` |

**Releases**: GitHub Releases via goreleaser (MIT). One tag triggers multi-platform binary build and upload.

### CI/CD Pipeline Requirements

| Step | Stage | Notes |
|------|-------|-------|
| `go-arch-lint` | pre-build | Enforces hexagonal import rules; fast (<5s) |
| `go vet ./...` | pre-build | Static analysis |
| `golangci-lint` | pre-build | Composite linter |
| `go test ./...` | test | Unit + integration tests |
| `go build` | build | Compile binary |
| goreleaser | release | Multi-platform binary publish on tag push |

### External Integrations

This system has no third-party REST or webhook integrations. All external communication is via:
- `git` CLI process (local)
- Filesystem (`os` stdlib)

No contract testing is required. There are no external API boundaries.

### Recommended Test Strategy (for acceptance-designer)

- Domain unit tests: pure functions, no mocks
- Use case unit tests: in-memory mock adapters
- Adapter integration tests: real filesystem in `t.TempDir()`, real git repo initialised with `git init`
- End-to-end tests: invoke the compiled binary as a subprocess; assert stdout, stderr, exit codes, and file contents

---

## Deliverables Produced in DESIGN Wave

| Artifact | Path |
|---------|------|
| Architecture design + C4 diagrams | `docs/feature/kanban-tasks/design/architecture-design.md` |
| Technology stack | `docs/feature/kanban-tasks/design/technology-stack.md` |
| Component boundaries | `docs/feature/kanban-tasks/design/component-boundaries.md` |
| Data models + file schema | `docs/feature/kanban-tasks/design/data-models.md` |
| Wave decisions (this document) | `docs/feature/kanban-tasks/design/wave-decisions.md` |
| ADR-001: Hexagonal architecture | `docs/adrs/ADR-001-hexagonal-architecture.md` |
| ADR-002: Task file format | `docs/adrs/ADR-002-task-file-format.md` |
| ADR-003: Language (Go) | `docs/adrs/ADR-003-language-runtime.md` |
| ADR-004: Git hook strategy | `docs/adrs/ADR-004-git-hook-strategy.md` |
| ADR-005: CI/CD integration | `docs/adrs/ADR-005-ci-cd-integration.md` |
| Development paradigm | `CLAUDE.md` |
