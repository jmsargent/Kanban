# Data Models — task-creator-attribution

## Domain Model Changes

### `domain.Task` (internal/domain/task.go)

Add `CreatedBy string` field:

```
Task {
  ID          string
  Title       string
  Status      TaskStatus
  Priority    string
  Due         *time.Time
  Assignee    string
  CreatedBy   string     ← NEW (plain string, empty string = no creator recorded)
  Description string
}
```

**Invariant**: `CreatedBy` is set at creation time and never mutated by any use case.
The domain type does not enforce this — enforcement is structural (edit temp file exclusion).

**Zero value**: Empty string. Backward compatible with all existing task files.

---

## Port Type Changes

### `ports.Identity` (internal/ports/git.go)

New type, defined alongside the `GitPort` interface:

```
Identity {
  Name  string   // git config user.name — the value stored in created_by
  Email string   // git config user.email — read for completeness, not stored
}
```

`Email` is read from git config but never written to task files (per DISCUSS D1: name only).
It is included in the struct because `git config` operations naturally return both,
and future callers may need the email for other purposes without requiring a second subprocess call.

### `ports.GitPort` (internal/ports/git.go)

Add method:
```
GetIdentity() (Identity, error)
```

Returns `ErrGitIdentityNotConfigured` (new sentinel error) when `user.name` is empty.
This error is handled in `cli/new.go` — the use case never sees it.

**Alternative considered**: Return `(Identity, error)` where empty `Name` is not an error —
let the caller guard. Rejected: making the port return an error for missing identity centralizes
the "what counts as valid identity" rule in one place (the adapter), consistent with how
`RepoRoot` returns `ErrNotGitRepo` rather than an empty string.

---

## Serialization Format

### YAML Front Matter Schema (task file)

Before:
```yaml
---
id: TASK-001
title: Fix login bug
status: todo
priority: ""
due: ""
assignee: ""
---
```

After:
```yaml
---
id: TASK-001
title: Fix login bug
status: todo
priority: ""
due: ""
assignee: ""
created_by: Jonathan Sargent
---
```

**Pre-existing files** (no `created_by` field): parsed without error; `CreatedBy = ""`.
**New files**: `created_by` always present and non-empty.

### `taskFrontMatter` struct (internal/adapters/filesystem/task_repository.go)

```
taskFrontMatter {
  ID        string  `yaml:"id"`
  Title     string  `yaml:"title"`
  Status    string  `yaml:"status"`
  Priority  string  `yaml:"priority"`
  Due       string  `yaml:"due"`
  Assignee  string  `yaml:"assignee"`
  CreatedBy string  `yaml:"created_by"`   ← NEW
}
```

`omitempty` is **not used** — even an empty `created_by: ""` is acceptable for new tasks.
Pre-existing files that lack the field are parsed as empty string by `gopkg.in/yaml.v3` naturally.

### `editFields` struct (internal/adapters/filesystem/task_repository.go)

**Unchanged.** `created_by` is absent from this struct by design, enforcing immutability.

---

## Use Case Input Changes

### `usecases.AddTaskInput` (internal/usecases/add_task.go)

```
AddTaskInput {
  Title       string
  Priority    string
  Due         *time.Time
  Assignee    string
  Description string
  CreatedBy   string   ← NEW
}
```

Backward compatible: existing callers that construct `AddTaskInput{}` get `CreatedBy = ""`
(empty string). Only `cli/new.go` sets it.

---

## JSON Output Changes

### `boardTaskJSON` (internal/adapters/cli/board.go)

```
boardTaskJSON {
  ID        string  `json:"id"`
  Title     string  `json:"title"`
  Status    string  `json:"status"`
  Priority  string  `json:"priority"`
  Due       *string `json:"due"`
  Assignee  string  `json:"assignee"`
  CreatedBy string  `json:"created_by"`   ← NEW
}
```

Pre-existing tasks with empty `CreatedBy`: emitted as `"created_by": ""`.
New tasks: emitted as `"created_by": "Jonathan Sargent"`.

---

## New Sentinel Error

### `ports.ErrGitIdentityNotConfigured`

```go
var ErrGitIdentityNotConfigured = errors.New("git identity not configured")
```

Defined in `internal/ports/errors.go` alongside existing `ErrTaskNotFound`, `ErrNotInitialised`, etc.

**Not propagated to the use case.** The CLI adapter handles this error and exits before
calling `uc.Execute`. The use case has no knowledge of git identity concerns.
