# Acceptance Criteria — explicit-state-transitions

**Feature**: Explicit State Transitions
**Date**: 2026-03-22

---

## US-EST-01: `kanban done`

**AC-01-1**: Given a task with status `in-progress`, when `kanban done TASK-NNN` is run, then the task's YAML front matter `status:` field is updated to `done` atomically (write-to-tmp + rename).

**AC-01-2**: Given `kanban done TASK-NNN` completes successfully, then stdout contains `kanban: TASK-NNN moved in-progress -> done` and the process exits 0.

**AC-01-3**: Given a task with status `todo`, when `kanban done TASK-NNN` is run, then stdout contains `kanban: TASK-NNN moved todo -> done` and exits 0.

**AC-01-4**: Given `kanban done NONEXISTENT`, then stderr contains an error message, no files are modified, and the process exits 1.

**AC-01-5**: Given `kanban done TASK-NNN` completes, then no `git add`, `git commit`, or any git subprocess is invoked.

**AC-01-6**: Given a task already at status `done`, when `kanban done TASK-NNN` is run, then stdout contains `kanban: TASK-NNN already done` and exits 0 (idempotent).

---

## US-EST-02: `kanban ci-done` (no commit)

**AC-02-1**: Given `kanban ci-done --since=<base-ref>` is run, when commit messages in `<base-ref>..HEAD` reference TASK-NNN, then TASK-NNN's `status:` is updated to `done` in its YAML front matter.

**AC-02-2**: Given `kanban ci-done` completes, then no `git add`, `git commit`, or staging subprocess is invoked.

**AC-02-3**: Given no task IDs are found in the commit range, then `kanban ci-done` exits 0 with no output.

**AC-02-4**: Given a task referenced in commits is already `done`, then `kanban ci-done` skips it (idempotent — does not double-write).

**AC-02-5**: Given `NO_COLOR` is set or no TTY is detected, then all output is plain text with no ANSI escape codes.

**AC-02-6**: Given `kanban ci-done` runs in a repo with no kanban tasks, then it exits 0 with no output.

---

## US-EST-03: `kanban board` reads from YAML

**AC-03-1**: Given task files with varying `status:` values, when `kanban board` is run, then tasks are grouped into TODO / IN PROGRESS / DONE columns derived solely from their YAML `status:` field.

**AC-03-2**: Given no transitions.log exists, then `kanban board` runs without error.

**AC-03-3**: Given a task file has `status: in-progress`, then `kanban board` places that task in the IN PROGRESS column.

**AC-03-4**: Given a task file has no `status:` field (legacy task), then `kanban board` treats it as `todo`.

---

## US-EST-04: Commit-msg hook removed

**AC-04-1**: Given the updated binary, when `kanban --help` is run, then `install-hook` does not appear in the command list.

**AC-04-2**: Given the updated binary, when `kanban install-hook` is run, then exit 1 with message: `install-hook has been removed; kanban no longer manages git hooks`.

**AC-04-3**: Given the updated binary, when `kanban _hook commit-msg <msg-file>` is run (e.g. by a leftover hook on a developer's machine), then the process exits 0 immediately with no side effects (safe no-op for backwards compatibility).

**AC-04-4**: Given the `.kanban/hooks/` directory exists in the repo, it is deleted as part of this feature.

---

## US-EST-05: transitions.log and TransitionLogRepository removed

**AC-05-1**: Given the build runs, then `internal/adapters/filesystem/transition_log_adapter.go` does not exist in the compiled binary.

**AC-05-2**: Given `go-arch-lint check` runs, then it passes with no references to the removed packages.

**AC-05-3**: Given `go build ./...` runs targeting `windows_amd64`, then the build succeeds (no `syscall.Flock` references remain).

**AC-05-4**: Given `go test ./...` runs, then all tests pass with the TransitionLogRepository mocks removed.
