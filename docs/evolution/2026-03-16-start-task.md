# Evolution: start-task

**Date**: 2026-03-16
**Type**: New feature — intent-based CLI command
**Feature ID**: start-task

---

## Summary

Added `kanban start <task-id>` to the CLI. The command explicitly transitions a task from `todo` to `in-progress`, moving the CLI API from passive CRUD verbs (`add`, `delete`) toward intent-based commands that reflect developer workflow. The feature spans two new files: `internal/usecases/start_task.go` and `internal/adapters/cli/start.go`.

---

## Motivation

Prior to this feature the only way to record that work had begun on a task was the automatic `commit-msg` hook transition. `kanban start` gives developers an explicit, intentional signal they are beginning a task — useful when opening a task in an editor, creating a branch, or beginning a focused work session before any commit exists. It also reinforces the CLI's vocabulary of intent rather than implementation actions.

---

## Delivery Execution

**Roadmap**: 1 phase, 2 steps + 1 refactor pass + 1 review-fix pass
**Created**: 2026-03-16T19:07:02Z
**Completed**: 2026-03-16T19:37:43Z
**Elapsed**: ~31 minutes

### Steps

| Step | Name | Outcome |
|------|------|---------|
| 01-01 | StartTask use case | PASS |
| 01-02 | start CLI adapter and root registration | PASS |
| refactor | L1–L4 refactoring pass | PASS |
| review-fix | Fix adversarial review findings | PASS |

### TDD Phases per Step

| Step | RED_ACCEPTANCE | RED_UNIT | GREEN | COMMIT |
|------|---------------|----------|-------|--------|
| 01-01 | PASS | PASS | PASS | PASS |
| 01-02 | PASS | PASS | PASS | PASS |
| refactor | SKIPPED (no new behaviour) | SKIPPED (no new behaviour) | PASS | PASS |
| review-fix | SKIPPED (no new behaviour) | SKIPPED (no new behaviour) | PASS | PASS |

---

## Key Decisions

### Already-in-progress is NOT an error — returns typed result, exit 0

Running `kanban start` on a task that is already in progress is idempotent: the command prints an informational message to stdout and exits 0. This was modelled as a typed `StartTaskResult{AlreadyInProgress: true}` rather than an error. Rationale: the developer's intent (ensure the task is in progress) is already satisfied — returning an error would punish the user for a valid workflow where they re-run the command or execute a script that calls `start` unconditionally.

### Already-done returns wrapped ErrInvalidTransition, exit 1

Running `kanban start` on a task that is `done` is a genuine semantic error — the workflow has moved past in-progress and cannot return. This returns `fmt.Errorf("task %s: %w", taskID, ports.ErrInvalidTransition)` from the use case, and the CLI adapter maps it to stderr + exit 1. The distinction between "already-in-progress (informational)" and "already-done (error)" reflects real developer intent: in-progress is the desired state; done is a final state that should not be silently ignored.

### Exit code pattern aligns with codebase standard (os.Exit calls in RunE)

The existing CLI commands (`done`, `board`, etc.) call `os.Exit` directly inside `RunE` rather than returning a non-nil error, to avoid cobra printing a redundant "Error: ..." suffix to stderr. The `start` command follows this same pattern: each error path writes its own message to `errOut` then calls `osExit(1)`. A package-level `var osExit = os.Exit` variable allows tests to capture exit-code intent without terminating the test process.

### errCommandFailed sentinel introduced then removed

During the refactoring pass, a `errCommandFailed` sentinel error was briefly introduced as a typed return from `RunE` to signal that the command had already written its own error message and cobra should suppress its own. The sentinel was removed during the adversarial review fix because it added complexity without benefit: the `SilenceErrors: true` cobra flag already suppresses cobra's own error output, and returning `nil` from `RunE` after calling `osExit` is sufficient. Removing the sentinel simplified the control flow back to the established codebase pattern.

---

## Issues Encountered and Resolutions

### Issue 1: Metadata fields included in use-case unit test

An initial unit test for `StartTask` asserted `UpdatedAt` timestamp fields on the mutated task. This was identified during the adversarial review as fragile and environment-sensitive. The fix removed the timestamp assertion, leaving the test focused on the fields the use case actually owns (`Status`, `Transitioned` result flag).

### Issue 2: osExit variable and SetOsExit helper scope

The test-override mechanism for `osExit` needed to be package-accessible without leaking into production API. The `var osExit = os.Exit` declaration in `start.go` and exported `SetOsExit` helper serve this purpose. The helper was added only after the acceptance tests required capturing exit codes — it was not part of the initial design.

---

## Implementation Files

| File | Role |
|------|------|
| `internal/usecases/start_task.go` | `StartTask` use case + `StartTaskResult` type |
| `internal/usecases/start_task_test.go` | Five unit tests using in-memory mocks |
| `internal/adapters/cli/start.go` | `NewStartCommand` cobra command, `osExit` override, `writeLine` helper |
| `internal/adapters/cli/root.go` | Registration of `start` subcommand on root cobra command |

---

## Prior Wave Artifacts

No prior wave artifacts exist for this feature. This was a direct-to-DELIVER feature with no DISCUSS, DESIGN, or DISTILL wave documentation.

---

## Lessons Learned

- Typing the "already-in-progress" case as a result field rather than an error surface (`AlreadyInProgress bool`) keeps the use case error contract focused on genuine failures and makes the CLI adapter's branching logic straightforward to read.
- The `osExit` variable pattern for CLI test isolation is established in this codebase and should be used consistently across all commands that call `os.Exit` in `RunE`.
- Introducing a sentinel error during refactoring to signal "already handled" is a common instinct but adds indirection. When `SilenceErrors: true` is set and the command writes its own messages, returning `nil` is cleaner.
- Adversarial review caught a fragile timestamp assertion in unit tests — asserting only the fields the unit under test actually writes keeps tests stable under infrastructure changes.
