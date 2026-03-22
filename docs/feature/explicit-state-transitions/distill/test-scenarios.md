# Test Scenarios — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Walking Skeleton

| Test | File | ACs |
|------|------|-----|
| TestDoneCommand_InProgressTaskTransitionsToDone | done_command_test.go | AC-01-1, AC-01-2 |

The walking skeleton proves the `kanban done` command exists as a driving port, reads YAML status, writes the updated status atomically, and exits 0 with the correct message. All other tests layer on top of this proof.

---

## US-EST-01: `kanban done` (5 scenarios)

| Scenario | Test Function | ACs |
|----------|--------------|-----|
| in-progress → done | TestDoneCommand_InProgressTaskTransitionsToDone | AC-01-1, AC-01-2 |
| todo → done | TestDoneCommand_TodoTaskTransitionsToDone | AC-01-3 |
| non-existent task exits 1 | TestDoneCommand_NonexistentTask | AC-01-4 |
| already done is idempotent | TestDoneCommand_AlreadyDoneIsIdempotent | AC-01-6 |
| no auto-commit | TestDoneCommand_NoAutoCommit | AC-01-5 |

**Driving port**: `kanban done <TASK-ID>` CLI command
**State assertion**: `TaskFileStatusIs` — reads YAML `status:` field directly, no transitions.log fallback.
**No-commit assertion**: `CaptureGitHeadSHA` before + `GitHeadSHAIs` after.

---

## US-EST-02: `kanban ci-done` (3 scenarios)

| Scenario | Test Function | ACs |
|----------|--------------|-----|
| marks task done from commit range | TestCiDone_UpdatesTaskStatusFromCommitMessages | AC-02-1, AC-02-2 |
| no tasks in range exits clean | TestCiDone_NoTasksInRangeExitsClean | AC-02-3 |
| already done task is skipped | TestCiDone_AlreadyDoneTaskIsSkipped | AC-02-4 |

**Driving port**: `kanban ci-done --since <SHA>` CLI command
**Pattern**: `CaptureGitHeadSHA(&sinceSHA)` before the referencing commit, passed to `IRunKanbanCiDoneFrom(&sinceSHA)` at step execution time.

---

## US-EST-03: `kanban board` reads YAML (3 scenarios)

| Scenario | Test Function | ACs |
|----------|--------------|-----|
| board works without transitions.log | TestBoard_ReadsStatusFromYAMLWithNoTransitionsLog | AC-03-2, AC-03-3 |
| tasks grouped by YAML status | TestBoard_GroupsTasksByYAMLStatus | AC-03-1 |
| legacy task (no status field) treated as todo | TestBoard_LegacyTaskWithNoStatusFieldTreatedAsTodo | AC-03-4 |

**Driving port**: `kanban board` CLI command
**Note**: AC-03-4 is skipped pending a `ALegacyTaskWithNoStatusField` DSL step.

---

## US-EST-04: commit-msg hook removed (3 scenarios)

| Scenario | Test Function | ACs |
|----------|--------------|-----|
| install-hook absent from --help | TestHookRemoved_InstallHookAbsentFromHelp | AC-04-1 |
| install-hook command exits 1 | TestHookRemoved_InstallHookCommandExitsWithError | AC-04-2 |
| _hook commit-msg is safe no-op | TestHookRemoved_CommitMsgHookDelegationIsNoOp | AC-04-3 |

**Driving port**: `kanban --help`, `kanban install-hook`, `kanban _hook commit-msg` CLI commands

---

## Design discovery D7: init no-commit (1 scenario)

| Scenario | Test Function | Source |
|----------|--------------|--------|
| kanban init does not auto-commit | TestInit_DoesNotAutoCommit | upstream-changes.md AC-init-1 |

**Driving port**: `kanban init` CLI command

---

## DSL Extensions

New steps added in `tests/acceptance/dsl/done_steps.go`:

| Step | Purpose |
|------|---------|
| `IRunKanbanDone(taskID)` | Runs `kanban done <taskID>` |
| `IRunKanbanHookCommitMsg(content)` | Runs `kanban _hook commit-msg <tempfile>` |
| `IRunKanbanCiDoneFrom(sha *string)` | Runs `kanban ci-done --since *sha` (evaluated at runtime) |
| `CaptureGitHeadSHA(sha *string)` | Captures current HEAD SHA for before/after comparison |
| `GitHeadSHAIs(sha *string)` | Asserts HEAD SHA unchanged (verifies no auto-commit) |
| `TransitionsLogAbsent()` | Asserts `.kanban/transitions.log` does not exist |
| `TaskFileStatusIs(taskID, expected)` | Reads YAML `status:` only — no transitions.log fallback |
| `InitDidNotAutoCommit()` | Asserts most recent commit is not a kanban init commit |
| `KanbanDotKanbanDirectoryExists()` | Asserts `.kanban/` directory was created |
