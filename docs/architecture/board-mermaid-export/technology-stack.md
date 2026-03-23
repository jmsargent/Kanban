# Technology Stack — board-mermaid-export

## No New Technologies

This feature is a brownfield extension of the existing `kanban board` command. No new languages, frameworks, or libraries are required.

## Existing Stack (unchanged)

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go 1.22+ | Project-wide. Implicit interface satisfaction enables hexagonal architecture. |
| CLI framework | cobra (Apache 2.0) | Existing — new flags added to existing `board` command via `cmd.Flags()`. |
| String formatting | `fmt`, `strings`, `bytes` | Stdlib — sufficient for Mermaid block generation and string scanning. |
| File I/O | `os` | Stdlib — `os.ReadFile`, `os.WriteFile`, `os.Rename`, `os.Stat`. |

## Mermaid Syntax

The Mermaid kanban diagram format is a **documentation concern**, not a library dependency. The renderer produces Mermaid v11 syntax using string formatting only. No Mermaid library is used at runtime.

- Diagram type: `kanban` (Mermaid v11+)
- Node format: `TASK-NNN@{ label: "Title text" }`
- GitHub ships Mermaid v11 — no version action required (documented in DISCUSS wave-decisions.md D-Technology)

## Constraint: No New Dependencies (NFR-01)

The Go module (`go.mod`) must not gain any new `require` entries as a result of this feature. All rendering is achieved with Go's standard library.
