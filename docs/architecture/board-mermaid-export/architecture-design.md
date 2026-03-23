# Architecture Design — board-mermaid-export

## System Overview

`board-mermaid-export` is a brownfield CLI extension. The existing `kanban board` command gains two new flags: `--mermaid` (switches output format to a Mermaid kanban diagram) and `--out FILE` (writes the output into a file instead of stdout). No new domain concepts, ports, or use cases are introduced.

The primary constraint from DISCUSS is NFR-02: rendering logic stays in the CLI adapter. This is consistent with the existing pattern — `printBoardJSON` lives alongside `printBoard` in `internal/adapters/cli/board.go`.

---

## C4 System Context

```mermaid
C4Context
    title System Context — kanban board --mermaid

    Person(dev, "Developer", "Runs kanban CLI on local machine or in CI")
    System(kanban, "kanban CLI", "Git-native kanban task management. Reads tasks from .kanban/tasks/, renders board in multiple formats.")
    System_Ext(readme, "README.md", "Markdown file in the git repository. Rendered by GitHub as a visual kanban board.")

    Rel(dev, kanban, "kanban board --mermaid [--out README.md]", "CLI invocation")
    Rel(kanban, readme, "Writes Mermaid kanban block", "atomic file write")
```

---

## C4 Container

```mermaid
C4Container
    title Container — kanban binary

    Person(dev, "Developer")

    Container_Boundary(bin, "kanban binary (Go)") {
        Component(board_cmd, "board command", "cobra.Command in board.go", "Parses --mermaid, --out, --json, --me flags. Validates mutual exclusions. Dispatches to renderer.")
        Component(get_board, "GetBoard use case", "usecases.GetBoard", "Loads board config and all tasks, groups tasks by column. Returns domain.Board.")
        Component(mermaid_renderer, "Mermaid renderer", "board_mermaid.go (package cli)", "renderBoardMermaid(): pure function, Board -> string. Sanitisation helpers for titles and labels.")
        Component(file_writer, "File writer", "board_mermaid.go (package cli)", "writeMermaidToFile(): create new / replace in-place / error on missing block. Atomic write via .tmp + rename.")
        Component(config_repo, "ConfigRepository", "filesystem adapter", "Reads .kanban/config.yml. Returns column definitions.")
        Component(task_repo, "TaskRepository", "filesystem adapter", "Reads .kanban/tasks/*.md. Returns []domain.Task.")
    }

    System_Ext(readme, "README.md / any file", "Target file for --out")

    Rel(dev, board_cmd, "kanban board --mermaid [--out FILE]", "CLI")
    Rel(board_cmd, get_board, "Execute(repoRoot, filterAssignee)")
    Rel(get_board, config_repo, "Read(repoRoot)")
    Rel(get_board, task_repo, "ListAll(repoRoot)")
    Rel(board_cmd, mermaid_renderer, "renderBoardMermaid(board)")
    Rel(board_cmd, file_writer, "writeMermaidToFile(filename, content)")
    Rel(file_writer, readme, "write to .tmp, os.Rename to target")
```

---

## Component Design

### New file: `internal/adapters/cli/board_mermaid.go` (package `cli`)

Mermaid-specific logic is placed in a dedicated file within the `cli` package rather than appended to `board.go`. `board.go` is currently 178 lines; adding ~150 lines of Mermaid rendering and file-writing logic would compromise readability. The `cli` package already spans multiple files (`board.go`, `add.go`, `edit.go`, etc.) — a dedicated `board_mermaid.go` follows this established convention. The hexagonal boundary is unchanged: the file lives in `internal/adapters/cli`.

**Functions exported within the package:**

| Function | Signature | Notes |
|----------|-----------|-------|
| `renderBoardMermaid` | `(board domain.Board) string` | Pure function — no I/O. Returns a complete fenced Mermaid block. |
| `sanitiseMermaidTitle` | `(s string) string` | Removes/replaces `"`, `[`, `]`, newline. Called per task title. |
| `sanitiseMermaidLabel` | `(s string) string` | Strips characters that break Mermaid `section` header syntax (colon). |
| `writeMermaidToFile` | `(filename, content string) error` | Handles create / in-place replace / missing-block error. Returns typed errors for exit code routing. |

### Modified file: `internal/adapters/cli/board.go`

Add two flags to `NewBoardCommand`:
- `--mermaid bool` — switches output format to Mermaid
- `--out string` — file path; requires `--mermaid`

Add mutual-exclusion checks in `RunE` before rendering (both return exit 2).

---

## Rendering Algorithm

### `renderBoardMermaid(board domain.Board) string`

```
output  = "```mermaid\n"
output += "kanban\n"
for each col in board.Columns:
    output += "  section " + sanitiseMermaidLabel(col.Label) + "\n"
    for each task in board.Tasks[TaskStatus(col.Name)]:
        output += "    " + task.ID + "@{ label: \"" + sanitiseMermaidTitle(task.Title) + "\" }\n"
output += "```\n"
return output
```

### `writeMermaidToFile(filename, content string) error`

```
1. os.Stat(filename)
   → os.IsNotExist: write content atomically to filename; return nil

2. os.ReadFile(filename) → fileBytes

3. Scan lines for fenced Mermaid kanban block:
   - Find line equal to "```mermaid"
   - Within that fence, find line equal to "kanban"
   - Find the closing "```"
   → Not found: return ErrNoKanbanBlock (caller exits 1 with descriptive message)

4. Reconstruct file: lines before opening fence + content + lines after closing fence

5. os.WriteFile(filename+".tmp", reconstructed, 0644)
   os.Rename(filename+".tmp", filename)
   return nil
```

### Sanitisation rules

| Input character | Replacement |
|-----------------|-------------|
| `"` | `'` (straight single quote) |
| `[` | `(` |
| `]` | `)` |
| `\n`, `\r` | ` ` (space) |
| `:` (label only) | ` ` (space) |

---

## Mutual Exclusion Validation

Validation order in `RunE` (before use case execution):

1. `--out` provided without `--mermaid` → `fmt.Fprintln(os.Stderr, "--out requires --mermaid")` + `os.Exit(2)`
2. `--mermaid` and `--json` both provided → `fmt.Fprintln(os.Stderr, "--mermaid and --json are mutually exclusive")` + `os.Exit(2)`

---

## Architectural Compliance

| Rule | Status |
|------|--------|
| `internal/domain` has zero imports from adapters/usecases | Not affected — no domain changes |
| `internal/usecases` has zero imports from adapters | Not affected — no use case changes |
| No adapter imports another adapter | `board_mermaid.go` is in `cli` package, imports `domain` only |
| No new external dependencies | `strings`, `os`, `fmt`, `bytes` — all stdlib |
| Atomic writes | `writeMermaidToFile` uses `.tmp` + `os.Rename` consistent with project pattern |
| Exit code convention | 0=success, 1=runtime error, 2=usage error — enforced explicitly |
| No new ports or use cases | `domain.Board` flows from existing `GetBoard` use case unchanged |
