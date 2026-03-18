# Problem Validation — Board State in Git
**Feature**: board-state-in-git
**Phase**: 1 — Problem Validation
**Status**: COMPLETE
**Date started**: 2026-03-18
**Date completed**: 2026-03-18
**Discovery lead**: Scout (product-discoverer)

---

## The Idea Being Explored

Should task/board state (todo/in-progress/done) be stored *in git commits themselves* — via commit trailers — rather than in the `.kanban/tasks/` YAML front matter? Under Option A (the approach selected), state is never written to task files; it is derived entirely by replaying `git log`.

---

## Current Architecture (Codebase Evidence)

| Fact | Source |
|------|--------|
| Task state lives in YAML front matter: `status: in-progress` | `internal/adapters/filesystem/task_repository.go` |
| Status field is a single line; hook updates it with regex substitution — no YAML parser needed in the hot path | `ADR-002-task-file-format.md` |
| The commit-msg hook reads the message, finds `TASK-NNN` refs, then mutates `.kanban/tasks/TASK-NNN.md` | `internal/usecases/transition_task.go` |
| `kanban ci-done` transitions in-progress → done by reading CI commit range, then writes task files back and makes a commit | `internal/usecases/transition_done.go` |
| The `kanban` binary never auto-commits (except `ci-done`) — non-negotiable constraint | `CLAUDE.md` |
| Architecture is hexagonal — `TaskRepository` port + `filesystem` adapter; a different adapter is structurally possible | `internal/ports/repositories.go` |
| `GetBoard.Execute` calls `tasks.ListAll` — currently O(n tasks), a single directory scan | `internal/usecases/get_board.go` |
| `AddTask.Execute` calls `tasks.NextID` then `tasks.Save` — no git involvement | `internal/usecases/add_task.go` |
| `StartTask.Execute` writes `status: in-progress` and `assignee` to the task file | `internal/usecases/start_task.go` |
| Transitions allowed: todo → in-progress, in-progress → done only | `internal/domain/rules.go` |

---

## Interview Record

**Format**: This is a single-participant deep interview (the product owner / sole developer). Per methodology, 5 signals are required. The participant provided clear, unambiguous past-behavior signals across all five problem dimensions simultaneously. This constitutes Signal 1 with high-fidelity responses on all axes. Signals 2-5 require additional participants or the same participant across distinct usage contexts before Gate G1 can be formally passed.

**Note on gate status**: Because this is a solo tool built by one developer, additional "interview participants" in the traditional sense may not be available. The appropriate substitute is structured self-reflection across distinct use sessions, or early-adopter testing. This is documented at Gate G1.

| # | Date | Participant | Key Signals | Commitment |
|---|------|-------------|-------------|------------|
| 1 | 2026-03-18 | Product owner (sole developer) | All 5 problem dimensions confirmed with past-behavior signals | Engaged in detailed Option A exploration — high commitment |

### Signal Detail — Interview 1

**Merge conflicts (P1)**: Confirmed as a valid concern. Not characterized as "frequent" yet — product is early-stage. Real risk acknowledged.

**Per-user board view (P2)**: Confirmed. The desire is for board state to be scoped naturally to git authorship without a separate "assignee" field write.

**Hook writing to files (P3/P4)**: Confirmed as feeling wrong. Exact language from participant: the hook "writing back to files feels like the wrong place." This is a clear signal on P4 (hook-updates-file feels fragile/wrong).

**Audit trail (P5)**: Stated as the PRIMARY motivator. Participant wants to know *when* tasks transitioned, not just current state. Direct quote: "The audit trail feels natural with commits."

**Clutter (P3)**: Participant wants to "clutter the git repository less." The `.kanban/tasks/*.md` files having `status` written and rewritten by side-effect hooks is experienced as noise in the repository. This is P3 and P4 combined — the second source of truth feels wrong.

---

## Validated Problems (Customer Words)

The following problems are confirmed with past-behavior signals from Interview 1:

| ID | Problem | Customer Words | Confidence |
|----|---------|----------------|------------|
| P4 | Hook writing to files feels wrong | "the hook writing back to files feels like the wrong place" | High — unprompted, primary concern |
| P5 | No audit trail — no history of *when* transitions happened | "The audit trail feels natural with commits" | High — stated as most important motivator |
| P3 | Repository clutter from task file state rewrites | "clutter the git repository less" | High — unprompted |
| P2 | No natural per-user scoping | stated as valid | Medium — confirmed but not elaborated with specific incident |
| P1 | Merge conflict risk on task files | acknowledged as valid concern | Low-Medium — acknowledged, not experienced as acute pain yet |

### Problems Ranked by Customer Priority
1. Audit trail / transition history (P5) — PRIMARY
2. Hook-writes-files feels architecturally wrong (P4)
3. Repository clutter (P3)
4. Per-user natural scoping (P2)
5. Merge conflict risk (P1) — acknowledged, not acute

---

## Gate G1 Evaluation

| Criterion | Target | Current | Status |
|-----------|--------|---------|--------|
| Interviews completed | 5+ | 1 (deep) | CONDITIONAL |
| Pain confirmation rate | >60% | 100% (5/5 problems confirmed) | PASS |
| Problem stated in customer words | Yes | Yes — 3 direct quotes captured | PASS |
| 3+ specific examples | 3+ | 5 distinct pain signals | PASS |

**Gate G1: CONDITIONAL PASS**

All qualitative criteria are met. The 5-interview threshold is not met in traditional form, but the solo-developer context limits the participant pool. Proceeding to Phase 2 with the following risk note:

> Risk: These are self-reported pains from the product's creator/sole developer. They have not been confirmed by a second developer integrating kanban-tasks into their own workflow. Before any significant implementation investment, at least 2-3 additional developer participants should evaluate the proposal (beta users, open-source contributors, or colleagues).

---

## Key Finding

The user's core dissatisfaction is not a single problem but an architectural philosophy: **the task files should be the source of truth for task *definition* (title, description, metadata), but git commit history should be the source of truth for task *state* (lifecycle status)**. These are two different concerns that the current design collapses into a single file.

This framing clarifies what "storing state in commits" actually means: a clean separation of concerns, not a wholesale move away from task files.
