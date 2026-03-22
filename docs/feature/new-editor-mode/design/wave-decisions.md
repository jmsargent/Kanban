# DESIGN Decisions — new-editor-mode

**Wave**: DESIGN
**Date**: 2026-03-22

---

## Key Decisions

- [D1] Extract `openEditor()` to `internal/usecases/editor.go` as exported `OpenEditor(filePath string) error` (Option B): The function is already a use-case concern; extraction to its own file with a clear name gives it canonical ownership without adding interface overhead. Both `EditTask` and the new editor-mode branch in `new.go` call `usecases.OpenEditor()`. (see: ADR-014, architecture-design.md §OI-01)

- [D2] Add `WriteTempNew() (string, error)` as a new method on `EditFilePort` rather than adding a mode flag to `WriteTemp`: A separate method has a single clear purpose, is independently testable, keeps the existing `WriteTemp` contract stable, and avoids coupling `EditTask` to awareness of the new-task flow. (see: architecture-design.md §OI-02, component-boundaries.md §EditFilePort)

- [D3] Config pre-flight check (`config.Read`) runs from the CLI adapter before the editor opens in the zero-arg path: This satisfies WD-04 (all pre-flight checks before editor launches) and is consistent with hexagonal architecture — the CLI adapter (primary adapter) is permitted to call driven ports. `AddTask.Execute` retains its own `config.Read` call for the inline-title path; this is redundant but harmless for the editor-mode path. (see: architecture-design.md §OI-03)

- [D4] `WriteTempNew` omits the `due` field and includes YAML comment guidance: WD-01 excludes `due` from the editor template. `ReadTemp` is already comment-safe (`yaml.Unmarshal` ignores comment lines). No changes to `ReadTemp`. (see: architecture-design.md §OI-04, data-models.md)

- [D5] Title validation (empty → exit 2) runs in the CLI adapter after `ReadTemp`, before `AddTask.Execute`: WD-02 requires exit code 2 for empty title. The CLI adapter owns the validation branch and calls `os.Exit(2)` directly, consistent with the existing inline-path pattern. (see: architecture-design.md §Execution Flow)

---

## Architecture Summary

- Pattern: modular monolith with ports-and-adapters (hexagonal) — unchanged
- Paradigm: OOP (Go implicit interface satisfaction) — unchanged
- Key components modified:
  - `internal/adapters/cli/new.go` — new zero-arg RunE branch
  - `internal/ports/repositories.go` — `EditFilePort` gains `WriteTempNew()`
  - `internal/adapters/filesystem/task_repository.go` — implements `WriteTempNew()`
  - `internal/usecases/editor.go` — NEW: exports `OpenEditor()`
  - `internal/usecases/edit_task.go` — one-line change to call `OpenEditor()`

---

## Technology Stack

- Go 1.22+: existing choice, unchanged
- cobra: existing choice, unchanged
- gopkg.in/yaml.v3: existing choice, unchanged; used by `WriteTempNew`
- No new dependencies introduced

---

## Constraints Established

- C-06 (new): The `WriteTempNew()` template must not include a `due` field (WD-01). If `due` is later added to the editor template, a new ADR is required.
- C-07 (new): The zero-arg editor-mode path must produce byte-identical success output to the inline-title path (WD-05). The success-printing block must be shared or duplicated identically in `new.go RunE`.

All prior constraints (C-01 through C-05) remain active.

---

## Upstream Changes

No DISCUSS wave assumptions were changed. All five DISCUSS decisions (WD-01 through WD-05) are carried forward as architectural constraints. All four open items (OI-01 through OI-04) are resolved in this wave.
