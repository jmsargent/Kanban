# DISCOVER Wave Decisions — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCOVER
**Status**: COMPLETE
**Date**: 2026-03-18
**Author**: Scout (product-discoverer)

---

## Executive Summary

This document records the final decisions from the DISCOVER wave for the feature "board-state-in-git." The central question was: should task/board state be stored in git commits (via commit trailers) rather than in `.kanban/tasks/` YAML front matter?

**Decision**: The full commit-trailer approach (Option A) is NOT recommended. An append-only transition log file (Concept B) IS recommended, with a phased implementation beginning with a low-cost `kanban log` command.

---

## The Original Proposal — Evaluated

The user proposed: store state in git commit trailers, deriving board state by replaying `git log`. This "naturally maps to users" via git authorship.

### What the discovery process revealed

The proposal rests on four genuine, validated problems:

1. **Audit trail gap** (PRIMARY): No way to know *when* a task transitioned or *who* moved it. Current state is visible; history is not surfaced.
2. **Hook-writes-files feels wrong**: The commit-msg hook mutating `.kanban/tasks/*.md` files as a side effect is experienced as architecturally incorrect.
3. **Repository noise**: Status field rewrites mix task-state housekeeping with real code changes in commits and diffs.
4. **No natural per-user view**: Board state is global; there is no "show me my tasks" based on git authorship.

All four problems are real. The audit trail is the most important. The proposal correctly identifies git history as the right place for state transitions to live. The architectural instinct is sound.

**However**: the *specific mechanism* proposed (commit trailers as the sole state store) fails on three hard constraints discovered through technical analysis:

| Blocker | Detail |
|---------|--------|
| `kanban start` cannot record state | Explicit manual start has no commit to attach a trailer to — requires auto-commit, which violates the non-negotiable CLAUDE.md constraint |
| Board performance degrades with commits | `kanban board` must replay `git log` to derive state — O(commits), not O(tasks) — unacceptable at scale without a cache that reintroduces file writes |
| Rebase silently destroys state | `git rebase -i` is a standard developer operation; commit message editing during rebase can silently remove or corrupt trailers |

---

## Key Discovery: The Audit Trail Problem Is Already Partially Solved

Before committing to any architectural change, this finding must be acknowledged:

**The current system already creates an implicit audit trail.**

When the hook fires and updates `TASK-001.md`, that file change is committed to git. Running:

```
git log --follow -- .kanban/tasks/TASK-001.md
```

shows every commit that changed that file — including the status transition — with timestamps, authors, and commit messages. This is a full audit trail. It is not surfaced through the CLI, but the data exists.

**Implication**: The highest-priority pain point (audit trail) has a 2-hour implementation path that requires zero architectural change: add `kanban log <TASK-ID>` as a CLI command wrapping `git log -- .kanban/tasks/<id>.md`.

This does not solve the other problems (hook-writes-files, clutter), but it should be built first to validate that the audit trail need is the actual driver — not a proxy for a deeper architectural discomfort.

---

## Decisions

### Decision 1: Option A (full commit-trailer approach) is NOT recommended

**Status**: Invalidated by technical analysis
**Rationale**: Three hard blockers with no clean resolution within the existing constraints. The approach is conceptually appealing but architecturally incompatible with the `kanban start` command and the "binary never auto-commits" constraint.

**Evidence**:
- Hypothesis H5 (kanban start viability): FAIL — see solution-testing.md
- Hypothesis H4 (rebase safety): FAIL — see solution-testing.md
- Hypothesis H2 (board performance): CONDITIONAL FAIL without cache — see solution-testing.md

---

### Decision 2: Build `kanban log <TASK-ID>` first (Phase 1 of recommended change)

**Status**: RECOMMENDED — immediate
**Rationale**: Addresses the primary stated pain (audit trail) with a 2-hour implementation. Provides real-world validation that the audit trail is the actual root need before investing in larger architectural change.

**Implementation**: `kanban log TASK-001` runs `git log --follow --format="%h %ai %an: %s" -- .kanban/tasks/TASK-001.md` and formats output for the terminal.

**Validates**: Whether the participant's primary motivator (audit trail) is satisfied by surfacing existing data, or whether the underlying concern is really about the hook's file-writing behavior.

---

### Decision 3: Concept B (append-only transition log) IS recommended as the architectural direction

**Status**: RECOMMENDED — after Decision 2 is validated
**Rationale**: Solves all four validated problems, passes 5.5/6 hypotheses, respects all architectural constraints, and is compatible with the hexagonal architecture.

**What changes**:

| Component | Current | After Concept B |
|-----------|---------|----------------|
| Task file format | Contains `status:` field | No `status:` field — definition only |
| State storage | `status:` in YAML front matter | `.kanban/transitions.log` — append-only |
| Commit-msg hook | Rewrites `status:` in task file | Appends one line to transitions.log |
| `kanban start` | Writes `status:` to task file | Appends one line to transitions.log |
| `kanban ci-done` | Rewrites task files + makes commit | Appends to transitions.log + commits log |
| `kanban board` | Reads `status:` from task files | Reads transitions.log for latest status per task |
| `kanban log` (new) | Does not exist | Reads transitions.log filtered by task ID |
| Merge conflicts | Possible on status field writes | Not possible on append-only log |

