# Walking Skeleton — auto-assign-on-start

## Base Feature

The `kanban start` command (US-08) is fully implemented. This feature is a brownfield enhancement — no walking skeleton is needed.

## Minimum Viable Slice

**Test**: `TestAutoAssign_UnassignedTaskGetsAssignedOnStart` (AC-09-1)

This is the first test to implement and pass. It covers:
- `GetIdentity()` called before use case
- `assignee string` passed to `StartTask.Execute`
- `task.Assignee` set and persisted
- `StartTaskResult.PreviousAssignee` populated (empty in this case)
- Normal stdout output unchanged

## Implementation Order

Once the walking skeleton test passes, implement remaining tests in order:

| Order | Test | Requires |
|-------|------|---------|
| 1 | `TestAutoAssign_UnassignedTaskGetsAssignedOnStart` | Core wiring (start.go + start_task.go) |
| 2 | `TestAutoAssign_PreviouslyAssignedTaskWarnsAndReassigns` | PreviousAssignee field + Note: output |
| 3 | `TestAutoAssign_SameAssigneeNoWarning` | PreviousAssignee == assignee suppresses warning |
| 4 | `TestAutoAssign_MissingIdentityFailsAndLeavesTaskUnchanged` | Hard fail before use case call |
| 5 | `TestAutoAssign_AlreadyInProgressTaskPreservesExistingAssignee` | AlreadyInProgress path unchanged |

## Unit Test Updates Required

Before or alongside the acceptance tests, the following unit tests must be updated
(signature change: `Execute(repoRoot, taskID)` → `Execute(repoRoot, taskID, assignee string)`):

- `internal/usecases/start_task_test.go` — all 5 existing tests
- `internal/adapters/cli/start_test.go` — all existing tests (fake GitPort gains `GetIdentity()` stub)
