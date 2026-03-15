# Component Boundaries: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DESIGN
**Date**: 2026-03-15

---

## Package Structure

```
cmd/
  kanban/
    main.go              # Binary entry point: wires adapters to use cases, hands to cobra

internal/
  domain/                # Domain core — zero external imports
    task.go              # Task aggregate, TaskStatus enum, Column, Board, Transition types
    rules.go             # Business rules: ValidateNewTask, CanTransitionTo, IsOverdue, NextID

  ports/                 # Port interfaces — imported by use cases and adapters
    repositories.go      # TaskRepository, ConfigRepository interfaces
    git.go               # GitPort interface

  usecases/              # Application layer — imports domain + ports only
    init_repo.go
    add_task.go
    get_board.go
    transition.go        # ToInProgress and ToDone
    edit_task.go
    delete_task.go

  adapters/
    cli/                 # Primary adapter — cobra command definitions
      root.go
      init.go
      add.go
      board.go
      edit.go
      delete.go
      hook.go            # `kanban _hook commit-msg` entry point
      ci.go              # `kanban ci-done` entry point

    filesystem/          # Secondary adapter — implements TaskRepository, ConfigRepository
      task_repository.go
      config_repository.go

    git/                 # Secondary adapter — implements GitPort
      git_adapter.go
```

---

## Dependency Rules (Enforced by go-arch-lint)

| From | May Import | Must Not Import |
|------|-----------|----------------|
| `internal/domain` | stdlib only | `internal/ports`, `internal/usecases`, `internal/adapters`, any third-party library |
| `internal/ports` | `internal/domain`, stdlib | `internal/usecases`, `internal/adapters` |
| `internal/usecases` | `internal/domain`, `internal/ports`, stdlib | `internal/adapters`, any I/O library |
| `internal/adapters/cli` | `internal/usecases`, `internal/ports`, `internal/domain`, `cobra`, `fatih/color` | Other adapter packages |
| `internal/adapters/filesystem` | `internal/ports`, `internal/domain`, `gopkg.in/yaml.v3`, `adrg/frontmatter`, stdlib | `internal/usecases`, other adapter packages |
| `internal/adapters/git` | `internal/ports`, stdlib | `internal/usecases`, other adapter packages |
| `cmd/kanban` | All `internal/*`, all adapters | (entry point — only wiring allowed here) |

---

## Port Interface Contracts

### TaskRepository (secondary port)

Implemented by: `internal/adapters/filesystem.TaskRepository`

| Operation | Behaviour |
|-----------|---------|
| `FindByID(repoRoot, taskID string) (Task, error)` | Returns task or `ErrTaskNotFound` |
| `Save(repoRoot string, task Task) error` | Atomic write (tmp + rename). Creates or overwrites. |
| `ListAll(repoRoot string) ([]Task, error)` | Reads all `*.md` files in `.kanban/tasks/`. Skips unparseable files with a warning. |
| `Delete(repoRoot, taskID string) error` | Removes task file. Returns `ErrTaskNotFound` if absent. |
| `NextID(repoRoot string) (string, error)` | Atomically determines next TASK-NNN. Uses `O_CREATE\|O_EXCL` to handle concurrent adds. |

### ConfigRepository (secondary port)

Implemented by: `internal/adapters/filesystem.ConfigRepository`

| Operation | Behaviour |
|-----------|---------|
| `Read(repoRoot string) (Config, error)` | Reads `.kanban/config`. Returns `ErrNotInitialised` if absent. |
| `Write(repoRoot string, config Config) error` | Writes `.kanban/config`. Used by `kanban init`. |

### GitPort (secondary port)

Implemented by: `internal/adapters/git.GitAdapter`

| Operation | Behaviour |
|-----------|---------|
| `RepoRoot() (string, error)` | Runs `git rev-parse --show-toplevel`. Returns `ErrNotGitRepo` if not in a repo. |
| `CommitMessagesInRange(from, to string) ([]string, error)` | Runs `git log --format=%s <from>..<to>`. Used by CI step. |
| `CommitFiles(repoRoot, message string, paths []string) error` | Runs `git add <paths> && git commit -m <message>`. Used by CI step. |
| `InstallHook(repoRoot string) error` | Writes `.git/hooks/commit-msg` delegating to kanban binary. |
| `AppendToGitignore(repoRoot, entry string) error` | Appends entry to `.gitignore` if not already present. |

---

## Component Responsibilities Summary

| Component | Owns | Does Not Own |
|-----------|-----|-------------|
| Domain core | Task entity, status enum, business rules, transition validation | File I/O, git operations, CLI output formatting |
| Use cases | Orchestration of domain operations and port calls | How files are read/written, how git is invoked, how output is formatted |
| CLIAdapter | Terminal I/O, flag parsing, output formatting, exit codes | Business logic, file format details |
| GitHookAdapter | Commit message file reading, hook exit-0 guarantee, hook.log writing | Business logic, output formatting for TTY |
| CIPipelineAdapter | Non-TTY output, git commit after transition, pipeline-specific env reads | Business logic, file format details |
| FileSystemAdapter | Task file serialisation/deserialisation, atomic writes, ID collision detection | Business rules, transition logic |
| GitAdapter | git process invocation, repo root detection, commit log reading | Business rules, file format |

---

## Error Types (Domain-Defined)

| Error | Raised By | Meaning |
|-------|----------|---------|
| `ErrTaskNotFound` | `TaskRepository.FindByID`, `TaskRepository.Delete` | Task ID does not exist in `.kanban/tasks/` |
| `ErrNotInitialised` | `ConfigRepository.Read` | `.kanban/` does not exist; user must run `kanban init` |
| `ErrNotGitRepo` | `GitPort.RepoRoot` | Current directory is not inside a git repository |
| `ErrInvalidTransition` | Domain `CanTransitionTo` | Requested status transition is not directionally valid |
| `ErrDuplicateID` | `TaskRepository.NextID` | Concurrent ID generation collision (retry internally) |
| `ErrInvalidInput` | Domain `ValidateNewTask` | Missing title, past due date, invalid status value |

All errors carry a human-readable message. The CLI adapter maps domain errors to exit codes: `ErrInvalidInput` -> exit 2; all others -> exit 1.
