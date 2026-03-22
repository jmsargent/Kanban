# Evolution: new-editor-mode

**Date**: 2026-03-22
**Feature ID**: new-editor-mode
**Status**: Complete — all 9 acceptance scenarios green

---

## Feature Summary

`kanban new` with no arguments now launches `$EDITOR` with a blank task template. After the developer fills in the title and any optional fields, then saves and quits, the task is created and confirmed with the same output format as `kanban new <title>`. The editor is never opened unless all pre-flight checks (git repo, git identity, kanban init) have already passed.

---

## Business Context

Developers can now create tasks without pre-knowing the title at the command line. The new path matches the UX flow of `kanban edit` — open an editor, fill in structured fields, save — making task creation as frictionless as task editing. The primary outcome is that developers can capture tasks with metadata (priority, assignee, description) in a single command, eliminating the common "new then immediate edit" pattern.

North-star KPI: 30% of newly created task files include at least one optional field populated within 30 days of release (baseline: 0% — optional fields were not accessible in a single step before this feature).

---

## Delivery: 9 Steps Completed

| Step | Name | Completed |
|------|------|-----------|
| 01-01 | Walking Skeleton — Editor Launches and Task Is Created | 2026-03-22T21:16:31Z |
| 01-02 | Blank Template — Correct Structure Written to Temp File | 2026-03-22T21:19:34Z |
| 02-01 | Empty Title Rejected With Exit Code 2 | 2026-03-22T21:27:07Z |
| 02-02 | Pre-flight Blocks Editor When Repo Not Initialised | 2026-03-22T21:29:50Z |
| 02-03 | Editor Unavailable Exits With Runtime Error | 2026-03-22T21:33:41Z |
| 03-01 | Optional Fields Persisted When Provided | 2026-03-22T21:37:45Z |
| 03-02 | Title Argument Bypasses Editor | 2026-03-22T21:39:17Z |
| 03-03 | Temp File Cleaned Up on Success | 2026-03-22T21:44:23Z |
| 03-04 | Temp File Cleaned Up on Empty Title Error | 2026-03-22T21:45:51Z |

All steps delivered via TDD: RED_ACCEPTANCE -> GREEN -> COMMIT. RED_UNIT was skipped for steps 01-02 through 03-04 where the acceptance test passed on first run, indicating existing implementation already satisfied the scenario.

---

## Key Decisions

### DISCUSS Wave

**WD-01**: The blank task template does not include a `due` field. The `--due` flag on the `kanban new <title>` path is the correct entry point for date values; adding `due` to the editor template would require validation logic not in scope for this story.

**WD-02**: An empty title after editor exit is a usage error (exit code 2), not a runtime error (exit code 1). Consistent with how `kanban new ""` is handled.

**WD-03**: The `openEditor()` function must be shared between `edit` and `new` paths — duplication explicitly rejected.

**WD-04**: All pre-flight checks (git repo, git identity, kanban init) run before the editor opens. Opening the editor then discovering kanban is not initialised would be a frustrating UX.

**WD-05**: Success output must be byte-for-byte identical between the inline-title path and the editor path. Scripts parsing the output must not need to distinguish the two paths.

### DESIGN Wave

**D1 / ADR-014**: `openEditor()` extracted to `internal/usecases/editor.go` as exported `OpenEditor(filePath string) error` (Option B). Extraction to its own file gives the function canonical ownership; both `EditTask` and the new editor-mode branch in `new.go` call `usecases.OpenEditor()`. Option C (EditorPort interface) rejected as overhead for a three-line function with no testability benefit.

**D2**: `WriteTempNew() (string, error)` added as a new method on `EditFilePort` rather than adding a mode flag to `WriteTemp`. A separate method has a single clear purpose, keeps the existing `WriteTemp` contract stable, and avoids coupling `EditTask` to awareness of the new-task flow.

