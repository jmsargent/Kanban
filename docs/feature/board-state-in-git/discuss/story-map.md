# Story Map: Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## User: Jon Santos — developer using kanban-tasks in daily git workflow
## Goal: Maintain a clean, auditable record of task state transitions as part of natural git usage

---

## Backbone (User Activities — left to right)

| Create Task | Start Work | Commit Work | Complete Work (CI) | Review History | View Board |
|-------------|-----------|-------------|---------------------|----------------|------------|
| `kanban add` | `kanban start` | `git commit` (hook fires) | `kanban ci-done` | `kanban log` | `kanban board` |

Each activity is a verb phrase representing what Jon does in sequence during a typical feature cycle.

---

## Story Map

| Activity | Create Task | Start Work | Commit Work | Complete Work | Review History | View Board |
|----------|------------|-----------|-------------|---------------|----------------|------------|
| **Walking Skeleton** | Task created (existing) | Task started (existing) | Hook fires (existing) | CI done (existing) | **kanban log TASK-ID** (NEW) | Board shows state (existing) |
| | | | | | | |
| **Release 1: Audit Trail** | | | | | kanban log reads transitions.log | Board derives from log |
| (Concept B Phase A) | | Log append on start | Log append on commit | Log append on done | | |
| | | No task file write | No task file write | Commits log only | | |
| | | | | | | |
| **Release 2: Migration** | Task created without status field | | | | | |
| (Concept B Phase B) | Migration: strip status fields | Migration: backfill log from YAML | | | | |
| | | | | | | |
| **Release 3: Per-User View** | | | | | | kanban board --me |
| (Independent) | | | | | | Warns on unassigned tasks |

---

### Walking Skeleton

**Assessment: The walking skeleton is `kanban log TASK-ID`.**

This is a brownfield project. All activities in the backbone (Create Task, Start Work, Commit Work, Complete Work, View Board) already exist and work correctly. The walking skeleton is not the thinnest slice that makes the flow work — the flow already works.

Instead, the walking skeleton for this feature is the thinnest end-to-end slice that delivers the primary new value: **audit trail surfacing**.

Walking skeleton = US-BSG-01: `kanban log TASK-ID`

**Why this is the walking skeleton:**
- Delivers end-to-end value: user can run `kanban log TASK-007` and see a readable history
- Uses existing data: git history on task files already contains the information
- Touches every layer of the stack: CLI adapter → use case → git port → output
- Zero migration risk: no existing behavior changes
- Validates the highest-scored opportunity (OS-1: 16, OS-8: 14, OS-10: 14)
- Validates the core assumption: "Is the audit trail need satisfied by surfacing existing git history, or does the developer need the transitions log format?"

**The walking skeleton is complete when:**
- `kanban log TASK-007` shows a formatted history from git log
- `kanban log TASK-999` exits 1 with a clear error
- `kanban log TASK-003` shows "no transitions" gracefully
- The output uses domain language (todo/in-progress/done), not raw git messages

---

### Release 1: Architectural Audit Trail (Concept B Phase A)

**Outcome target**: Developer running `kanban board` sees state derived from `.kanban/transitions.log`; hook no longer writes to task files.

Stories:
- US-BSG-02: Append-only transition log replacing YAML status field

Activities covered: Start Work (log append), Commit Work (hook writes log not file), Complete Work (ci-done commits log), View Board (reads log for status)

**Dependencies**: Walking Skeleton (US-BSG-01) must ship first; DISCOVER validation that the audit trail need persists after Phase 1.

---

### Release 2: Migration (Concept B Phase B)

**Outcome target**: Existing repos with `status:` in task YAML can migrate to the transitions log without data loss.

Stories:
- Migration path from YAML status to transitions log (new story, not yet in scope)
- Task file template updated to remove `status:` field

**Note**: This release is a prerequisite for Concept B going live on existing repos. New repos initialized after Concept B ships do not need migration.

**Dependencies**: Release 1 (Concept B Phase A) must be complete and stable.

---

### Release 3: Per-User Board Filter (Independent)

**Outcome target**: Developer running `kanban board --me` sees only their own tasks.

Stories:
- US-BSG-03: `kanban board --me` filter

**Dependencies**: NONE. This release is independent of Releases 1 and 2. It can ship before, after, or in parallel with the transitions log work.

---

## Scope Assessment: PASS

- 3 user stories (US-BSG-01, US-BSG-02, US-BSG-03)
- 2 bounded contexts: CLI commands (primary adapter), TransitionLog storage (new secondary adapter)
- Estimated effort: US-BSG-01: 1 day | US-BSG-02: 8-10 days | US-BSG-03: 1 day | Total: 10-12 days
- Walking skeleton identified: US-BSG-01

**Elephant Carpaccio assessment**: US-BSG-02 (10 days) is at the upper boundary of a single story. However, it delivers one coherent user outcome (transitions log replaces YAML status) that cannot be meaningfully split further without breaking end-to-end value. The complexity is implementation depth, not scope breadth. The story has 5-7 UAT scenarios and is demonstrable in a single session. Assessment: ACCEPTABLE — proceed as single story with close monitoring.

If US-BSG-02 reveals hidden complexity during design (>7 scenarios, >3 bounded contexts), split into:
- US-BSG-02a: Hook writes to transitions log (no task file writes)
- US-BSG-02b: kanban board reads from transitions log
- US-BSG-02c: kanban start writes to transitions log

---

## Backbone Completeness Check

| Activity | Walking Skeleton Covered? | Note |
|----------|--------------------------|------|
| Create Task | Yes (existing) | No change needed |
| Start Work | Yes (existing + log append in R1) | kanban log shows result |
| Commit Work | Yes (existing + log append in R1) | kanban log shows result |
| Complete Work | Yes (existing + log append in R1) | kanban log shows result |
| Review History | Yes — THIS IS THE SKELETON | New: US-BSG-01 |
| View Board | Yes (existing + log-derived in R1) | No visible change to user |
