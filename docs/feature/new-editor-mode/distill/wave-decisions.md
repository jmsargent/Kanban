# Wave Decisions — DISTILL (new-editor-mode)

**Wave**: DISTILL
**Date**: 2026-03-22

---

## DI-01: Capture template via cp script rather than a dedicated DSL step

**Decision**: Template structure (AC-02) is tested by having the editor script `cp "$1" <capturePath>` before doing nothing else, then reading the capture file after the binary exits. The binary will exit 2 (empty title) as a side effect.

**Rationale**: The binary writes the temp file to `os.TempDir()` and the path is opaque to the test process. A capture script is the only reliable way to observe the template contents without the test knowing the temp path. This mirrors how the existing editor mock pattern works — the test gives the binary an EDITOR script that does something observable.

**Alternative considered**: Inject a `KANBAN_TMPDIR` env var and configure the binary to write there, so tests could enumerate files. Rejected — requires production code change solely for testability, which is test-induced design damage.

---

## DI-02: NoTempFileFromNewEditor delegates to NoTempFilesRemain

**Decision**: AC-09 (temp file cleanup) is asserted via `NoTempFilesRemain`, which checks `.kanban/tasks/` for `.tmp` files, not the system temp directory.

**Rationale**: The temp file is in `os.TempDir()`. We cannot enumerate it safely across test runs (parallel tests, different platforms, random suffixes). The clean exit code (0 or 2) is the structural proxy: if `os.Remove` panicked or was skipped, the binary would exit non-zero or produce unexpected output.

**Limitation**: Documented as Risk 3 in acceptance-review.md. The crafter may add a more direct check using a `TMPDIR` override if they judge it necessary.

---

## DI-03: Walking skeleton covers AC-01 and AC-03 together

**Decision**: The single walking skeleton covers both AC-01 (editor opens) and AC-03 (task created). AC-01 and AC-03 are inseparable in the success path — you cannot verify the editor was opened without also verifying the task was created.

**Rationale**: Splitting them would require a script that opens the editor but prevents the title from being set, which would produce exit 2 — making it an error-path test, not a walking skeleton. The walking skeleton must trace the complete success journey.

---

## DI-04: AC-08 pre-flight test uses a real editor script on disk

**Decision**: The pre-flight test (`TestNewEditorMode_KanbanNotInitialised_PreflightBlocksEditor`) creates a real editor script that would set a title, but the script is never invoked because the pre-flight check exits first.

**Rationale**: This is the strongest pre-flight test. If the binary mistakenly opens the editor before checking init status, the script will run, the task will not be created (no `.kanban/tasks/` directory), and the exit code assertion will still catch it. Having a real script on disk makes the test self-documenting.

---

## DI-05: Error path ratio target exceeded intentionally

**Decision**: 56% error/edge scenarios (5 of 9), above the 40% target.

**Rationale**: This feature has three explicit error exits (exit 2 empty title, exit 1 editor unavailable, exit 1 pre-flight failed) and one non-functional cleanup requirement (AC-09). Covering all three error exits and both cleanup paths produces 5 error-focused scenarios naturally. No scenarios were added artificially to hit the ratio.
