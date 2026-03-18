# Requirements: Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)
**Status**: Ready for DESIGN wave handoff

---

## Business Context

The kanban-tasks CLI tool stores task state (todo/in-progress/done) in YAML front matter within task files. Three validated problems motivate this feature:

1. **No audit trail** (PRIMARY): Developers cannot see *when* tasks transitioned or *who* moved them. The data exists in git history but is not surfaced through the CLI. Opportunity score: 16.
2. **Hook writes to task files** (ARCHITECTURAL): The commit-msg hook mutating `.kanban/tasks/*.md` as a side effect is experienced as architecturally wrong. Opportunity score: 13.
3. **Repository noise** (SECONDARY): Status field rewrites appear in commit diffs alongside real code changes, creating clutter. Opportunity score: 13.

The DISCOVER wave evaluated three solution options. Option A (commit trailers as state store) was invalidated by three hard blockers: `kanban start` cannot record state without an auto-commit (violates CLAUDE.md), board performance degrades O(commits), and interactive rebase silently destroys state. Concept B (append-only transitions log) was validated as the correct architectural direction. A `kanban log` quick win was recommended as Phase 1 before the full architectural change.

---

## Scope

This feature is delivered in three independent stories across two phases:

| Story | Phase | Effort | Dependency |
|-------|-------|--------|-----------|
| US-BSG-01: `kanban log TASK-ID` | 1 (quick win) | 1 day | None |
| US-BSG-02: Append-only transitions log (Concept B) | 2 (full redesign) | 8-10 days | US-BSG-01 validated |
| US-BSG-03: `kanban board --me` | 3 (independent) | 1 day | None |

---

## Domain Language (Ubiquitous Vocabulary)

| Term | Definition |
|------|-----------|
| **task** | A unit of work tracked on the board. Has an ID (TASK-NNN), title, priority, and lifecycle status. |
| **status** | The lifecycle state of a task: `todo`, `in-progress`, `done`. |
| **transition** | A change of status. Valid: todo → in-progress, in-progress → done. |
| **transitions log** | `.kanban/transitions.log` — append-only file recording every transition with timestamp, task ID, from-status, to-status, author email, and trigger. |
| **trigger** | The mechanism that caused a transition: `manual` (kanban start), `commit:<sha>` (hook), `ci-done:<sha>` (CI). |
| **audit trail** | The complete ordered history of transitions for a task. |
| **task definition** | The non-state fields of a task: id, title, priority, due, assignee, description. |
| **git identity** | The developer's `git config user.email` — the natural key for per-user board filtering. |
| **hook** | The `commit-msg` git hook installed by `kanban init`. Fires on every `git commit`. |

---

## Non-Functional Requirements

### Performance
- `kanban log TASK-ID`: output within 500ms on a repo with 1,000+ commits (Phase 1: git log based)
- `kanban log TASK-ID`: output within 200ms on a log with 1,000 transitions (Phase 2: log file based)
- `kanban board`: unaffected latency after Phase 2 (O(tasks) + O(transitions) is faster than or equal to current O(tasks))

### Reliability
- The commit-msg hook MUST exit 0 in all cases. Wrapped in `recover()`. If the transitions log write fails, the hook prints a warning to stderr and exits 0.
- All writes to `.kanban/transitions.log` must be atomic (file locking prevents concurrent corruption).
- `kanban ci-done` must remain idempotent: running twice for the same commit range does not duplicate transitions.

### Compatibility
- `kanban board` output format is unchanged (user-visible behavior preserved).
- `kanban start` command syntax is unchanged.
- `kanban add`, `kanban edit`, `kanban delete` commands are unaffected.
- Exit codes remain: 0=success, 1=runtime error, 2=usage error.

### Architecture (CLAUDE.md non-negotiables)
- `internal/domain` has zero imports from non-stdlib packages.
- The `kanban` binary never auto-commits except `kanban ci-done`.
- All file writes are atomic (write to `.tmp`, then `os.Rename`) OR use file locking for append operations.
- The commit-msg hook always exits 0.

---

## Driving Port

For all three stories, the primary driving port is the **CLI adapter** (`internal/adapters/cli/`). The CLI commands (`kanban log`, `kanban start`, `kanban board --me`) are the hexagonal boundary through which users interact with use cases.

New secondary port required for US-BSG-02: **TransitionLogRepository** (`internal/ports/` — new interface).

---

## Business Rules

| Rule | Applies To | Description |
|------|-----------|-------------|
| BR-1 | All transitions | Valid transitions: todo → in-progress, in-progress → done only. `domain.CanTransitionTo()` is authoritative. |
| BR-2 | US-BSG-02 | A task with no log entries is implicitly `todo`. |
| BR-3 | US-BSG-02 | Latest status is determined by the LAST entry for a task in the log. |
| BR-4 | US-BSG-02 | The hook records a commit reference but does NOT advance status — status is only advanced by explicit commands (kanban start) or CI (kanban ci-done). |
| BR-5 | US-BSG-03 | `kanban board --me` filters by `assignee` field matching `git config user.email`. Warns when unassigned tasks exist. |
| BR-6 | US-BSG-01 | `kanban log` output uses domain language. Raw git commit messages are shown as supplementary context only. |
| BR-7 | US-BSG-02 | `kanban ci-done` commits ONLY `.kanban/transitions.log`. No task files are committed. |

---

## Open Questions (Red Cards)

| # | Question | Owner | Impact if Unresolved | Target Resolution |
|---|----------|-------|---------------------|------------------|
| RQ-1 | What is the exact format of each transitions.log line? Should trigger be structured or free text? | DESIGN wave (solution-architect) | Affects parsing in LatestStatus() and History() implementations | Before US-BSG-02 design begins |
| RQ-2 | Does `git log --follow` reliably track task files that have been renamed (e.g., TASK-001 → TASK-001-renamed)? | US-BSG-01 spike during implementation | If unreliable, Phase 1 output may have gaps for renamed tasks | Investigate during US-BSG-01 implementation |
| RQ-3 | Should `kanban log` support `--json` output for scripting? | Product owner | Low urgency; `--json` is a standard CLI pattern (clig.dev) | Deferred to post-MVP |
| RQ-4 | How does the migration from Phase 1 (YAML status) to Phase 2 (transitions log) work for in-progress tasks? | DESIGN wave | Blocking for US-BSG-02 on existing repos | Before US-BSG-02 implementation |

---

## Constraints Inherited from CLAUDE.md

1. The commit-msg hook always exits 0 — no exceptions.
2. All task file writes are atomic (write-to-tmp then rename).
3. `internal/domain` has zero imports from non-stdlib.
4. The `kanban` binary never auto-commits (except `kanban ci-done`).
5. Exit codes: 0=success, 1=runtime error, 2=usage error.
6. Architecture: hexagonal. New `TransitionLogRepository` is a secondary port with a filesystem adapter — no architecture rule violations expected.

---

## ADR Implications

| ADR | Status | Change Required |
|-----|--------|----------------|
| ADR-002 (task file format) | UPDATE REQUIRED | Remove `status:` field from task YAML spec; document transitions log format as the state source |
| ADR-004 (git hook strategy) | UPDATE REQUIRED | Hook behavior changes: append to log instead of mutate task file |
| ADR-005 (CI/CD integration) | UPDATE REQUIRED | `kanban ci-done` commits only transitions log, not task files |
| New ADR (transitions log) | CREATE | Document the transitions log format, append-only contract, file locking strategy |
