# Data Models — new-editor-mode

**Wave**: DESIGN
**Date**: 2026-03-22

---

## Summary

This feature introduces no new domain types. All data models are reused from the existing system. This section documents the data that flows through the new editor-mode path and clarifies how existing types are used.

---

## Data Flow: Zero-arg Editor Path

```
WriteTempNew()
  └─ Produces: temp YAML file (on disk, temp path)

editor.ReadTemp(tmpFile)
  └─ Returns: ports.EditSnapshot{Title, Priority, Assignee, Description}
              (Due field is present in EditSnapshot struct but will be empty string
               since WriteTempNew omits due from the template)

usecases.AddTaskInput (assembled by CLI adapter)
  └─ Title:     from EditSnapshot.Title (trimmed)
     Priority:  from EditSnapshot.Priority
     Assignee:  from EditSnapshot.Assignee
     Description: from EditSnapshot.Description
     Due:       nil (not collected via editor — WD-01)
     CreatedBy: from git.GetIdentity().Name

usecases.AddTask.Execute()
  └─ Returns: domain.Task (persisted)
              ID, Title, Status=todo, Priority, Due=nil, Assignee, Description, CreatedBy
```

---

## Blank Task Template (WriteTempNew output)

The temporary YAML file produced by `WriteTempNew()` has this shape:

```yaml
# title is required — save with a non-empty title to create the task
title: ""
# optional fields — leave blank to skip
priority: ""
assignee: ""
description: ""
```

Properties:
- `due` field is absent (WD-01: due date excluded from editor template)
- Comment lines are valid YAML comments, ignored by `yaml.Unmarshal` in `ReadTemp`
- All value fields default to empty string
- Field order matches the existing `editFields` struct minus `due`

---

## Existing Types Used (unchanged)

### `ports.EditSnapshot`

```go
type EditSnapshot struct {
    Title       string
    Priority    string
    Due         string
    Assignee    string
    Description string
}
```

When returned from `ReadTemp` after a `WriteTempNew` session: `Due` will always be an empty string (the template did not include a `due` field, so `yaml.Unmarshal` leaves it zero-valued). The CLI adapter does not use `EditSnapshot.Due` in the editor-mode path — `AddTaskInput.Due` is set to `nil`.

### `usecases.AddTaskInput`

```go
type AddTaskInput struct {
    Title       string
    Priority    string
    Due         *time.Time   // nil for editor-mode path
    Assignee    string
    Description string
    CreatedBy   string
}
```

No new fields. The editor-mode path sets `Due` to `nil` explicitly (not derived from `EditSnapshot.Due`).

### `domain.Task`

Unchanged. Created by `AddTask.Execute` with `Status = domain.StatusTodo`. The task is written to `.kanban/tasks/TASK-NNN.md` by `TaskRepository.Save` using the existing atomic write pattern.

---

## File on Disk: Resulting Task File

The task file written by `TaskRepository.Save` for an editor-mode task is identical in format to any other new task. Example:

```
---
# Status tracked in .kanban/transitions.log
id: TASK-007
title: "Build the login form"
priority: P2
due: ""
assignee: alice
created_by: Alice Smith
---

Optional description text here.
```

No special metadata distinguishes an editor-created task from an inline-title task. This satisfies WD-05 (identical output format).
