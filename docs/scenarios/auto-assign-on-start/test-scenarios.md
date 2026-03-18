# Test Scenarios — auto-assign-on-start

## Story Reference

- US-09: Auto-assign task to developer on start
- US-09a: Warn when reassigning a task
- US-09b: Hard fail when git identity not configured

## Test File

`tests/acceptance/auto_assign_test.go`

## New DSL Steps

`tests/acceptance/dsl/assign_steps.go`

| Step | Type | Purpose |
|------|------|---------|
| `GitIdentityConfiguredAs(name)` | Setup | Override git user.name after InAGitRepo() |
| `GitIdentityUnset()` | Setup | Unset user.name + isolate global/system config |
| `ATaskAssigneeSetTo(taskID, assignee)` | Setup | Patch assignee field in existing task file |
| `TaskHasAssignee(taskID, expected)` | Assertion | Verify assignee field value |
| `TaskAssigneeRemains(taskID, expected)` | Assertion | Verify assignee unchanged (semantic alias) |
| `TaskHasNoAssignee(taskID)` | Assertion | Verify assignee is empty or absent |
| `StdoutDoesNotContain(text)` | Assertion | Negative stdout assertion |

## Scenario Inventory

| # | Test Function | AC | Category | Driving Port |
|---|-------------|-----|----------|-------------|
| 1 | `TestAutoAssign_UnassignedTaskGetsAssignedOnStart` | AC-09-1 | Walking Skeleton / Happy Path | `kanban start <id>` |
| 2 | `TestAutoAssign_PreviouslyAssignedTaskWarnsAndReassigns` | AC-09-2 | Reassignment warning | `kanban start <id>` |
| 3 | `TestAutoAssign_SameAssigneeNoWarning` | AC-09-3 | No-op warning path | `kanban start <id>` |
| 4 | `TestAutoAssign_MissingIdentityFailsAndLeavesTaskUnchanged` | AC-09-4 | Error path | `kanban start <id>` |
| 5 | `TestAutoAssign_AlreadyInProgressTaskPreservesExistingAssignee` | AC-09-5 | Idempotence | `kanban start <id>` |

## Coverage Summary

- Total scenarios: 5
- Walking skeleton / happy path: 1 (20%)
- Feature scenarios (warning, no-warning): 2 (40%)
- Error / edge scenarios: 2 (40%)
- All 5 AC criteria covered: 100%

## Step Reuse

Existing reused steps: `InAGitRepo`, `KanbanInitialised`, `ATaskWithStatusAs`, `IRunKanbanStart`,
`ExitCodeIs`, `StdoutContains`, `StderrContains`, `TaskHasStatus`, `TaskStatusRemains`.

New steps added: 7 (see table above).
