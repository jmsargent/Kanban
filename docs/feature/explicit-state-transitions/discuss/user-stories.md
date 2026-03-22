# User Stories — explicit-state-transitions

**Feature**: Explicit State Transitions
**Date**: 2026-03-22
**Wave**: DISCUSS

---

## Story Map Summary

| Backbone Activity | Stories |
|-------------------|---------|
| Mark a task done | US-EST-01 |
| Run CI pipeline | US-EST-02 |
| View board state | US-EST-03 |
| Maintain clean git history | US-EST-04 |
| Remove auto-hook | US-EST-05 |

---

## Stories

### US-EST-01: Explicit done transition via CLI

**As a** developer,
**I want to** run `kanban done <TASK-ID>` to mark a task as done,
**So that** I control exactly when a task is marked complete, and can commit that change as part of my own git workflow.

**Effort**: XS (0.5 day)
**Priority**: P0 — walking skeleton

**Notes**:
- Updates `status: done` in the task's YAML front matter
- Prints: `kanban: TASK-NNN moved in-progress -> done`
- Exits 0
- Does NOT stage or commit the file
- Error if task does not exist: exit 1

---

### US-EST-02: CI pipeline marks tasks done without auto-committing

**As a** CI pipeline,
**I want to** run `kanban ci-done --since=<base-ref>` to update task files to done,
**So that** my CI config explicitly owns the subsequent git commit step with full visibility.

**Effort**: S (1 day)
**Priority**: P0 — walking skeleton

**Notes**:
- Scans commit messages from `<base-ref>` to HEAD
- For each task ID found: sets `status: done` in YAML if currently `in-progress`
- Does NOT run `git add` or `git commit`
- Prints each transition: `kanban: TASK-NNN moved in-progress -> done`
- Exits 0 when no tasks to transition (silent success)
- CI config is responsible for: `git add .kanban/tasks/ && git commit -m "mark tasks done [skip ci]"`

---

### US-EST-03: Board reads state from task files

**As a** developer,
**I want** `kanban board` to display task state derived from the `status:` field in each task's YAML front matter,
**So that** the board accurately reflects current state without requiring a separate log file.

**Effort**: XS (0.5 day)
**Priority**: P0 — walking skeleton

**Notes**:
- This restores the pre-board-state-in-git behaviour
- No transitions.log is read
- State grouping: `todo` / `in-progress` / `done`

---

### US-EST-04: Remove commit-msg hook

**As a** developer,
**I want** the kanban commit-msg hook to be removed,
**So that** my git commits are not intercepted or modified by kanban, and I do not need to run `kanban install-hook` after cloning.

**Effort**: XS (0.5 day)
**Priority**: P0 — walking skeleton

**Notes**:
- `kanban _hook commit-msg` subcommand is removed
- `kanban install-hook` is removed
- `.kanban/hooks/commit-msg` hook script is removed from the repo
- Existing `.git/hooks/commit-msg` on developer machines becomes a no-op (delegates to a removed binary subcommand — will fail silently or be manually removed)
- Migration note: teams should remove `.git/hooks/commit-msg` manually; document in changelog

---

### US-EST-05: Remove transitions.log and TransitionLogRepository

**As a** developer maintaining the codebase,
**I want** the transitions.log file and its port/adapter removed from the architecture,
**So that** there is one clear source of board state and no dead code shipping in the binary.

**Effort**: S (1 day)
**Priority**: P0 — walking skeleton

**Notes**:
- Delete `internal/adapters/filesystem/transition_log_adapter.go` and related files
- Delete `internal/adapters/filesystem/flock_unix.go` and `flock_windows.go`
- Remove `TransitionLogRepository` from `internal/ports/`
- Remove `TransitionEntry` domain type if no longer used
- Update wiring in `cmd/kanban/main.go`
- Remove `go-arch-lint` references to the removed packages
- Update all tests that reference TransitionLogRepository
