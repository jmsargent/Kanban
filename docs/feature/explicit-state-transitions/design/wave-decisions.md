# DESIGN Decisions — explicit-state-transitions

**Date**: 2026-03-22

---

## Key Decisions

- **[D1] Architecture: Hexagonal unchanged** — this is a pure reduction. No new ports, no new patterns. The hexagonal dependency rule is preserved. (see: architecture-design.md)

- **[D2] State source: task.Status YAML field** — `GetBoard`, `StartTask`, and `TransitionToDone` all read `domain.Task.Status` directly from the YAML-populated task struct. `TaskRepository.ListAll()` already populates this field — no change to the TaskRepository interface. (see: data-models.md)

- **[D3] New `CompleteTask` use case** — mirrors `StartTask`: receives `repoRoot` + `taskID`, reads current status from YAML, updates to done via `TaskRepository.Update()`, returns idempotent result if already done. No git. (see: architecture-design.md)

- **[D4] Remove `CommitFiles` and `InstallHook` from `GitPort`** — no callers remain. Removing dead interface methods is the correct call per YAGNI. Both are incompatible with constraint C-03. (see: component-boundaries.md)

- **[D5] Move `CommitEntry` to `git.go`** — `CommitEntry` was in `transition_log.go` but is the return type of `GitPort.LogFile()`. Moved to `git.go` where it belongs. No structural change. (see: data-models.md)

- **[D6] `GetTaskHistory` simplification** — remove `TransitionLogRepository` dependency; use only `GitPort.LogFile()`. Deduplication logic removed (single source). `kanban log` now shows git commit history of the task file. (see: architecture-design.md)

- **[D7] `InitRepo` no longer commits** — design-time discovery: `InitRepo.Execute()` also called `CommitFiles` and `InstallHook`. Both removed per C-03. Developer runs `git add .kanban/ && git commit` after `kanban init`. (see: architecture-design.md, upstream-changes.md)

- **[D8] `_hook commit-msg` safe no-op** — per AC-04-3, `kanban _hook commit-msg` must exit 0 for backwards compatibility with leftover hooks on developer machines. Implementation: the subcommand is replaced with a no-op handler that exits 0 immediately (not deleted from the cobra tree). (see: architecture-design.md)

---

## Architecture Summary

- **Pattern**: Modular monolith with hexagonal (ports-and-adapters) — unchanged
- **Paradigm**: OOP with interfaces as ports (Go idiomatic) — unchanged
- **Key components changed**: `GetBoard`, `StartTask`, `TransitionToDone`, `GetTaskHistory`, `InitRepo` (simplified); `CompleteTask` (new); `GitPort` (2 methods removed)
- **Key components deleted**: `TransitionLogRepository` port, `TransitionLogAdapter`, `domain.TransitionEntry`, hook handler, flock files

---

## Technology Stack

- Go 1.22+, cobra — unchanged
- No new dependencies

---

## Constraints Established

- **C-03 (extended)**: Binary never calls `git commit` or `git add`. Now applies to `InitRepo` as well (was previously only `TransitionToDone`).

---

## Upstream Changes (relative to DISCUSS)

### Design-time discovery: `kanban init` also commits

**DISCUSS assumed**: only `kanban ci-done` auto-commits.

**Reality discovered in code**: `InitRepo.Execute()` also calls `git.CommitFiles()` (commits initial config) and `git.InstallHook()`. Both must be removed per C-03.

**Impact on acceptance criteria**: Add AC-init-1: Given `kanban init` runs, then no git commit or git add subprocess is invoked. Noted in `upstream-changes.md`.

### `_hook commit-msg` safe no-op implementation detail

**DISCUSS assumed**: `kanban _hook commit-msg` exits 0 (AC-04-3).

**Design clarifies**: The `_hook` cobra subcommand is retained in the binary as a no-op handler (not removed from the cobra command tree) to satisfy AC-04-3 without relying on cobra's "unknown command" handling. This is an implementation choice that does not affect the user story.
