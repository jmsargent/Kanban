# Component Boundaries — board-mermaid-export

## Dependency Rule (unchanged)

All dependencies point inward. No new boundaries are introduced by this feature.

```
Primary Adapters (CLI)
        |
        v
   Use Cases  <-->  Port Interfaces
        |
        v
   Domain Core
        ^
        |
Secondary Adapters (FileSystem, Git)
```

## Affected Components

### `internal/adapters/cli` (modified)

**New file**: `board_mermaid.go`
**Modified file**: `board.go`

**Allowed imports** (unchanged from existing `board.go`):
- `internal/domain` — reads `domain.Board`, `domain.Column`, `domain.Task`, `domain.TaskStatus`
- `internal/ports` — uses `ports.GitPort`, `ports.ConfigRepository`, `ports.TaskRepository`
- `internal/usecases` — calls `usecases.NewGetBoard`
- stdlib: `fmt`, `os`, `strings`, `bytes`

**Prohibited** (enforced by go-arch-lint):
- Importing any other adapter package (`filesystem`, `git`)
- Importing external libraries not already in go.mod

### `internal/domain` (not modified)

`domain.Board`, `domain.Column`, `domain.Task` are the inputs to `renderBoardMermaid`. No changes to domain types. Zero new imports to domain.

### `internal/usecases` (not modified)

`GetBoard.Execute` is called unchanged. The use case returns `domain.Board` with no awareness of the output format chosen by the CLI layer.

### `internal/ports` (not modified)

No new port interfaces. The `--out FILE` file-writing logic is presentation-layer I/O, not a port — it does not need to be swappable behind an interface.

## Why No New Port for File Writing

The `writeMermaidToFile` function writes to the local filesystem as a direct CLI output operation, analogous to `fmt.Fprintln(os.Stdout, ...)`. It is not infrastructure that needs to be tested in isolation via a mock — the file write is tested with `t.TempDir()` in the adapter integration test layer. Creating a `FileWriterPort` for this single use would add abstraction without benefit.

## go-arch-lint Compliance

The existing `cicd/go-arch-lint.yml` rules are satisfied:
- `board_mermaid.go` is in package `cli` — same rules as `board.go`
- No cross-adapter imports introduced
- No imports from `internal/ports` into `internal/domain`
