# DISTILL Decisions — explicit-state-transitions

**Date**: 2026-03-22

---

## Key Decisions

- **[D1] State assertion via `TaskFileStatusIs`**: New step reads YAML `status:` directly (no transitions.log fallback). This is the correct assertion for the new architecture where YAML is the sole state source. `TaskHasStatus` (which falls back to transitions.log) is intentionally NOT used for the new tests. (see: done_steps.go)

- **[D2] No-commit assertion pattern**: `CaptureGitHeadSHA(&sha)` before the command + `GitHeadSHAIs(&sha)` after. The `*string` pointer evaluates at step execution time, not at step construction time — required because steps are constructed eagerly but run lazily. (see: done_steps.go, done_command_test.go)

- **[D3] `IRunKanbanCiDoneFrom(sha *string)`**: Added to `done_steps.go` rather than extending the existing `DeveloperRunsKanbanCiDone(since string)`. The pointer variant is necessary because the `since` SHA is not known at test setup time — it is captured by a prior `CaptureGitHeadSHA` step. (see: explicit_state_transitions_test.go)

- **[D4] AC-03-4 skipped**: The "legacy task with no status: field" test requires a DSL step to create a task without a `status:` field. All current setup steps inject a status field. Skipped with `t.Skip` — can be unblocked by adding `ALegacyTaskWithNoStatusField` to setup.go in DELIVER. (see: explicit_state_transitions_test.go)

- **[D5] `install-hook` removal tested as exit-1**: AC-04-2 specifies exit 1 + message. The cobra command is retained in the binary as an error handler (not deleted from the cobra tree) to emit the removal message. This is consistent with D8 from DESIGN (which specifies `_hook` is retained as no-op). (see: explicit_state_transitions_test.go)

- **[D6] Init no-commit test uses dual assertion**: Both `GitHeadSHAIs` (precise: no new commit SHA) and `InitDidNotAutoCommit` (semantic: log message check) are used together. The SHA check is authoritative; the log message check provides a readable failure message. (see: explicit_state_transitions_test.go)

---

## Test Coverage Summary

- **Total scenarios**: 15 (14 active, 1 skipped)
- **Walking skeleton**: `TestDoneCommand_InProgressTaskTransitionsToDone`
- **Milestones**: done command (5), board reads YAML (2), ci-done no-commit (3), hook removed (3), init no-commit (1)
- **Test framework**: internal Go BDD DSL (ADR-006) — compiled binary as subprocess
- **Integration approach**: real services — compiled kanban binary + real git repo via `t.TempDir()`

---

## Constraints Established

- Tests MUST NOT import `internal/` packages — driving port only (binary as subprocess)
- `TaskFileStatusIs` is the canonical assertion for new tests; `TaskHasStatus` is for legacy tests that include transitions.log scenarios
- AC-05-1 through AC-05-4 (build and arch-lint) are structural — validated by the existing CI pipeline, not by acceptance tests

---

## Upstream Issues

None. All DISCUSS acceptance criteria are testable as written. The only deviation is AC-03-4 (skipped pending DSL helper), which is a test infrastructure gap, not a requirements gap.

The design-time discovery (D7: init also commits) introduced AC-init-1, which is covered by `TestInit_DoesNotAutoCommit` in `explicit_state_transitions_test.go`.
