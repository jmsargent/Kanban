# Test Scenarios — kanban start

## Story Reference

US-08: As a developer managing work in the terminal, I want to mark a task as in-progress with a single command, so that the board reflects my active work without requiring a git commit first.

## Feature File

`tests/acceptance/kanban-tasks/milestone-3-start-command.feature`

## Scenario Inventory

| # | Scenario | Category | Story Step |
|---|----------|----------|------------|
| 1 | Developer starts a todo task and sees it move to in-progress | Walking Skeleton / Happy Path | Status todo → in-progress, stdout "Started <id>: <title>", exit 0 |
| 2 | Starting a task that is already in-progress reports no change needed | Edge Case | Status in-progress → no change, stdout "already in progress", exit 0 |
| 3 | Starting a completed task is rejected as an error | Error Path | Status done → rejected, stderr "already finished", exit 1 |
| 4 | Starting a task that does not exist reports a clear error | Error Path | Task ID not found → stderr "not found", exit 1 |
| 5 | Starting a task without kanban initialised reports setup guidance | Error Path | Not initialised → stderr "kanban not initialised", exit 1 |

## Coverage Summary

- Total scenarios: 5
- Happy path / walking skeleton: 1 (20%)
- Error / edge scenarios: 4 (80%) — exceeds the 40% target
- All 5 behavior spec cases covered

## Step Reuse

All `Then` assertions reuse existing steps (`output contains`, `the exit code is`, `the task "..." has status "..."`, `the task "..." status remains "..."`).

Two new `When` steps were added: `I run "kanban start" on that task` and `I run "kanban start" on task "<id>"`.
