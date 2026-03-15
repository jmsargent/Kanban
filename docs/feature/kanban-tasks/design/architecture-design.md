# Architecture Design: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DESIGN
**Date**: 2026-03-15
**Architect**: Morgan (solution-architect)

---

## 1. System Context

The kanban CLI is a developer tool that lives entirely inside a git repository. There is no external database, no API server, and no synchronisation service. The system's external actors are the developer (interactive), the git process (hook invocation), and the CI platform (pipeline step invocation).

### C4 Level 1 — System Context

```mermaid
C4Context
  title System Context — kanban CLI

  Person(dev, "Developer", "Rafael / Priya: runs kanban commands in terminal, makes git commits")
  Person(ci, "CI/CD Pipeline", "GitHub Actions or GitLab CI: runs test suite and kanban ci-done step")

  System(kanban, "kanban CLI", "Git-native kanban board. Manages task files in .kanban/tasks/. Auto-transitions status on commits and CI pass.")

  System_Ext(git, "Git", "Version control system. Provides commit hooks, repo root detection, and history.")
  System_Ext(editor, "Developer Editor", "$EDITOR env var. Any text editor: vim, VS Code, nano.")
  System_Ext(ciplatform, "CI Platform", "GitHub Actions / GitLab CI. Executes test suite and kanban CI step after tests pass.")

  Rel(dev, kanban, "Runs commands via")
  Rel(kanban, git, "Detects repo root, installs hook, reads commit log via")
  Rel(kanban, editor, "Opens task file in")
  Rel(git, kanban, "Invokes commit-msg hook via")
  Rel(ci, kanban, "Invokes ci-done subcommand via")
  Rel(kanban, ciplatform, "Commits updated task files back to repo via")
```

---

## 2. Container Diagram

The kanban system is a single deployable unit: one Go binary. However, it has three distinct entry points that activate different primary adapters against the shared domain core.

### C4 Level 2 — Containers

```mermaid
C4Container
  title Container Diagram — kanban CLI

  Person(dev, "Developer")
  Person(ci, "CI Pipeline")

  System_Ext(git, "Git")
  System_Ext(editor, "$EDITOR")

  Container(cli, "kanban binary", "Go / cobra", "Single compiled binary. Entry point for all three execution contexts.")
  ContainerDb(taskfiles, ".kanban/tasks/", "Markdown + YAML front matter files", "Task state store. Committed to the git repository.")
  ContainerDb(config, ".kanban/config", "YAML file", "Column definitions and ci_task_pattern. Read by hook and CI step.")
  ContainerDb(hooklog, ".kanban/hook.log", "Plain text log", "Error log for hook failures. In .gitignore.")

  Rel(dev, cli, "Runs interactively via")
  Rel(git, cli, "Invokes as commit-msg hook via")
  Rel(ci, cli, "Invokes kanban ci-done via")
  Rel(cli, taskfiles, "Reads and writes task files in")
  Rel(cli, config, "Reads configuration from")
  Rel(cli, hooklog, "Logs hook errors to")
  Rel(cli, git, "Reads commit log, installs hook via")
  Rel(cli, editor, "Opens task file in")
```

---

## 3. Component Diagram — Hexagonal Core

The kanban binary is structured as a hexagonal (ports-and-adapters) application. This diagram shows the internal component boundaries.

### C4 Level 3 — Components (Hexagonal Core)

```mermaid
C4Component
  title Component Diagram — kanban CLI (Hexagonal Architecture)

  Person(dev, "Developer")
  Person(ci, "CI Pipeline")
  System_Ext(git, "Git process")

  Container_Boundary(binary, "kanban binary") {

    Component(cliadapter, "CLIAdapter", "cobra commands", "Primary adapter. Parses CLI flags and args, calls use cases, formats terminal output. Entry point for interactive use.")
    Component(hookadapter, "GitHookAdapter", "commit-msg entry point", "Primary adapter. Invoked by git as commit-msg hook. Parses commit message file, calls TransitionToInProgress use case. Always exits 0.")
    Component(ciadapter, "CIPipelineAdapter", "ci-done entry point", "Primary adapter. Invoked in CI after tests pass. Reads git log, calls TransitionToDone use case, commits result.")

    Component(usecases, "Use Cases", "Go interfaces + implementations", "Application layer. InitRepo, AddTask, GetBoard, TransitionToInProgress, TransitionToDone, EditTask, DeleteTask. Depends only on port interfaces.")

    Component(domain, "Domain Core", "Pure Go structs and functions", "Task entity, TaskStatus enum, Column definitions, business rules (BR-1 through BR-7). Zero external imports.")

    Component(fsadapter, "FileSystemAdapter", "os + gopkg.in/yaml.v3 + adrg/frontmatter", "Secondary adapter. Implements TaskRepository port. Reads/writes task files as Markdown with YAML front matter. Implements atomic writes.")
    Component(configadapter, "ConfigAdapter", "gopkg.in/yaml.v3", "Secondary adapter. Implements ConfigRepository port. Reads/writes .kanban/config YAML.")
    Component(gitadapter, "GitAdapter", "os/exec wrapping git CLI", "Secondary adapter. Implements GitPort. Detects repo root, reads commit log, installs hook, runs git add/commit for CI step.")
  }

  Rel(dev, cliadapter, "Runs commands via")
  Rel(git, hookadapter, "Invokes as hook via")
  Rel(ci, ciadapter, "Invokes ci-done via")

  Rel(cliadapter, usecases, "Calls use cases via")
  Rel(hookadapter, usecases, "Calls TransitionToInProgress via")
  Rel(ciadapter, usecases, "Calls TransitionToDone via")

  Rel(usecases, domain, "Creates and validates domain entities via")
  Rel(usecases, fsadapter, "Persists tasks via TaskRepository port")
  Rel(usecases, configadapter, "Reads config via ConfigRepository port")
  Rel(usecases, gitadapter, "Reads git context via GitPort")
```