**What does NOT change** (user-visible behavior):
- `kanban board` output format
- `kanban start` command syntax
- `kanban add` command syntax
- `kanban new`, `kanban edit`, `kanban delete` — unaffected
- Hook exit-0 guarantee — preserved
- Atomic write guarantee — preserved (log append uses file locking or atomic rename)

**Scope estimate**: 8-10 developer-days for full implementation + test coverage.

**ADR required**: Yes — ADR-002 must be updated to document the new task file format (no `status:` field) and the transitions log format. A new ADR for the transitions log is recommended.

---

### Decision 4: Per-user board view is a separate, lower-priority feature

**Status**: DEFERRED
**Rationale**: The per-user board view (O4, score: 10) is valid but does not require an architectural change to implement. The existing `Assignee` field is set by `kanban start` using `git config user.name`. A `--me` flag on `kanban board` filtering by `assignee == current identity` solves this problem with a single-day implementation and no architectural dependencies.

**This should not block or be coupled to the transitions log work.**

---

### Decision 5: Assumptions invalidated by discovery

| Assumption | Verdict | Evidence |
|-----------|---------|---------|
| A1 — Merge conflicts are a frequent pain | WEAK SIGNAL — acknowledged but not acute | No specific incident reported |
| A4 — Board performance not affected by log replay | INVALIDATED — O(commits) is unacceptable | Technical analysis: 10k commits = 800ms+ |
| A5 — Git notes are viable (Option B from Phase 1) | INVALIDATED — notes not pushed by default | Git documentation; dropped early |
| A6 — Commit trailers preserve the no-auto-commit constraint | INVALIDATED — `kanban start` has no commit | H5 failure |

| Assumption | Verdict | Evidence |
|-----------|---------|---------|
| A2 — Per-user scoping desired | CONFIRMED | Participant confirmed; solvable with simpler approach |
| A3 — Hook-writes-files feels wrong | CONFIRMED — unprompted, clear signal | Direct quote captured |
| A7 — Audit trail desired | CONFIRMED — primary motivator | Direct quote: "most important" |
| A10 — Hexagonal architecture compatible | CONFIRMED | New port is clean addition; no violations |

---

## Risk Register (at handoff)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Audit trail need is satisfied by `kanban log` (current data), making Concept B unnecessary | Medium | Low — good outcome | Build `kanban log` first; evaluate after 2-4 weeks |
| Transition log merge conflicts (two developers appending simultaneously) | Low | Low — append conflicts are trivially resolved (keep both lines) | File locking in log adapter; documented merge strategy |
| Developer confusion when `status:` disappears from task files | Medium | Low — display unchanged | Migration guide; in-file comment pointing to transitions log |
| `kanban ci-done` committing only the log file feels fragile | Low | Medium — CI pipeline behavior change | Acceptance test coverage for ci-done with log-only commit |
| Performance of transitions log on very large projects (10k+ transitions) | Low | Low — file I/O is fast; index if needed | Benchmark before optimizing; not a day-1 concern |

---

## Recommended Implementation Sequence

**Phase 1 (now — 1 day)**: `kanban log <TASK-ID>`
- Wraps `git log --follow -- .kanban/tasks/<id>.md`
- Zero architectural change
- Directly addresses the primary pain

**Phase 2 (after Phase 1 validation — 8-10 days)**: Concept B full implementation
- Only proceed if Phase 1 does not fully resolve the audit trail need, OR if the hook-writes-files / clutter problems remain significant after seeing Phase 1 in practice

**Phase 3 (independent — 1 day)**: `kanban board --me`
- Filter board by current git identity via assignee field
- No dependency on Phase 1 or 2

---

## What This Discovery Wave Answered

| Question | Answer |
|----------|--------|
| Should state live in git commits (commit trailers)? | No — `kanban start` blocker, rebase fragility, performance |
| Is the architectural instinct correct? | Yes — state and definition should be separated |
| What is the right mechanism? | Append-only transitions log, not commit trailers |
| Does the current system already partially solve the audit trail? | Yes — git log on task files exists but is not surfaced |
| What should be built first? | `kanban log <TASK-ID>` — 1 day, zero risk, directly validates the primary pain |
| What is the full solution if Phase 1 is insufficient? | Concept B — transitions log, ~8-10 days, all risks green |

---

## Handoff Package

All artifacts are in `docs/feature/board-state-in-git/discover/`:

| File | Contents |
|------|---------|
| `problem-validation.md` | 5 validated problems with customer quotes, G1 evaluation |
| `opportunity-tree.md` | OST with 6 opportunities scored, 3 solution concepts analyzed, hard constraint mapping |
| `solution-testing.md` | 6 hypotheses tested against Concept A and Concept B, consolidated results |
| `lean-canvas.md` | Full Lean Canvas, 4 big risks, go/no-go recommendation |
| `wave-decisions.md` | This document — final decisions, implementation sequence |
