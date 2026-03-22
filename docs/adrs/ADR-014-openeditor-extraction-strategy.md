# ADR-014: OpenEditor Extraction Strategy for new-editor-mode

## Status
Accepted

## Date
2026-03-22

## Context

The `kanban new` command is being extended so that running it with no arguments launches `$EDITOR` with a blank task template (the new-editor-mode feature). This path requires launching an external editor process — the same operation already performed by `kanban edit`.

The editor-launch logic currently lives as a package-level unexported function `openEditor(filePath string) error` in `internal/usecases/edit_task.go`. The DISCUSS wave decision WD-03 explicitly prohibits duplication: the new-editor-mode path must reuse the existing function, not implement a separate one.

The architectural question is: how should `openEditor()` be made accessible to the new code path without violating hexagonal architecture rules or introducing unjustified complexity?

The function's implementation is three lines: read `$EDITOR` env var, fall back to `vi`, call `exec.Command`. It is not domain logic (no business rules) and not an adapter (it does not implement a port interface). It is application orchestration — use case layer behaviour.

The caller in the new code path is `internal/adapters/cli/new.go`, which cannot directly reference unexported symbols from `internal/usecases`.

## Decision

Extract `openEditor()` to a new file `internal/usecases/editor.go` and export it as `OpenEditor(filePath string) error`.

The unexported `openEditor` function is removed from `edit_task.go`. The existing call site in `EditTask.Execute` is updated to reference `OpenEditor` (same package, one character change). The new editor-mode branch in `cli/new.go` calls `usecases.OpenEditor(tmpFile)`.

## Alternatives Considered

### Option A: Export in place (rename to `OpenEditor` in `edit_task.go`)
Export the function as `OpenEditor` without moving it to a new file.

**Evaluated against requirements**:
- Satisfies WD-03 (no duplication): yes
- Architecture compliance: yes
- Discoverability: poor — a function named `OpenEditor` co-located in a file named `edit_task.go` is surprising. A developer reading the codebase would not expect a shared utility to live in an unrelated use case file.
- Rejected: discoverability cost with no compensating benefit over Option B.

### Option B: Extract to `internal/usecases/editor.go` (selected)
Create a new file whose sole purpose is the editor-launch utility. Export `OpenEditor`.

**Evaluated against requirements**:
- Satisfies WD-03 (no duplication): yes
- Architecture compliance: yes — `usecases/editor.go` imports `os` and `os/exec` (stdlib only)
- Discoverability: good — a developer searching for editor behaviour finds it immediately
- go-arch-lint compliance: yes — new file is in `usecases` package, covered by existing rules
- Overhead: zero — one new file, no new interface, no new injection site
- Selected.

### Option C: `EditorPort` interface injected into both use cases
Define a new port interface `EditorPort` with a single method `Open(filePath string) error`. Implement it as a filesystem/process adapter. Inject into `EditTask` and the new-editor-mode use case.

**Evaluated against requirements**:
- Satisfies WD-03 (no duplication): yes
- Architecture compliance: yes — most architecturally pure
- Testability improvement: marginal — `openEditor` calls `exec.Command`; the acceptance test pattern (compiled binary + EDITOR=script) already provides test coverage without mocking the interface. An interface here enables substitution of the editor at test time, but the project already substitutes via the `$EDITOR` environment variable, which is simpler and sufficient.
- Overhead: one new interface in `internal/ports`, one new adapter file, injection parameter changes on two use case constructors, wiring site changes
- Complexity-to-benefit ratio: unfavourable. The interface adds surface area (two new files, three signature changes) for a function that is three lines and has no realistic alternative implementation beyond `exec.Command`.
- Rejected: adds interface overhead for a simple function where the benefit does not justify the cost.

### Option D: Duplicate
Copy the three-line implementation into the new code path.

**Evaluated against requirements**:
- Explicitly rejected by DISCUSS wave WD-03.
- Rejected: maintenance divergence risk for a function that may evolve (e.g., `VISUAL` env var support, `--wait` flag for GUI editors).

## Consequences

### Positive
- Single canonical implementation of editor-launch logic — one place to add `VISUAL` env var support, GUI editor `--wait` flags, or error wrapping in the future
- `usecases/editor.go` has a clear, discoverable purpose
- No new port interface — architectural surface area stays minimal
- `go-arch-lint` compliance maintained with no configuration change
- `EditTask` use case is substantively unchanged — one call site update, zero behavioural change

### Negative
- `OpenEditor` is now part of the `usecases` package public API. Future callers outside the package can call it. This is a minor leakage of a utility function into the package API — acceptable given Go's package-level visibility model and the fact that this function belongs in this layer.
- The `internal/usecases` package now has a file that is not a use case struct. This is a low-cost deviation from a strict "one use case per file" convention; Go packages are not required to contain only use case types.

## References
- WD-03: openEditor must be shared, not duplicated (discuss/wave-decisions.md)
- OI-01: Sharing strategy for openEditor (discuss/wave-decisions.md)
- architecture-design.md §OI-01 resolution
