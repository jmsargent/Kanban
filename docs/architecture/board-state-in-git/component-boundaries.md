# Component Boundaries — board-state-in-git

**Feature**: board-state-in-git
**Date**: 2026-03-18
**Status**: Approved

---

## 1. Dependency Rule

All dependencies point inward. The dependency order is:

```
Adapters (CLI, Hook, Filesystem, Git)
    -> Use Cases
        -> Domain Core
        -> Ports (interfaces)
```

No adapter imports another adapter. No use case imports an adapter. Domain has zero imports from any layer above it.

---

## 2. Package Responsibilities

### `internal/domain/`

**Owner of**: `Task`, `TaskStatus`, `Board`, `Column`, `Transition`, `TransitionEntry`, `ValidationError`

**New in this feature**: `TransitionEntry` type with `Validate()` method.

**Allowed imports**: stdlib only (`time`, `fmt`, `strings`, `errors`).

**Prohibited**: any import from `internal/ports/`, `internal/usecases/`, `internal/adapters/`.

### `internal/ports/`

**Owner of**: all port interfaces and value types that cross the boundary between use cases and adapters.

**New in this feature**:
- `TransitionLogRepository` interface (`Append`, `LatestStatus`, `History`)
- `CommitEntry` value type (added to `git.go`)
- `LogFile` method added to `GitPort` interface

**Allowed imports**: `internal/domain/` and stdlib only.

**Prohibited**: any import from `internal/usecases/` or `internal/adapters/`.

### `internal/usecases/`

**Owner of**: `GetTaskHistory`, `GetBoard`, `StartTask`, `TransitionToDone`, `AddTask`, `DeleteTask`, `EditTask`, `ListTasks`

**Modified in this feature**:
- `GetBoard` — accepts `filterAssignee`; reads status from `TransitionLogRepository`
- `StartTask` — calls `TransitionLogRepository.Append` instead of `TaskRepository.Update`
- `TransitionToDone` — calls `TransitionLogRepository.Append` and `GitPort.CommitFiles` for transitions.log only

**New in this feature**:
- `GetTaskHistory` — takes `repoRoot`, `taskID`; returns `TaskHistory`

**Allowed imports**: `internal/domain/`, `internal/ports/`, stdlib.

**Prohibited**: any import from `internal/adapters/`.

### `internal/adapters/cli/`

**Owner of**: all cobra commands, flag parsing, output formatting to stdout.

**Modified in this feature**:
- `kanban board`: passes `--me` flag and caller email to `GetBoard`
- New command: `kanban log <TASK-ID>` wired to `GetTaskHistory`

**Allowed imports**: `internal/usecases/`, `internal/ports/`, `internal/domain/`, stdlib, `cobra`.

**Prohibited**: direct imports from `internal/adapters/filesystem/` or `internal/adapters/git/`.

### `internal/adapters/filesystem/`

**Owner of**: `TaskFileAdapter` (implements `TaskRepository`) and — new — `TransitionLogAdapter` (implements `TransitionLogRepository`).

**New in this feature**:
- `TransitionLogAdapter` with `Append`, `LatestStatus`, `History` implementations
- File locking via `syscall.Flock`
- Atomic write for task files (no change to mechanism; `status:` field write removed)

**Allowed imports**: `internal/domain/`, `internal/ports/`, stdlib, `syscall`.

**Prohibited**: any import from `internal/adapters/git/` or `internal/adapters/cli/`.

### `internal/adapters/git/`

**Owner of**: `GitAdapter` implementing `GitPort`.

**Extended in this feature**: `LogFile` method invoking `git log --follow`.

**Allowed imports**: `internal/domain/`, `internal/ports/`, stdlib.

**Prohibited**: any import from `internal/adapters/filesystem/` or `internal/adapters/cli/`.

### `cmd/kanban/`

**Owner of**: binary entry point only. Wires all adapters to all use cases via constructor injection.

**New wiring in this feature**:
- `TransitionLogAdapter` injected into `GetBoard`, `StartTask`, `TransitionToDone`, hook handler
- Extended `GitAdapter` injected into `GetTaskHistory`
- `GetTaskHistory` use case wired to `kanban log` cobra command

---

## 3. New Dependency Edges (added by this feature)

| From | To | Via |
|------|----|-----|
| `GetBoard` use case | `TransitionLogRepository` port | `LatestStatus` call per task |
| `StartTask` use case | `TransitionLogRepository` port | `Append` call |
| `TransitionToDone` use case | `TransitionLogRepository` port | `Append` call |
| `GetTaskHistory` use case | `GitPort` | `LogFile` call |
| `FilterBoardByAssignee` (GetBoard + filter) | `GitPort` | `GetIdentity` call |
| `TransitionLogAdapter` | `TransitionLogRepository` port | implements |
| `GitAdapter` | `GitPort` | implements (extended) |

All new edges respect the dependency rule. No new cross-adapter dependencies introduced.

---

## 4. Interface Contracts — Acceptance Boundaries

These are the observable contracts that acceptance tests will verify. They describe behaviour, not implementation.

### `TransitionLogRepository.Append`
- When called with a valid `TransitionEntry`, a new line appears at the end of `.kanban/transitions.log`
- The written line matches the format defined in ADR-011
- Subsequent calls from concurrent processes do not corrupt existing lines
- Returns error only when the file cannot be opened or written; never returns error for duplicate entries

### `TransitionLogRepository.LatestStatus`
- When no entries exist for `taskID`, returns `(_, false, nil)`
- When one or more entries exist, returns the `To` field from the most recent entry and `true`
- Does not return an error for a missing or empty log file

### `TransitionLogRepository.History`
- Returns entries in ascending chronological order
- Returns empty slice (not error) for a task with no entries

### `GitPort.LogFile`
- Returns commits touching the file in descending chronological order (most recent first)
- Returns empty slice (not error) for a file with no commit history
- Follows renames via `--follow`

---

## 5. go-arch-lint Rule Extensions

The existing `.go-arch-lint.yml` must be extended to include the new `TransitionLogAdapter`. No structural rule changes are required; the new adapter follows the same pattern as `TaskFileAdapter`. The software-crafter must add `TransitionLogAdapter` to the filesystem adapter section of the arch-lint configuration during GREEN phase.
