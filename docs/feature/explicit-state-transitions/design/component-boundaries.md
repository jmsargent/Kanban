# Component Boundaries — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Package Dependency Map (after feature)

```
cmd/kanban/
  └── imports: adapters/cli, adapters/filesystem, adapters/git
              (wiring point only — no business logic)

internal/adapters/cli/
  └── imports: ports, usecases, domain (for display only)
  └── NEVER imports: adapters/filesystem, adapters/git

internal/adapters/filesystem/
  └── imports: ports, domain
  └── NEVER imports: adapters/cli, adapters/git, usecases

internal/adapters/git/
  └── imports: ports
  └── NEVER imports: adapters/cli, adapters/filesystem, usecases

internal/usecases/
  └── imports: ports, domain
  └── NEVER imports: any adapter

internal/ports/
  └── imports: domain
  └── NEVER imports: adapters, usecases

internal/domain/
  └── imports: stdlib only
  └── NEVER imports: anything from this module
```

---

## File-Level Change Map

### DELETED files

| File | Why |
|------|-----|
| `internal/ports/transition_log.go` | Port removed; `CommitEntry` moved to `git.go` |
| `internal/domain/transition_entry.go` | Only consumer was `TransitionLogRepository.History()` |
| `internal/adapters/filesystem/transition_log_adapter.go` | Adapter for removed port |
| `internal/adapters/filesystem/transition_log_adapter_test.go` | Tests for deleted adapter |
| `internal/adapters/filesystem/flock_unix.go` | Used only by deleted adapter |
| `internal/adapters/filesystem/flock_windows.go` | Used only by deleted adapter |
| `internal/adapters/cli/hook.go` | Hook concept removed |
| `internal/adapters/cli/hook_test.go` | Tests for deleted handler |

### ADDED files

| File | What |
|------|------|
| `internal/usecases/complete_task.go` | `CompleteTask` use case |
| `internal/usecases/complete_task_test.go` | Unit tests with in-memory mock TaskRepository |
| `internal/adapters/cli/done.go` | `kanban done <TASK-ID>` cobra command |
| `internal/adapters/cli/done_test.go` | Acceptance-style test for the done command |

### CHANGED files

| File | Change summary |
|------|---------------|
| `internal/ports/git.go` | Remove `CommitFiles`, `InstallHook`; add `CommitEntry` type (moved from transition_log.go) |
| `internal/adapters/git/git_adapter.go` | Remove `CommitFiles()`, `InstallHook()` implementations |
| `internal/adapters/git/git_adapter_test.go` | Remove tests for deleted methods |
| `internal/adapters/cli/root.go` | Remove `log` param; remove `_hook` command; add `done` command |
| `internal/adapters/cli/ci_done.go` | Remove `log` param; update use case call (no CommitFiles) |
| `internal/adapters/cli/board.go` | Remove `log` param |
| `internal/adapters/cli/start.go` | Remove `log` param |
| `internal/adapters/cli/log.go` | Remove `log` param |
| `internal/usecases/get_board.go` | Remove `log` field; read `t.Status` directly from task YAML |
| `internal/usecases/get_board_test.go` | Remove TransitionLogRepository mock; simplify tests |
| `internal/usecases/start_task.go` | Remove `log` field; read `task.Status`; call `tasks.Update()` |
| `internal/usecases/start_task_test.go` | Remove log mock; add assertions on YAML status update |
| `internal/usecases/transition_done.go` | Remove `log` field + `CommitFiles`; call `tasks.Update()` |
| `internal/usecases/transition_done_test.go` | Remove log mock; assert YAML update, no git commit |
| `internal/usecases/transition_task.go` | Delete `TransitionToInProgress`; keep/delete `TransitionTask` |
| `internal/usecases/get_task_history.go` | Remove `log` field; use only `git.LogFile()` |
| `internal/usecases/get_task_history_test.go` | Remove log mock |
| `internal/usecases/init_repo.go` | Remove `InstallHook` + `CommitFiles` calls; remove hook.log gitignore entry |
| `internal/usecases/init_repo_test.go` | Remove assertions on hook install and init commit |
| `cmd/kanban/main.go` | Remove `log` variable; simplify `NewRootCommand` call |
| `.kanban/hooks/commit-msg` (repo file) | Delete this file from the repo |

---

## Interface Compliance After Change

`TransitionLogRepository` is removed. All types that previously implemented it (`TransitionLogAdapter`) are deleted.

Remaining port implementations:

| Port | Implementor | Unchanged? |
|------|-------------|-----------|
| `TaskRepository` | `filesystem.TaskRepository` | Yes |
| `ConfigRepository` | `filesystem.ConfigRepository` | Yes |
| `EditFilePort` | `filesystem.TaskRepository` (dual role) | Yes |
| `GitPort` | `git.GitAdapter` | Partial — 2 methods removed |

---

## Mock Impact (Tests)

All test files that define a `mockLog` or similar struct implementing `TransitionLogRepository` must remove that mock. Affected use case test files: `get_board_test.go`, `start_task_test.go`, `transition_done_test.go`, `get_task_history_test.go`, `transition_done_test.go`, `advance_by_commit_test.go`.

The `ports/ports_test.go` and `ports/compile_test.go` files may reference `TransitionLogRepository` and should be audited.

---

## go-arch-lint Compliance

After deletions, the go-arch-lint configuration (if present) must have all references to `transition_log_adapter`, `transition_log`, and `flock_unix/windows` removed. The architecture rule that `internal/domain` has zero non-stdlib imports is unaffected.
