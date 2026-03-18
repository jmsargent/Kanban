# Data Models — board-state-in-git

**Feature**: board-state-in-git
**Date**: 2026-03-18
**Status**: Approved

---

## 1. transitions.log — Line Format

**File path**: `.kanban/transitions.log`

### 1.1 Format Specification

Each line is a single transition event. Fields are space-separated. There is no quoting. No field may contain a space (enforced by value constraints below).

```
<timestamp> <task-id> <from>-><to> <author_email> <trigger>
```

**Field constraints**:

| Field | Format | Example | Constraint |
|-------|--------|---------|------------|
| `timestamp` | ISO 8601 UTC, no spaces: `2026-03-18T14:22:01Z` | `2026-03-18T14:22:01Z` | Must parse as `time.Time`; always UTC; no fractional seconds |
| `task-id` | `TASK-NNN` | `TASK-042` | Matches pattern `TASK-[0-9]+` |
| `from->to` | `<status>-><status>` | `todo->in-progress` | Hyphenated statuses, no spaces; `->` literal separator |
| `author_email` | RFC 5321 email, no spaces | `alex@example.com` | No spaces allowed; anonymous fallback: `unknown` |
| `trigger` | One of three forms | `manual`, `commit:a1b2c3d`, `ci-done:a1b2c3d` | SHA is 7-character short SHA; no spaces |

### 1.2 Example Lines

```
2026-03-18T14:22:01Z TASK-001 todo->in-progress alex@example.com commit:a1b2c3d
2026-03-18T16:45:00Z TASK-001 in-progress->done alex@example.com ci-done:b2c3d4e
2026-03-19T09:10:33Z TASK-002 todo->in-progress morgan@example.com manual
2026-03-19T11:00:00Z TASK-003 todo->in-progress unknown commit:c3d4e5f
2026-03-20T00:00:00Z TASK-001 done->todo alex@example.com migration
```

### 1.3 Parsing Rules

- Lines beginning with `#` are comments; ignored by all readers.
- Blank lines are ignored.
- Lines that do not parse cleanly (wrong field count, unrecognised status values) are skipped with a warning written to stderr. They do not cause a fatal error.
- Field order is positional. No key-value pairs.
- A line MUST have exactly 5 space-separated tokens. A line with 4 or 6+ tokens is malformed.

### 1.4 Special Trigger Values

| Trigger | Meaning | When Used |
|---------|---------|-----------|
| `manual` | Explicit `kanban start` command | `StartTask` use case, interactive CLI |
| `commit:<sha7>` | Automatic on `git commit` referencing the task ID | Hook handler |
| `ci-done:<sha7>` | Automatic on CI pipeline pass | `TransitionToDone` use case |
| `migration` | Synthesised during `kanban migrate` | Migration command only; `from` is always the YAML status value; `to` is the same value (preserved status, not a transition) — see note |

**Migration note**: `kanban migrate` records the current state, not a transition. The `from` field is `bootstrap` (a reserved non-status value) and `to` is the existing YAML status. This allows `LatestStatus` to find the initial state while making it clear the entry is synthetic.

Revised migration entry format:
```
2026-03-18T00:00:00Z TASK-001 bootstrap->todo alex@example.com migration
```

`bootstrap` is not a valid `TaskStatus` for any use case other than reading by `LatestStatus`, which returns the `to` value. All other use cases that validate status values treat `bootstrap` as invalid input, preventing any direct transition from or to it.

---

## 2. Task File Format — Changes from ADR-002

### 2.1 New Task File Structure

The `status:` field is removed. A Markdown comment in the body references the state authority.

```markdown
---
id: TASK-042
title: Fix OAuth login bug
priority: P2
due: 2026-03-20
assignee: alex@example.com
created_by: alex@example.com
---

<!-- State managed by .kanban/transitions.log -->

Optional freeform description in Markdown body.
```

### 2.2 Changed Fields

| Field | Before | After |
|-------|--------|-------|
| `status` | `status: todo` in YAML front matter | Removed. Authoritative source is transitions.log |
| `assignee` | Any string | Normalised to email address (matches author_email in transitions.log for `--me` filtering) |

### 2.3 Legacy Task Files

Task files created before the `kanban migrate` command is run retain their `status:` field. `GetBoard` reads status from transitions.log first; if `LatestStatus` returns `bool=false` (no log entry for this task), it falls back to reading `status:` from the YAML front matter. This fallback is removed in a future release after migration adoption is confirmed.

---

## 3. Domain Type: TransitionEntry

`internal/domain/transition_entry.go`

Fields:
- `Timestamp time.Time` — UTC; parsed from ISO 8601 line token
- `TaskID string` — e.g., `TASK-001`
- `From TaskStatus` — the status being left; `bootstrap` treated as opaque string by LatestStatus reader
- `To TaskStatus` — the status being entered
- `Author string` — email address or `unknown`
- `Trigger string` — one of `manual`, `commit:<sha7>`, `ci-done:<sha7>`, `migration`

Validation method: `Validate() error`
- `Timestamp` must not be zero
- `TaskID` must match `TASK-[0-9]+`
- `From != To` (except migration entries where From is `bootstrap`)
- `To` must be a valid `TaskStatus`
- `Author` must not be empty
- `Trigger` must match one of the four allowed forms

---

## 4. Port Value Type: CommitEntry

`internal/ports/git.go`

Fields:
- `Hash string` — full 40-character SHA
- `Timestamp time.Time` — author timestamp, UTC
- `AuthorEmail string` — from `%ae` git format specifier
- `Message string` — commit subject line (first line only)

`CommitEntry` is an infrastructure value type produced by the git adapter. It is not a domain type. `GetTaskHistory` maps it to `HistoryEntry` using message parsing.

---

## 5. Use Case Value Type: TaskHistory

`internal/usecases/get_task_history.go`

Fields:
- `TaskID string`
- `TaskTitle string`
- `Entries []HistoryEntry`

`HistoryEntry`:
- `Timestamp time.Time`
- `From TaskStatus`
- `To TaskStatus`
- `AuthorEmail string`
- `Trigger string`

`GetTaskHistory` parses commit messages to extract `from->to` transition information. Commits whose messages do not contain a recognisable transition pattern contribute a `HistoryEntry` with `From=""` and `To=""` — these are displayed as "state change recorded" without explicit from/to values, preserving visibility of all commits touching the file.

---

## 6. File Locking Strategy

`.kanban/transitions.log` uses two complementary safety mechanisms:

| Mechanism | Scope | Purpose |
|-----------|-------|---------|
| `O_APPEND` flag | Kernel-level atomic write for POSIX writes ≤ PIPE_BUF | Prevents interleaved writes from concurrent processes |
| `syscall.Flock(LOCK_EX)` | Advisory; same-machine processes only | Ensures read-modify context is stable; prevents race between LatestStatus read and Append write |

`Flock` is advisory: processes that do not use `Flock` are not excluded. All kanban processes use `Flock`, so concurrent kanban invocations are safe. External editors opening transitions.log directly are out of scope.

The lock is acquired, the write is performed, the lock is released. The lock is never held across user interaction or I/O beyond the single append.