---

## 4. Hexagonal Architecture Description

### Dependency Rule

All dependencies point inward. The domain core has zero imports from any adapter or external library. Use cases import only domain types and port interfaces. Adapters import use cases (for wiring) and implement port interfaces.

```
CLIAdapter ──────────────┐
GitHookAdapter ──────────┼──► UseCases ──► Domain Core
CIPipelineAdapter ───────┘        │
                                  │ (via port interfaces)
                          ┌───────┴───────┐
                          ▼               ▼
                   FileSystemAdapter  GitAdapter
                   ConfigAdapter
```

### Primary Ports (Driving — Inbound)

These are the interfaces the use cases expose to primary adapters.

| Port | Operations |
|------|-----------|
| `InitRepoUseCase` | `Init(repoRoot string) error` |
| `AddTaskUseCase` | `Add(input AddTaskInput) (Task, error)` |
| `GetBoardUseCase` | `GetBoard(repoRoot string) (Board, error)` |
| `TransitionUseCase` | `ToInProgress(repoRoot, taskID string) (Transition, error)` |
| `TransitionUseCase` | `ToDone(repoRoot string, taskIDs []string) ([]Transition, error)` |
| `EditTaskUseCase` | `Edit(repoRoot, taskID string) (TaskDiff, error)` |
| `DeleteTaskUseCase` | `Delete(repoRoot, taskID string, force bool) error` |

### Secondary Ports (Driven — Outbound)

These are the interfaces the use cases depend on; adapters implement them.

| Port | Responsibility |
|------|---------------|
| `TaskRepository` | Read task by ID, write task, list all tasks, delete task, generate next ID |
| `ConfigRepository` | Read .kanban/config, write .kanban/config |
| `GitPort` | Detect repo root, read commit messages in range, run git commit, install hook |

### Domain Core

Contains only pure Go types and functions with no external imports beyond the standard library.

| Type | Role |
|------|------|
| `Task` | Aggregate root. Fields: ID, Title, Status, Priority, Due, Assignee, Description |
| `TaskStatus` | Enum: `todo`, `in-progress`, `done` |
| `Column` | Value object: name, display label |
| `Board` | Collection of tasks grouped by status |
| `Transition` | Value object: TaskID, FromStatus, ToStatus |
| Business rules | `ValidateNewTask`, `CanTransitionTo`, `IsOverdue`, `NextID` |

---

## 5. Architecture Enforcement

Style: Hexagonal (Ports and Adapters)
Language: Go
Tool: `go-arch-lint` (MIT, `github.com/fe3dback/go-arch-lint`)

Rules to enforce (configured in `.go-arch-lint.yml`):
- Domain package (`internal/domain`) has zero imports from `internal/adapters` or `internal/infrastructure`
- Use case package (`internal/usecases`) has zero imports from `internal/adapters`
- No adapter-to-adapter dependencies (e.g., `adapters/cli` must not import `adapters/filesystem`)
- All dependencies on secondary ports cross via interfaces defined in `internal/ports`

`go-arch-lint` runs in CI as a pre-build fast check, before unit tests. A failing arch-lint check fails the CI pipeline with an actionable error message identifying the violating import.

---

## 6. Quality Attribute Strategies

### Performance (NFR-1)

- Board command reads task files sequentially from `.kanban/tasks/`. At 500 files, sequential I/O completes in ~10-20ms on a modern SSD; well within the 100ms budget.
- Hook command reads a single commit message file and one matching task file. Target: <10ms Go execution + git invocation overhead.
- No caching layer in MVP (requirements explicitly state "no cache for MVP"); the performance budget is met without it.

### Reliability (NFR-2)

- All task file writes use atomic write pattern: write to `.kanban/tasks/TASK-NNN.md.tmp`, then `os.Rename`. `Rename` is atomic on POSIX filesystems.
- Hook wraps all execution in a top-level recover; logs to `.kanban/hook.log`; always exits 0.
- ID generation uses filesystem-level atomicity: `os.OpenFile` with `O_CREATE|O_EXCL` to detect collision before writing.

### Testability

- Domain core: unit-tested with zero mocks (pure functions and value objects)
- Use cases: unit-tested with in-memory mock implementations of `TaskRepository`, `ConfigRepository`, and `GitPort`
- Adapters: integration-tested against real filesystem and real git repository in a temp directory
- Hook entry point: tested with fabricated commit message files

### Security

- The hook and CI step do not execute user-supplied strings as shell commands. Task IDs extracted by regex are used as file path components after validation (alphanumeric + hyphen only).
- The CI step commits only to `.kanban/tasks/`. No other paths are written.
- `.kanban/hook.log` is added to `.gitignore` by `kanban init` so internal error details are not committed.
