# Data Models — auto-assign-on-start

## Affected Types

### `domain.Task` — no struct change

The `Assignee` field already exists. This feature populates it automatically via `kanban start`; previously it was only set by `kanban new --assignee` or `kanban edit`.

```
Task.Assignee string
  Previously: set only via --assignee flag on kanban new, or via kanban edit
  After:      also set automatically to git config user.name on kanban start
```

### `usecases.StartTaskResult` — new field

```
PreviousAssignee string
  Zero value (empty string): task had no prior assignee, or assignee was identical to new value
  Non-empty:                 task had a different assignee; CLI adapter emits Note: warning
```

The field is populated by `StartTask.Execute` before calling `tasks.Update`:

```
previousAssignee = task.Assignee
task.Assignee = assignee          // set new value
// ...
result.PreviousAssignee = previousAssignee  // only when previousAssignee != assignee && previousAssignee != ""
```

---

## Task File — YAML Front Matter

No schema change. `assignee` is an existing front matter key. The only change is that it is now written automatically on `kanban start` rather than requiring explicit input.

```yaml
---
id: TASK-001
title: Fix login bug
status: in-progress
priority: ""
assignee: Jonathan Sargent   ← written automatically by kanban start
created_by: Jonathan Sargent
---
```

---

## No New Persistent State

This feature introduces no new files, no new front matter keys, and no schema migrations. All data flows through the existing `Task` struct and task file format (ADR-002).