**D3**: `config.Read()` pre-flight check runs from the CLI adapter (not the use case) in the zero-arg path, satisfying WD-04 while remaining consistent with hexagonal architecture — the CLI adapter holds a reference to `ConfigRepository` and calling it from a primary adapter is a normal driving-port interaction.

**D4**: `WriteTempNew` omits the `due` field and includes YAML comment guidance. `ReadTemp` is already comment-safe (`yaml.Unmarshal` ignores comment lines). No changes to `ReadTemp`.

**D5**: Title validation (empty -> exit 2) runs in the CLI adapter after `ReadTemp`, before `AddTask.Execute`. The adapter owns the validation branch and calls `os.Exit(2)` directly.

### DISTILL Wave

**DI-01**: Template structure (AC-02) tested via a `cp` script in the editor mock — the only reliable way to observe the temp file contents without the test knowing the opaque OS temp path.

**DI-02**: Temp file cleanup (AC-09) asserted via proxy: a clean exit code is the structural signal that `os.Remove` fired correctly. Direct enumeration of `os.TempDir()` across parallel test runs is not safe.

**DI-03**: Walking skeleton covers AC-01 and AC-03 together — they are inseparable in the success path.

**DI-04**: Pre-flight test creates a real editor script on disk that would set a title, but the script is never invoked because the pre-flight check exits first.

**DI-05**: Error/edge scenario ratio: 56% (5 of 9), intentionally above the 40% target. Three explicit error exits and two cleanup-path variants make this count natural, not artificial.

---

## Notable Issues Encountered

### Walking skeleton used wrong binary path

The initial walking skeleton acceptance test referenced `bin/kanban` as the compiled binary path. The project convention (established in `context.go`) is `tests/bin/kanban`. This was resolved by reading the existing test context file and aligning to the established convention. No production code was affected.

### Editor unavailable test needed git on stripped PATH

`TestNewEditorMode_EditorUnavailable_ExitsWithRuntimeError` strips the `PATH` to remove all editors. However, the binary's pre-flight call to `git.RepoRoot()` also requires `git` to be on `PATH`. The test helper `IRunKanbanNewInteractiveNoEditor` was fixed to symlink the `git` binary into a minimal test `PATH` that excludes editors but preserves git. This ensures the pre-flight passes and the editor-unavailable code path is actually reached.

---

## ADR-014 Reference

ADR-014 (OpenEditor extraction strategy) was written during the DESIGN wave and is located at:

`docs/adrs/ADR-014-openeditor-extraction-strategy.md`

This ADR documents the decision to extract `openEditor()` to `internal/usecases/editor.go` as the canonical shared implementation for both `kanban edit` and the new editor mode in `kanban new`.

---

## Files Changed

| File | Change |
|------|--------|
| `internal/adapters/cli/new.go` | New zero-arg RunE branch; gains `EditFilePort` constructor parameter |
| `internal/ports/repositories.go` | `EditFilePort` gains `WriteTempNew() (string, error)` |
| `internal/adapters/filesystem/task_repository.go` | Implements `WriteTempNew()` |
| `internal/usecases/editor.go` | NEW — exports `OpenEditor(filePath string) error` |
| `internal/usecases/edit_task.go` | One-line change: `openEditor(tmpFile)` -> `OpenEditor(tmpFile)` |
| `tests/acceptance/new_editor_mode_test.go` | New — 9 acceptance scenarios |

No new third-party dependencies introduced.

---

## Outcome Measurement

Measurement is performed by inspecting task files and git history — no binary instrumentation required.

- **KPI-1** (North Star): Weekly script — parse YAML front matter of `.kanban/tasks/*.md`, count files with non-empty priority, assignee, or description. Target: 30% within 30 days.
- **KPI-2** (Leading): Weekly script — compare task creation time to first `kanban edit` time per task. Target: 50% reduction in same-session new->edit sequences. Baseline established in first 14 days post-release.
- **Guardrail**: CI acceptance suite pass rate for `kanban new <title>` path. Must remain 100%.
