# DISCUSS Decisions — explicit-state-transitions

**Date**: 2026-03-22

---

## Key Decisions

- **[D1] Feature type: Infrastructure/Backend** — removes auto-commit behavior from CI step and hook, adds explicit CLI command. No UI changes. (see: requirements.md)

- **[D2] Walking skeleton: No** — this is a brownfield correction/reduction, not a new greenfield feature. All 5 stories ship together as a single atomic change. (see: prioritization.md)

- **[D3] UX depth: Lightweight** — the problem is clearly stated by user feedback. Happy path and migration story are sufficient. (see: journey-explicit-transitions-visual.md)

- **[D4] JTBD analysis: Skipped** — motivation is explicit in the feedback: "negative feedback from having kanban perform commits." No ambiguity about the job to be done.

- **[D5] State source: Task YAML front matter** — user confirmed: "the folder .kanban contains the state, and that is going to change over commits." transitions.log is removed. (see: requirements.md FR-01)

- **[D6] Hook: Removed** — commit-msg hook and `kanban install-hook` are removed. Leftover hooks on developer machines become a safe no-op via `kanban _hook commit-msg` exiting 0 immediately. (see: requirements.md FR-02, AC-04-3)

- **[D7] `kanban ci-done`: File-update only** — no `--commit` flag, no git subprocesses. CI config is responsible for committing. (see: requirements.md FR-04)

- **[D8] New command: `kanban done`** — explicit developer transition from any state to done. Replaces the automatic hook-triggered and CI-triggered paths. (see: requirements.md FR-03)

---

## Requirements Summary

- **Primary user need**: Remove kanban's ability to initiate git commits; keep git identity reading for attribution.
- **Feature type**: Infrastructure/Backend — modification and removal of existing behavior.
- **Walking skeleton scope**: All 5 stories ship atomically (3–4 days).

---

## Constraints Established

- `C-03`: The Go binary MUST NOT call `git commit` or `git add` anywhere. (new hard constraint, extends the existing list from CLAUDE.md)
- Atomic file writes remain mandatory (NFR-03).
- Hexagonal architecture unchanged — board-state-in-git's TransitionLogRepository port is removed cleanly.

---

## Upstream Changes (Changed Assumptions)

### Changed assumption from board-state-in-git

**Original assumption** (ADR-004 Amendment, 2026-03-18): The commit-msg hook appends to `.kanban/transitions.log` on each developer commit; `kanban board` reads from transitions.log as the single source of board state.

**New assumption**: transitions.log is removed. `kanban board` reads `status:` from task YAML front matter. The commit-msg hook is removed entirely. State changes are explicit via CLI commands.

**Rationale**: User feedback confirmed the transitions.log adds architectural complexity without justification. The task files already carry state. Automatic transitions via hooks are undesirable as they constitute an implicit side effect of developer git activity that the team does not want.

**Documents not modified**: ADR-004, ADR-005, ADR-011 (transitions log ADR) — these represent historical decisions. New ADR recommended: `ADR-013-explicit-state-transitions` to supersede ADR-004 amendment, ADR-005 amendment, and ADR-011.
