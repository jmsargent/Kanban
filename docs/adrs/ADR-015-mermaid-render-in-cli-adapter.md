# ADR-015 — Mermaid rendering stays in CLI adapter, split into board_mermaid.go

**Date**: 2026-03-23
**Status**: Accepted
**Feature**: board-mermaid-export

---

## Context

The `kanban board --mermaid` feature requires:
1. A Mermaid kanban diagram renderer (pure string generation from `domain.Board`)
2. A file writer that can create a new file or replace an existing Mermaid block in-place

DISCUSS wave requirement NFR-02 states: "The implementation MUST follow the existing hexagonal pattern: new rendering logic stays in the CLI adapter (`internal/adapters/cli/board.go`). No new use case or port is required."

Three structural options exist for placing this code:

1. **Append to `board.go`** — keeps all board-related code in one file
2. **New file `board_mermaid.go` in package `cli`** — dedicated file, same package
3. **New use case `RenderMermaidBoard`** — extract rendering into `internal/usecases`

---

## Decision

**Option 2**: Create `internal/adapters/cli/board_mermaid.go` within package `cli`.

---

## Rationale

**Option 1 rejected**: `board.go` is 178 lines. Adding ~150 lines of renderer, sanitisation helpers, and file-writer logic creates a 328-line file with two distinct concerns (command wiring and Mermaid output). The `cli` package already uses multiple files per command (`board.go`, `add.go`, `edit.go`). A dedicated file is consistent with this convention.

**Option 3 rejected**: Mermaid rendering is a presentation concern — it produces a string for a specific output format. It has no business logic, no domain rules, and no port dependency. Placing it in `internal/usecases` would violate the principle that use cases contain application logic, not presentation formatting. It would also contradict NFR-02's explicit instruction.

**Option 2 chosen**: `board_mermaid.go` in package `cli` satisfies NFR-02 (rendering is in the CLI adapter), keeps `board.go` readable, and requires zero architectural change. The file shares the `cli` package, so all functions are package-visible to `board.go`'s `RunE` closure without any exported surface.

---

## Consequences

- `board_mermaid.go` contains: `renderBoardMermaid`, `sanitiseMermaidTitle`, `sanitiseMermaidLabel`, `writeMermaidToFile`
- `board.go` adds two flags (`--mermaid`, `--out`) and dispatch logic only (~25 additional lines)
- go-arch-lint rules are unaffected — `board_mermaid.go` inherits the same import constraints as `board.go`
- `board_mermaid_test.go` provides unit tests for the pure renderer and sanitisation functions; integration tests cover `writeMermaidToFile` via `t.TempDir()`
