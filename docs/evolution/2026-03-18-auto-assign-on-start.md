# Evolution — auto-assign-on-start

**Date**: 2026-03-18
**Feature ID**: auto-assign-on-start
**Type**: Brownfield enhancement
**Status**: Delivered

---

## Feature Summary

When a developer runs `kanban start <task-id>`, the task is now automatically assigned to the developer's git identity (`git config user.name`) as part of the start transition. This eliminates a manual step and keeps the board accurate without extra effort.

### Business Context

The kanban board loses accuracy when assignees are not recorded at the moment of claiming work. Previously, a developer had to either supply `--assignee` on `kanban new` or run `kanban edit` after starting a task. Both paths were easy to skip. Auto-assignment on start closes this gap: the act of claiming a task and recording ownership becomes a single atomic operation.

Key outcome: zero-friction ownership — the board is accurate by default.

---

## Delivery Execution

**Delivered**: 2026-03-17
**Steps**: 3 (01-01 through 01-03)
**Commits**: 1 (all 3 acceptance criteria implemented in step 01-01)

| Step | Phase | Status | Notes |
|------|-------|--------|-------|
| 01-01 | PREPARE | PASS | |
| 01-01 | RED_ACCEPTANCE | PASS | |
| 01-01 | RED_UNIT | PASS | |
| 01-01 | GREEN | PASS | All 3 test scenarios satisfied in one TDD cycle |
| 01-01 | COMMIT | PASS | |
| 01-02 | PREPARE | PASS | |
| 01-02 | RED_ACCEPTANCE | SKIPPED | `TestAutoAssign_PreviouslyAssignedTaskWarnsAndReassigns` already green from step 01-01 |
| 01-02 | RED_UNIT | PASS | Verified |
| 01-02 | GREEN | PASS | Verified |
| 01-02 | COMMIT | SKIPPED | No new changes — committed in 01-01 |
| 01-03 | PREPARE | PASS | |
| 01-03 | RED_ACCEPTANCE | SKIPPED | `TestAutoAssign_MissingIdentityFailsAndLeavesTaskUnchanged` already green from step 01-01 |
| 01-03 | RED_UNIT | PASS | Verified |
| 01-03 | GREEN | PASS | Verified |
| 01-03 | COMMIT | SKIPPED | No new changes — committed in 01-01 |

---

## Key Wave Decisions

### DISCUSS Wave

| ID | Decision |
|----|---------|
| D1 | Identity resolution in CLI adapter, not use case. `start.go` calls `git.GetIdentity()` and passes `assignee string` to `StartTask.Execute`. Keeps `StartTask` free of `GitPort` dependency. |
| D2 | Hard fail on missing git identity. `kanban start` exits 1 with stderr error when `user.name` is not configured. No state change occurs. Consistent with `kanban new` (ADR-007). |
| D3 | Idempotence preserved for already-in-progress path. When task is already in-progress, no assignee update occurs. Existing behaviour unchanged. |
| D4 | Overwrite with stdout warning. When an existing assignee differs from the current git user, the assignee is overwritten and a `Note:` line is emitted to stdout (informational, not stderr). |
| D5 | Assignee format is name-only. Stored as raw `git config user.name` value. No email appended. |

### DESIGN Wave

| ID | Decision |
|----|---------|
| D1 | No new port interfaces. `GitPort.GetIdentity()` and `ports.ErrGitIdentityNotConfigured` are fully implemented. Identity is passed as plain data. |
| D2 | `StartTask.Execute` gains `assignee string` as a third positional parameter. Consistent with `AddTaskInput.CreatedBy` pattern. Simplest API change, no struct overhead. |
| D3 | `PreviousAssignee string` added to `StartTaskResult`. Use case captures pre-update assignee and returns it. Warning logic stays in adapter layer (where output decisions belong). |
| D4 | Hard-fail on identity error mirrors `new.go`. Identical pattern: check `ErrGitIdentityNotConfigured`, write to stderr, `osExit(1)` before invoking use case. Consistent with ADR-007. |
| D5 | No ADR required. All decisions are applications of existing ADRs (ADR-001, ADR-007). |

### DISTILL Wave

| ID | Decision |
|----|---------|
| D1 | Go DSL pattern, not Gherkin. All new acceptance tests follow the established Go DSL in `tests/acceptance/dsl/`. |
| D2 | `GitIdentityUnset()` approach. Identity is stripped after task setup (not before) to allow `kanban new` to succeed during fixture setup. More reliable for AC-09-4. |
| D3 | `ATaskAssigneeSetTo()` for pre-existing assignee. Task file patched directly after creation. Consistent with `ATaskWithStatusAs` pattern. |
| D4 | `StdoutDoesNotContain()` added to DSL. Required for AC-09-1 and AC-09-3 (absence of warning). |
| D5 | No upstream issues. All 5 acceptance criteria from DISCUSS are directly testable as written. |

---

## Lessons Learned

1. **Crafter implemented all 3 acceptance test scenarios in one TDD cycle (step 01-01).** The feature's scope was small and self-contained — once the core wiring was complete (`start.go` GetIdentity call + `StartTask.Execute` signature + `PreviousAssignee` field), all acceptance tests passed simultaneously. The 3-step delivery plan had anticipated incremental progress, but the implementation was holistic.

2. **Steps 01-02 and 01-03 were verification-only.** Because all acceptance tests passed in 01-01, steps 01-02 and 01-03 became confirmation phases (RED_ACCEPTANCE and COMMIT skipped as NOT_APPLICABLE). This is a valid and efficient delivery outcome — not a process gap. The TDD cycle ran once, not three times, because the change surface was minimal (two files, one new parameter, one new field).

3. **Refactoring extracted `exitError` closure.** During GREEN phase, a small refactor extracted the error-exit pattern into a local closure in `start.go`. This reduced repetition between the identity-not-configured exit path and the use-case error exit path. The refactor stayed within the file and did not change behaviour.

4. **Brownfield enhancements need no walking skeleton.** The `walking-skeleton.md` correctly identified that no skeleton was needed (base `kanban start` already existed). The DISTILL team's first test (`TestAutoAssign_UnassignedTaskGetsAssignedOnStart`) served as the minimum viable slice and drove the full implementation.

5. **DSL extensibility held up well.** Adding 7 new DSL steps to `assign_steps.go` required no changes to the DSL's wiring or test harness. The established pattern of composable step functions absorbed the new steps without friction.

---

## Migrated Permanent Artifacts

| Artifact | Permanent Location |
|----------|-------------------|
| Architecture design | `docs/architecture/auto-assign-on-start/architecture-design.md` |
| Component boundaries | `docs/architecture/auto-assign-on-start/component-boundaries.md` |
| Technology stack | `docs/architecture/auto-assign-on-start/technology-stack.md` |
| Data models | `docs/architecture/auto-assign-on-start/data-models.md` |
| Test scenarios | `docs/scenarios/auto-assign-on-start/test-scenarios.md` |
| Walking skeleton | `docs/scenarios/auto-assign-on-start/walking-skeleton.md` |
| Journey (YAML) | `docs/ux/auto-assign-on-start/journey-auto-assign.yaml` |
| Journey (visual) | `docs/ux/auto-assign-on-start/journey-auto-assign-visual.md` |
