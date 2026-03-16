# Evolution: human-task-commands

**Date**: 2026-03-16
**Type**: Rename / UX vocabulary improvement
**Feature ID**: human-task-commands

---

## Summary

Renamed the kanban CLI command `add` to `new` across the entire codebase for a more human-oriented vocabulary. The command `kanban add <title>` is now `kanban new <title>`. Five steps were completed across the CLI adapter, command registration, acceptance step definitions, acceptance feature files, and the board empty-state hint message.

---

## Motivation

The verb `add` is mechanical — it describes an implementation action (adding a record to a list). The verb `new` is human-oriented — it describes intent (creating something new). Aligning CLI vocabulary to natural language lowers cognitive friction for developers using the tool daily.

---

## Delivery Execution

**Roadmap**: 1 phase, 5 steps
**Created**: 2026-03-16T18:31:03Z
**Completed**: 2026-03-16T18:51:32Z
**Elapsed**: ~20 minutes

### Steps

| Step | Name | Outcome |
|------|------|---------|
| 01-01 | Rename add.go to new.go and update command constructor | PASS |
| 01-02 | Update root command registration | PASS |
| 01-03 | Update acceptance test step definitions | PASS |
| 01-04 | Update acceptance feature files | PASS |
| 01-05 | Update board.go user-visible hint message | PASS |

### TDD Phases per Step

| Step | RED_ACCEPTANCE | RED_UNIT | GREEN | COMMIT |
|------|---------------|----------|-------|--------|
| 01-01 | SKIPPED (pure rename; acceptance updated in 01-03/01-04) | SKIPPED (CLI adapter has no unit tests) | PASS | PASS |
| 01-02 | SKIPPED (wiring step; no acceptance scenario) | SKIPPED (verified by build + acceptance) | PASS | SKIPPED (root.go already updated in 01-01) |
| 01-03 | PASS | SKIPPED (step definitions are the test layer) | PASS | PASS |
| 01-04 | PASS | SKIPPED (feature files are test artifacts) | PASS | PASS |
| 01-05 | SKIPPED (no acceptance scenario for hint text) | SKIPPED (CLI adapter has no unit tests) | PASS | PASS |

---

## Changes Made

### `internal/adapters/cli/new.go` (renamed from `add.go`)
- Cobra `Use` field changed from `"add <title>"` to `"new <title>"`
- Exported constructor renamed from `NewAddCommand` to `NewCreateCommand`
- File renamed from `add.go` to `new.go`

### `internal/adapters/cli/root.go`
- Registration call updated from `NewAddCommand(...)` to `NewCreateCommand(...)`

### `tests/acceptance/kanban-tasks/steps/kanban_steps_test.go`
- Step pattern updated from `"kanban add"` to `"kanban new"`
- Helper function renamed from `iRunKanbanAddWithTitle` to `iRunKanbanNewWithTitle`
- Error message updated to reference `kanban new`

### `tests/acceptance/kanban-tasks/` feature files
- `walking-skeleton.feature`: all `kanban add` references replaced with `kanban new`
- `milestone-1-task-crud.feature`: all `kanban add` references replaced with `kanban new`
- `integration-checkpoints.feature`: all `kanban add` references replaced with `kanban new`
- `milestone-2-auto-transitions.feature`: no references; no change required

### `internal/adapters/cli/board.go`
- Empty-board hint updated from `"run 'kanban add <title>'"` to `"run 'kanban new <title>'"`

---

## Prior Wave Artifacts

No prior wave artifacts exist for this feature. No DISCUSS, DESIGN, or DISTILL waves were run — this was a targeted rename with no upstream design documentation.

---

## Lessons Learned

- Pure rename operations touching only CLI surface and tests are low-risk: no domain logic, no port interfaces, no use-case logic required changes.
- Acceptance feature files and step definitions are the primary test surface for CLI command names; keeping them aligned is the key quality gate for rename operations.
- The two-step pattern (rename file + update root wiring) can be collapsed into a single commit when the wiring change is trivially mechanical.
