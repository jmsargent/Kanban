# Evolution: board-state-in-git

**Date**: 2026-03-18
**Feature ID**: board-state-in-git
**Stories**: US-BSG-01, US-BSG-02, US-BSG-03
**Wave**: DELIVER (complete)
**Total Steps**: 18 across 3 phases

---

## Feature Summary

The `board-state-in-git` feature replaces the YAML `status:` field in task files
with an append-only transitions log at `.kanban/transitions.log`. Task state is now
derived from this log rather than from mutable YAML front matter in task files.

Three user stories were delivered:

- **US-BSG-01** — `kanban log <TASK-ID>`: displays the full transition history for a
  task, sourced from git commit metadata, in domain language (e.g., `todo->in-progress`).
- **US-BSG-02** — Append-only transitions log: state transitions produced by
  `kanban start`, the commit-msg hook, and `kanban ci-done` all write to
  `.kanban/transitions.log`. Task files are never mutated on a state change.
- **US-BSG-03** — `kanban board --me`: filters the board to the current developer's
  assigned tasks using `git config user.email` for identity resolution.

### Business Context

The North Star KPI for this feature was:

> "Automatic transitions feel invisible and the board always reflects current reality."

Prior to this feature, state transitions wrote to YAML front matter in task files,
which meant every `kanban start` or hook-fired commit mutated a task file. This
produced noisy diffs, complicated rebases, and no human-readable audit trail.

The four outcome KPIs from the DISCUSS wave:

| KPI | Outcome |
|-----|---------|
| KPI-1 | Developers view transition history with one command (`kanban log`) — delivered |
| KPI-2 | Zero task file mutations per hook-fired commit — delivered |
| KPI-3 | Task state preserved after git rebase — validated by acceptance test (02-08) |
| KPI-4 | Developers find their own tasks with `kanban board --me` — delivered |

---

## Steps Completed

### Phase 01 — Walking Skeleton: kanban log audit trail (US-BSG-01)

| Step | Name | Result |
|------|------|--------|
| 01-01 | Add `domain.TransitionEntry`, `TransitionLogRepository` port, `GitPort.LogFile`, and `GetTaskHistory` use case | COMMITTED |
| 01-02 | Handle empty history ("No transitions recorded yet.") | SKIPPED — covered by 01-01 |
| 01-03 | Exit 1 with "not found" for unknown task ID | SKIPPED — covered by 01-01 |
| 01-04 | Suggest "kanban board" in not-found error message | SKIPPED — covered by 01-01 |
| 01-05 | Exit 1 with "kanban init" hint when not initialised | SKIPPED — covered by 01-01 |
| 01-06 | Display full transition fields (timestamp, from->to, author, trigger) per entry | COMMITTED |

Steps 01-02 through 01-05 were skipped under approved APPROVED_SKIP protocol: the 01-01
implementation produced passing tests for all walking skeleton scenarios, eliminating the
need for follow-on implementation passes.

### Phase 02 — Append-only transitions.log storage (US-BSG-02)

| Step | Name | Result |
|------|------|--------|
| 02-01 | Remove `status:` field from task files; implement `TransitionLogAdapter` | COMMITTED |
| 02-02 | Board derives task status from `TransitionLogRepository` | COMMITTED |
| 02-03 | `kanban start` appends `todo->in-progress` to transitions.log, does not modify task file | COMMITTED |
| 02-04 | commit-msg hook appends transition to transitions.log, does not modify task files | COMMITTED |
| 02-05 | Hook exits 0 and writes stderr warning when transitions.log is unwritable | COMMITTED |
| 02-06 | `ci-done` appends `in-progress->done` and commits only transitions.log | COMMITTED |
| 02-07 | Deleted task excluded from board even when transitions.log has entries for it | COMMITTED |
| 02-08 | transitions.log entries survive a git rebase | COMMITTED |
| 02-09 | Concurrent appends are safe — correct line count, no truncated lines | COMMITTED |

Step 02-03 had the longest GREEN phase (11:39 to 16:30 — approximately 5 hours) due to
the concurrency bug described in the Lessons Learned section below.

### Phase 03 — kanban board --me filter (US-BSG-03)

| Step | Name | Result |
|------|------|--------|
| 03-01 | `board --me` shows only current developer's tasks via filterAssignee | COMMITTED |
| 03-02 | `board --me` warns about unassigned tasks hidden by filter | COMMITTED |
| 03-03 | `board --me` composes correctly with transitions.log-derived status | COMMITTED |

Step 03-03 was a validation-only step. The acceptance test passed immediately after
removing `t.Skip` — no production code change needed. The architecture composition held
by design.

---

## Key Decisions Made During This Wave

The following decisions were made or confirmed during DELIVER:

| Decision | Summary |
|----------|---------|
| WD-04 | No migration command required. kanban-tasks has not been publicly released; clean break is safe. ADR-012 superseded. |
| WD-06 | `GetBoard` does not fall back to YAML `status:`. Single authoritative state source from day one. |
| WD-07 | `GetTaskHistory` sources history from git commit log (not transitions.log), validated as correct by the walking skeleton. |
| 02-07 validation | Architecture holds by design: `GetBoard` iterates `TaskRepository.ListAll`, so deleted tasks are naturally excluded without additional defensive code. |
| 02-08 validation | transitions.log survives git rebase as an ordinary tracked file. No special rebase protection needed. |
| 03-03 validation | filterAssignee + TransitionLogRepository.LatestStatus compose correctly without modification. Integration held by design. |

See `docs/feature/board-state-in-git/design/wave-decisions.md` for the full WD record
(WD-01 through WD-10, DD-01 through DD-04).

---

## Lessons Learned

### 1. Concurrency bug in 02-09 surfaced a broader write-path issue (step 02-03, GREEN phase)

The GREEN phase for step 02-03 ran from 11:39 to 16:30 — approximately 5 hours, versus
the 3-hour estimate. The delay was caused by a concurrency defect in
`TransitionLogAdapter.Append` that only became visible once the concurrent-writes
acceptance test (02-09) was designed. The flock+O_APPEND combination did not prevent
duplicate log entries when two processes attempted to transition the same task
simultaneously. The fix added an idempotency check under the exclusive lock: read the
current last entry for the task before writing, and no-op if the transition is already
recorded.

**Lesson**: Concurrency correctness should be explicitly tested before the write adapter
is considered stable. The 02-09 step was planned at the outset; the concurrency bug was
caught within the wave rather than in production.

### 2. Architecture-holds-by-design validations accelerated Phase 02 and 03

Steps 02-07, 02-08, and 03-03 were designed expecting no production code changes. All
three passed on first run after removing `t.Skip` (or performing the scripted validation
action). This validates the hexagonal architecture principle: when responsibilities are
correctly assigned (task presence via repository, status via log, identity via git port),
compositions work without defensive glue code.

**Lesson**: "Architecture holds by design" validation steps are worth including in the
roadmap. Their expected outcome is SKIPPED/GREEN with no code change, which is a positive
signal — not a wasted step.

### 3. Rebase safety validated by design, not by implementation

Step 02-08 confirmed that transitions.log, as a tracked file, survives `git rebase -i`
without data loss. The concern (raised in the risk register) was that rebase might orphan
log entries. The test demonstrated this concern was only relevant to the discarded
commit-trailer approach — the append-only file approach has no such vulnerability.

**Lesson**: Validating architectural risk assumptions with acceptance tests (not just
reasoning) produces durable confidence.

### 4. Walking skeleton delivered 01-02 through 01-05 for free

The 01-01 implementation was comprehensive enough that all four subsequent walking
skeleton steps (01-02 through 01-05) passed without additional code. APPROVED_SKIP
protocol correctly handled this: steps were logged as SKIPPED with explicit justification
rather than silently omitted.

**Lesson**: Walking skeleton steps should be designed to be individually testable but the
implementation may collapse multiple steps. The step structure provides a backlog
checkpoint, not necessarily a separate code commit.

---

## Migrated Permanent Artifacts

| Original Location | Migrated To |
|-------------------|-------------|
| `docs/feature/board-state-in-git/design/architecture-design.md` | `docs/architecture/board-state-in-git/architecture-design.md` |
| `docs/feature/board-state-in-git/design/component-boundaries.md` | `docs/architecture/board-state-in-git/component-boundaries.md` |
| `docs/feature/board-state-in-git/design/technology-stack.md` | `docs/architecture/board-state-in-git/technology-stack.md` |
| `docs/feature/board-state-in-git/design/data-models.md` | `docs/architecture/board-state-in-git/data-models.md` |
| `docs/feature/board-state-in-git/distill/test-scenarios.md` | `docs/scenarios/board-state-in-git/test-scenarios.md` |
| `docs/feature/board-state-in-git/distill/walking-skeleton.md` | `docs/scenarios/board-state-in-git/walking-skeleton.md` |
| `docs/feature/board-state-in-git/discuss/journey-task-state-transitions.yaml` | `docs/ux/board-state-in-git/journey-task-state-transitions.yaml` |
| `docs/feature/board-state-in-git/discuss/journey-task-state-transitions-visual.md` | `docs/ux/board-state-in-git/journey-task-state-transitions-visual.md` |
| ADR-011 transitions log format | Already in `docs/adrs/ADR-011-transitions-log.md` |
| ADR-012 migration strategy | Already in `docs/adrs/ADR-012-migration-strategy.md` |

---

## Deferred Work (Post-MVP)

| Item | Reason Deferred |
|------|----------------|
| `kanban log --json` output | No current consumer; deferred to post-MVP (WD-08, DD-02) |
| In-memory index for transitions.log at scale (>100k entries) | Performance acceptable at expected scale; add when degradation is observed (DD-01) |
| Windows file locking support | Windows not a supported platform (DD-03) |
| Stripping legacy `status:` fields from existing task files | No existing user repos at release; non-issue (DD-04) |
| AC-01-11 performance benchmark (`kanban log` on 10k+ commit repo) | Run as local-only benchmark before releases; not in CI |
