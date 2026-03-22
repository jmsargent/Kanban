# Data Models — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Removed Types

### `domain.TransitionEntry` (deleted)

```go
// DELETED — was in internal/domain/transition_entry.go
type TransitionEntry struct {
    Timestamp time.Time
    TaskID    string
    From      TaskStatus
    To        TaskStatus
    Author    string
    Trigger   string
}
```

No remaining consumers after `TransitionLogRepository` is removed.

### `ports.TransitionLogRepository` (deleted)

```go
// DELETED — was in internal/ports/transition_log.go
type TransitionLogRepository interface {
    Append(repoRoot string, entry domain.TransitionEntry) error
    LatestStatus(repoRoot, taskID string) (domain.TaskStatus, error)
    History(repoRoot, taskID string) ([]domain.TransitionEntry, error)
}
```

---

## Moved Types

### `ports.CommitEntry` (moved from transition_log.go → git.go)

```go
// Moved to internal/ports/git.go — no structural change
type CommitEntry struct {
    SHA       string
    Timestamp time.Time
    Author    string
    Message   string
}
```

Reason: `CommitEntry` is the return type of `GitPort.LogFile()`. Its natural home is alongside `GitPort` in `git.go`, not in a file named after the now-deleted transition log.

---

## Unchanged Types

### `domain.Task`

```go
type Task struct {
    ID          string
    Title       string
    Status      TaskStatus   // ← THIS IS NOW THE AUTHORITATIVE STATE SOURCE
    Priority    string
    Due         *time.Time
    Assignee    string
    Description string
    CreatedBy   string
}
```

`Status` was always on `domain.Task` and populated by `TaskRepository.FindByID()` from the YAML `status:` field. The board-state-in-git feature bypassed it by reading from transitions.log instead. This feature restores `task.Status` as the authoritative source — no schema change required.

### `domain.TaskStatus`, `domain.Board`, `domain.Column`

Unchanged. `StatusTodo`, `StatusInProgress`, `StatusDone` remain the same constants.

---

## New Types

### `usecases.CompleteTaskResult`

```go
type CompleteTaskResult struct {
    AlreadyDone bool
    From        domain.TaskStatus
    Task        domain.Task
}
```

Mirrors the existing `StartTaskResult` pattern. `AlreadyDone` signals idempotent case (task was already done). `From` is the status before the transition (todo or in-progress).

---

## File State Model (unchanged schema)

Task files at `.kanban/tasks/<TASK-ID>.md`:

```yaml
---
id: TASK-001
title: "Fix login bug"
status: in-progress        ← authoritative board state
priority: high
due: ""
assignee: alice@example.com
created_by: bob@example.com
---

Description text here.
```

The `status:` field was already in the schema. The board-state-in-git feature planned to make it secondary (derived from transitions.log). This feature confirms it as primary and sole state source. No file format changes required.

---

## Transitions.log File (removed)

The file `.kanban/transitions.log` is no longer written or read. Any existing instance in a developer's repo can be deleted. No migration is required (project not yet publicly released).
