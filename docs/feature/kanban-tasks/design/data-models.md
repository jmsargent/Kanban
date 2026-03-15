# Data Models: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DESIGN
**Date**: 2026-03-15
**Resolves**: OD-01 (task file format), OD-02 (task ID generation), OD-03 (column configuration)

---

## Task File Schema (OD-01 Resolved)

Format: Markdown with YAML front matter. See ADR-002 for full rationale.

### File Location

```
<repo_root>/.kanban/tasks/TASK-NNN.md
```

### Schema

```
---
id: TASK-001
title: Fix OAuth login bug
status: todo
priority: P2
due: 2026-03-20
assignee: Rafael Rodrigues
---

Optional freeform description in standard Markdown.
Multiple paragraphs, lists, code blocks — any Markdown syntax.
```

### Field Definitions

| Field | Type | Required | Values / Constraints |
|-------|------|---------|----------------------|
| `id` | string | yes | Format: `TASK-[0-9]+`. Matches filename stem. Immutable after creation. |
| `title` | string | yes | Non-empty string. No length limit enforced in MVP. |
| `status` | string | yes | One of the configured column values. Default columns: `todo`, `in-progress`, `done`. |
| `priority` | string | no | `P1`, `P2`, `P3`, or absent. Absent displayed as `--` on board. |
| `due` | string | no | ISO 8601 date: `YYYY-MM-DD`. Must be today or future on creation. Absent displayed as `--`. |
| `assignee` | string | no | Free-text string. Absent displayed as `unassigned`. |
| Body | Markdown | no | Everything after the closing `---` delimiter. Used for description. |

### Canonical File Example

```markdown
---
id: TASK-003
title: Implement API rate limiting
status: in-progress
priority: P1
due: 2026-03-25
assignee: Alex Kim
---

## Context

Rate limiting is required for the public API endpoints. Token bucket algorithm preferred.
Throttle at 100 req/min per API key.

## Acceptance

- [ ] POST /api/v1/* endpoints enforce rate limit
- [ ] 429 response includes Retry-After header
```

### Atomic Write Protocol

All writes use the pattern:
1. Write to `.kanban/tasks/TASK-NNN.md.tmp`
2. `os.Rename(".tmp", target)` — atomic on POSIX filesystems

This prevents partial writes on interrupt (NFR-2).

---

## Task ID Generation (OD-02 Resolved)

**Decision**: Sequential `TASK-NNN` format with filesystem-level collision prevention.

Rationale: Sequential IDs are human-friendly (referenced in commit messages, discussed in standups). Content-addressed hashes (SHA-based) are collision-proof but produce IDs like `TASK-a3f9c1` that are harder to reference verbally. The collision risk (R-02 from DISCUSS wave) is addressed by the `O_CREATE|O_EXCL` open flag rather than by abandoning sequential IDs.

**ID generation algorithm** (owned by `TaskRepository.NextID`):
1. List all existing task files in `.kanban/tasks/`
2. Extract the highest numeric suffix
3. Attempt `os.OpenFile("TASK-(N+1).md", O_CREATE|O_EXCL, 0644)` — fails if concurrent add already created this file
4. On collision: increment and retry (up to 3 attempts, then return `ErrDuplicateID`)

The file lock approach is POSIX-safe for the expected concurrency level (small teams, rare concurrent adds). The `O_EXCL` flag prevents two simultaneous `kanban add` calls from writing to the same file.

---

## Column Configuration (OD-03 Resolved)

**Decision**: Fixed defaults (`todo`, `in-progress`, `done`) with configurable override per repo in `.kanban/config`.

Rationale: Fixed columns cover 95% of use cases and require no configuration. Configurable columns support teams with custom workflows (e.g., adding a `review` column). The validation cost is acceptable because it is localised to the `ConfigRepository` and domain `ValidateNewTask` function.

---

## .kanban/config Schema

Format: YAML.

```yaml
version: 1
columns:
  - todo
  - in-progress
  - done
ci_task_pattern: "TASK-[0-9]+"
```

### Config Field Definitions

| Field | Type | Required | Default | Notes |
|-------|------|---------|---------|-------|
| `version` | integer | yes | `1` | Schema version for future migrations |
| `columns` | string list | yes | `[todo, in-progress, done]` | Ordered list. First = initial status on `kanban add`. Last = terminal status. |
| `ci_task_pattern` | string | yes | `"TASK-[0-9]+"` | Go regex. Used by hook and CI step to extract task IDs from commit messages. |

**Validation**: `columns` must have at least 2 entries. `ci_task_pattern` must be a valid Go regex (validated on `kanban init` and on config read). Both hook and CI step use the same `ConfigRepository.Read` call — single source of truth, no duplication.

---

## Domain Entity: Task (Go)

```
Task
  ID          string        // "TASK-001"
  Title       string        // non-empty
  Status      TaskStatus    // enum: todo | in-progress | done (or configured values)
  Priority    string        // "P1" | "P2" | "P3" | ""
  Due         *time.Time    // nil if absent; date-only (time component ignored)
  Assignee    string        // free-text | ""
  Description string        // Markdown body; may be empty
```

## Domain Type: TaskStatus

```
TaskStatus = string  // constrained to configured column values
```

Constants for default configuration:
```
StatusTodo       = "todo"
StatusInProgress = "in-progress"
StatusDone       = "done"
```

## Domain Type: Board

```
Board
  Columns  []ColumnGroup

ColumnGroup
  Status  TaskStatus
  Label   string    // display label: "TODO" | "IN PROGRESS" | "DONE"
  Tasks   []Task
```

## Domain Type: Transition

```
Transition
  TaskID      string
  FromStatus  TaskStatus
  ToStatus    TaskStatus
```

## Domain Type: Config

```
Config
  Version         int
  Columns         []string
  CITaskPattern   string
```

---

## JSON Output Schema (kanban board --json)

Versioned contract. Breaking changes require a major version bump (NFR-4).

```json
[
  {
    "id": "TASK-001",
    "title": "Fix OAuth login bug",
    "status": "in-progress",
    "priority": "P2",
    "due": "2026-03-20",
    "assignee": "Rafael Rodrigues",
    "overdue": false
  }
]
```

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | Always present |
| `title` | string | Always present |
| `status` | string | Always present |
| `priority` | string \| null | `null` if not set |
| `due` | string \| null | ISO 8601 date or `null` |
| `assignee` | string \| null | `null` if not set |
| `overdue` | boolean | `true` if `due` is before today and status is not `done` |

Schema version is surfaced in the CLI via `kanban --version`. The `--json` flag output is stable within a major version.
