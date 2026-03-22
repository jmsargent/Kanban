# Walking Skeleton — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Skeleton Test

**File**: `tests/acceptance/done_command_test.go`
**Function**: `TestDoneCommand_InProgressTaskTransitionsToDone`

```
Given a git repo
  And kanban is initialised
  And a task "Fix OAuth login bug" with status "in-progress" (TASK-001)
When I run kanban done TASK-001
Then exit code is 0
 And output contains "moved in-progress -> done"
 And task file TASK-001 has YAML status "done"
```

---

## What the skeleton proves

1. **`kanban done` is wired** — the cobra command exists and is reachable via the compiled binary
2. **State is read from YAML** — `CompleteTask` use case reads `domain.Task.Status` from the task struct populated by `TaskRepository.ListAll()`
3. **State is written to YAML** — `TaskRepository.Update()` persists the new `status: done` field atomically
4. **Output contract** — stdout contains the transition message the AC specifies
5. **Exit code** — 0 on success

These are the five minimal proofs that the feature is end-to-end wired from CLI adapter → use case → domain → repository adapter.

---

## Skeleton implementation order (for DELIVER)

The skeleton test passes only when ALL of the following are in place:

1. `internal/usecases/complete_task.go` — new `CompleteTask` use case (mirrors `StartTask`)
2. `internal/adapters/cli/done.go` — new `done` cobra command wired to `CompleteTask`
3. `cmd/kanban/main.go` — `done` command registered in `NewRootCommand`
4. `internal/adapters/filesystem/task_repository.go` — `Update()` writes `status: done` atomically (already present from `StartTask`; verify it handles `done` status)

The `TransitionLogRepository` and `GitPort.CommitFiles` are NOT needed for the skeleton — the new use case does not touch them.

---

## After skeleton: milestone ordering

| Milestone | Tests | Prerequisite |
|-----------|-------|-------------|
| 1. Done command | done_command_test.go (all 5) | Walking skeleton |
| 2. Board reads YAML | TestBoard_* (2 passing) | Skeleton + remove log fallback from GetBoard |
| 3. ci-done no-commit | TestCiDone_* (3) | CompleteTask use case wired to ci-done |
| 4. Hook removed | TestHookRemoved_* (3) | install-hook and _hook commands changed |
| 5. Init no-commit | TestInit_DoesNotAutoCommit | InitRepo: remove CommitFiles + InstallHook calls |
| 6. Cleanup | AC-05-1 through AC-05-4 | Delete TransitionLogAdapter, flock files, transition_entry |
