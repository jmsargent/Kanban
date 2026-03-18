# ADR-011: Transitions Log — Append-Only State Authority

**Status**: Accepted
**Date**: 2026-03-18
**Feature**: board-state-in-git (US-BSG-02)
**Supersedes**: status field portion of ADR-002

---

## Context

US-BSG-02 replaces the mutable `status:` field in task YAML front matter with an append-only transitions log at `.kanban/transitions.log`. This architectural change requires decisions on:

1. **Line format**: what fields, in what order, with what separator
2. **Concurrency safety**: how to prevent torn writes when multiple processes write simultaneously (e.g., hook + CI running concurrently on a shared machine)
3. **git log reliability**: `git log --follow` behaviour on renamed files
4. **Migration path**: how existing repos with `status:` fields migrate to the log without data loss

These are distinct but interdependent decisions consolidated in a single ADR because they collectively define the append-only log contract.

---

## Decision

### D1: Line Format

Each line in `.kanban/transitions.log` is a single transition event with five space-separated fields:

```
<timestamp> <task-id> <from>-><to> <author_email> <trigger>
```

Example:
```
2026-03-18T14:22:01Z TASK-001 todo->in-progress alex@example.com commit:a1b2c3d
```

**Field details**:

| Field | Format | Constraint |
|-------|--------|-----------|
| `timestamp` | ISO 8601 UTC, no fractional seconds: `YYYY-MM-DDTHH:MM:SSZ` | Always UTC; zero time forbidden |
| `task-id` | `TASK-[0-9]+` | No spaces; validated on write |
| `from->to` | `<status>-><status>` using hyphenated status names | `todo`, `in-progress`, `done`, `bootstrap` (migration only) |
| `author_email` | RFC 5321 email, no spaces; `unknown` if unavailable | Never empty |
| `trigger` | `manual` \| `commit:<sha7>` \| `ci-done:<sha7>` \| `migration` | SHA is 7-character short SHA |

Lines starting with `#` are comments. Blank lines are ignored. Lines with a field count other than 5, or with unrecognised status values, are skipped with a stderr warning — they do not cause a fatal error. This tolerance prevents a single malformed line (e.g., from a manual editor) from breaking all kanban operations.

**Format rationale**: Space-separated is directly inspectable with standard Unix tools without a parser (`grep TASK-001 .kanban/transitions.log`, `tail -20 .kanban/transitions.log`). This aligns with the git-native, no-external-service philosophy. JSON lines (rejected alternative) are more verbose and require a parser. The `from->to` compound field keeps the transition relationship atomic and human-readable in a single token.

### D2: Write Safety — `O_APPEND` + Advisory Flock

Writes to `.kanban/transitions.log` use two complementary mechanisms:

1. **`O_APPEND|O_CREATE|O_WRONLY`**: On POSIX systems, writes to an `O_APPEND` file descriptor are atomic for sizes up to `PIPE_BUF` (4096 bytes on Linux, 512 bytes on macOS by POSIX minimum; each log line is ≤200 bytes). The kernel guarantees these writes do not interleave.

2. **`syscall.Flock(fd, LOCK_EX)`**: Advisory exclusive lock acquired before write, released after `fsync`. Prevents race conditions in the `LatestStatus` read-check-append pattern across kanban processes.

The `Flock` is held for the minimum duration: open, lock, write, sync, unlock, close. It is never held across I/O that could block (no network calls, no user input).

**Why not a third-party locking library**: `syscall.Flock` is available in Go's stdlib on all supported platforms (Linux, macOS). Windows is not a supported platform for this tool (POSIX hook dependency). A build-tagged no-op or compile error is acceptable on Windows.

**Concurrency scenario covered**: developer runs `kanban start TASK-001` (manual CLI) in the same second that a `commit-msg` hook fires for a different task. Both acquire `Flock` sequentially; both write complete lines; the log contains both entries in the order the lock was acquired.

### D3: `git log --follow` Reliability

