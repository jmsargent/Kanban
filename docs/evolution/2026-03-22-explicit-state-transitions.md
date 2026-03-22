# Explicit State Transitions — Evolution Archive

**Date**: 2026-03-22
**Feature**: explicit-state-transitions
**ADR**: ADR-013

## Summary

Removed auto-commit and transitions.log infrastructure. Task status is now
stored exclusively in YAML front matter. The binary never auto-commits (C-03).

## Changes Delivered

### Steps

| Step | Description |
|------|-------------|
| 01-01 | Added `kanban done` command and `CompleteTask` use case |
| 01-02 | Wired `done` into root command |
| 02-01 | Removed `TransitionLog` from `GetBoard` and `StartTask` |
| 02-02 | Removed `TransitionLog` from `TransitionToDone` and `GetTaskHistory` |
| 03-01 | Removed log param from all CLI adapters; rewired main.go |
| 03-02 | Added `install-hook` deprecation stub; `_hook commit-msg` no-op |
| 03-03 | `InitRepo` no longer calls `InstallHook`, `CommitFiles`, or `AppendToGitignore` |
| 04-01 | Removed `CommitFiles` and `InstallHook` from `GitPort`; deleted `TransitionToInProgress` |
| 04-02 | Deleted all transition log infrastructure; fixed Windows `syscall.Flock` build failure |

### Deleted Infrastructure

- `internal/adapters/filesystem/transition_log_adapter.go` (+ test, flock_unix, flock_windows)
- `internal/ports/transition_log.go`
- `internal/domain/transition_entry.go`
- `internal/usecases/transition_task.go` (TransitionToInProgress, TransitionTask)

### C-03 Compliance

The binary no longer auto-commits. `kanban ci-done` updates task YAML status
and exits — the developer owns all commits.

## Key Decisions

- Status authoritative from YAML `status:` field (not transitions.log)
- `install-hook` exits 1 with removal notice (hidden from help)
- `_hook commit-msg` is a no-op (exits 0 always)
- `kanban init` creates `.kanban/tasks/` dir and config only

## Test Results

- All 9 roadmap steps COMMIT/PASS
- DES integrity: all steps have complete 5-phase TDD traces
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` clean
- Feature acceptance tests: all new tests pass

## Commits

```
1ca6fb0 chore: fix outdated comment in ci_done.go
d79b4b5 fix: update ATaskWithStatus DSL helper to write status to YAML only
0fef45f feat: delete transition log infrastructure; fix Windows syscall.Flock build failure
ca586ce feat: remove CommitFiles and InstallHook from GitPort
a56e94d feat(03-03): InitRepo no longer calls git.InstallHook or git.CommitFiles
cc4bbc0 feat(03-02): install-hook deprecation and _hook commit-msg no-op
632ac5c feat(03-01): remove log param from all CLI adapters and rewire main.go
2d2c939 feat(02-02): remove TransitionLog from TransitionToDone and GetTaskHistory
d09f498 feat(02-01): remove TransitionLog from GetBoard and StartTask use cases
3eef35e feat(explicit-state-transitions): add kanban done command
862b004 feat(explicit-state-transitions): add CompleteTask use case
```
