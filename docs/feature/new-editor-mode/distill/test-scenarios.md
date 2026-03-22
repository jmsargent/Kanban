# Test Scenarios — new-editor-mode

**Wave**: DISTILL
**Date**: 2026-03-22

---

## Scenario Table

| # | Test Function | AC Coverage | Skip Status | Scenario Type |
|---|---------------|-------------|-------------|---------------|
| 1 | `TestNewEditorMode_WalkingSkeleton_TaskCreated` | AC-01, AC-03 | **ACTIVE (RED)** | Walking skeleton |
| 2 | `TestNewEditorMode_BlankTemplate_StructureCorrect` | AC-02 | pending | Focused — boundary |
| 3 | `TestNewEditorMode_OptionalFields_Persisted` | AC-04 | pending | Focused — happy path |
| 4 | `TestNewEditorMode_EmptyTitle_RejectedWithExitCode2` | AC-05 | pending | Focused — error path |
| 5 | `TestNewEditorMode_TitleArgument_NoEditorOpened` | AC-06 | pending | Focused — regression |
| 6 | `TestNewEditorMode_EditorUnavailable_ExitsWithRuntimeError` | AC-07 | pending | Focused — error path |
| 7 | `TestNewEditorMode_KanbanNotInitialised_PreflightBlocksEditor` | AC-08 | pending | Focused — error path |
| 8 | `TestNewEditorMode_TempFileCleanedUpOnSuccess` | AC-09 (success) | pending | Focused — boundary |
| 9 | `TestNewEditorMode_TempFileCleanedUpOnEmptyTitle` | AC-09 (error) | pending | Focused — error path |

**Total scenarios**: 9
**Walking skeletons**: 1
**Focused scenarios**: 8
**Error/edge scenarios**: 5 (scenarios 4, 6, 7, 8, 9) — 56% of total, exceeds 40% target

---

## AC Traceability

| AC | Description | Covered By |
|----|-------------|------------|
| AC-01 | No-argument invocation routes to editor mode | Scenario 1 (walking skeleton) |
| AC-02 | Blank task template presented in editor | Scenario 2 |
| AC-03 | Task created and confirmed after valid editor session | Scenario 1 (walking skeleton) |
| AC-04 | Optional fields are persisted when filled | Scenario 3 |
| AC-05 | Empty title after editor exits → exit code 2 | Scenario 4 |
| AC-06 | Existing kanban new behaviour unchanged | Scenario 5 |
| AC-07 | Runtime error when $EDITOR unavailable → exit code 1 | Scenario 6 |
| AC-08 | Pre-flight checks run before editor opens | Scenario 7 |
| AC-09 | Temp file cleaned up after editor exits | Scenarios 8 and 9 |

All 9 ACs have corresponding test coverage.

---

## Implementation Order for DELIVER Wave

Enable one test at a time. Suggested order follows the outside-in implementation path:

1. **Scenario 1** — ACTIVE NOW. Drives the complete feature skeleton: CLI routing, WriteTempNew, OpenEditor, ReadTemp, AddTask, success output.
2. **Scenario 2** — Enable after template format is settled. Validates template structure observable to editor.
3. **Scenario 4** — Enable after empty-title validation is implemented. Simplest error path.
4. **Scenario 7** — Enable after pre-flight ordering is verified.
5. **Scenario 6** — Enable to confirm $EDITOR-unavailable path is wired.
6. **Scenario 3** — Enable after field persistence is confirmed working.
7. **Scenario 5** — Enable as regression guard for existing `kanban new <title>` path.
8. **Scenario 8** — Enable to confirm deferred cleanup on success path.
9. **Scenario 9** — Enable to confirm deferred cleanup on error path.