`kanban log <TASK-ID>` uses `git log --follow --format="%H|%aI|%ae|%s" -- <file-path>` to retrieve commit history for the task file.

**`--follow` tracks file renames** within a single branch's linear history. This covers the common rename scenario (task file moved, e.g., by a future organisation feature).

**Known edge case**: if a task file is renamed in a branch that is subsequently rebased or force-pushed, the `--follow` chain may be broken for the rebased commits. The orphaned commits are not surfaced by `git log --follow`. This is a known limitation of git's rename detection under history rewriting.

**Mitigation**: `kanban log` is a read-only audit tool. Missing entries due to history rewrite are a cosmetic gap, not a data loss issue. The transitions.log (authoritative for current state) is unaffected by git history rewrites. The known limitation is documented in the `kanban log` help text.

**Decision**: accept the `--follow` limitation. The alternative (maintaining a second index) adds complexity that outweighs the benefit for a developer tool where history rewrites are infrequent.

---

## Alternatives Considered

### Alternative 1: YAML/JSON status field remains; append-only log is supplementary

Keep `status:` in task YAML as the mutable field; also write to a log for auditability. Both sources exist.

Evaluation:
- No migration required
- Dual-write complexity: every status change must update YAML AND append to log; inconsistency possible if one write fails
- `GetBoard` must decide which source to trust; ambiguity increases bug surface

Rejection rationale: the dual-write pattern is exactly the problem this feature is designed to eliminate. Two authoritative sources creates a consistency problem. Rejected.

### Alternative 2: SQLite database for transitions

Replace the text log with a SQLite database at `.kanban/kanban.db`.

Evaluation:
- Indexed queries; `LatestStatus` is O(1) with an index
- Atomic writes without advisory locking (SQLite WAL mode handles concurrency)
- Not inspectable with standard Unix tools
- Binary file in git: merges are impossible; conflicts are opaque
- Adds CGO or pure-Go SQLite dependency; increases binary size by ~5MB (modernc.org/sqlite)

Rejection rationale: binary files in git repositories cannot be meaningfully merged or diffed. The git-native design principle requires all kanban state to be human-readable text that git can diff and merge. SQLite is the correct choice for a server application; it is the wrong choice for a developer's local git repository. Rejected.

### Alternative 3: One log file per task (`.kanban/logs/TASK-NNN.log`)

Store transitions in per-task log files rather than a single combined log.

Evaluation:
- `LatestStatus` for a single task is faster (smaller file to scan)
- `GetBoard` requires opening N files for N tasks; more file system operations
- Atomic commit of all transitions: not possible without a separate index file
- More complex `TransitionLogRepository` implementation

Rejection rationale: the single-file approach allows `GetBoard` to read all current statuses in one file read with a single scan. For typical project sizes (10-100 tasks), this is faster than opening N files. The single log is also easier to inspect and audit as a whole. Rejected.

---

## Consequences

**Positive**:
- `.kanban/transitions.log` is the single authoritative state source; no dual-write inconsistency
- Log is inspectable with `grep`, `tail`, `awk` without any tooling
- Log commits to git as a text file; diffs are readable; merges are possible (append-only means no conflict on concurrent appends from different branches)
- Append-only contract means no line is ever deleted or modified; audit history is immutable
- Concurrent write safety provided by stdlib primitives; no external dependency

**Negative**:
- Linear scan for `LatestStatus` has O(n) complexity in log length; acceptable at MVP scale; may require optimisation at >100k entries (post-MVP)
- `bootstrap` status value is a special case in `LatestStatus` logic; not a domain status; requires explicit handling
- Windows support is excluded by `syscall.Flock`; acceptable given tool's POSIX hook dependency

**go-arch-lint impact**: No change to existing rules. The new `TransitionLogAdapter` in `internal/adapters/filesystem/` follows existing adapter patterns. Software-crafter must register it in `.go-arch-lint.yml` alongside existing filesystem adapters.
