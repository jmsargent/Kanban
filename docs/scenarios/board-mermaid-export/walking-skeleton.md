# Walking Skeleton — board-mermaid-export

## Definition

The walking skeleton is the minimum end-to-end slice that proves the feature is wired correctly through all layers: CLI flag parsing → use case execution → Mermaid renderer → stdout.

## Skeleton Test

**Test**: `TestBoardMermaid_WalkingSkeleton` (file: `tests/acceptance/board_mermaid_test.go`)

**AC covered**: AC-01

**Scenario**:
```
Given a git repo with kanban initialised and at least one task
When I run kanban board --mermaid via the CLI binary
Then the exit code is 0
And stdout begins with ```mermaid
And stdout contains "kanban" as the diagram type declaration
```

**Why this is the skeleton**: It proves the entire path is wired:
1. `--mermaid` flag is parsed by cobra ✓
2. `GetBoard` use case executes and returns a board ✓
3. `renderBoardMermaid` is called and produces output ✓
4. Output reaches stdout with correct format ✓

## Implementation Target

The crafter must implement exactly enough to pass `TestBoardMermaid_WalkingSkeleton`:
- Add `--mermaid bool` flag to `NewBoardCommand` in `board.go`
- Add `renderBoardMermaid(board)` function in `board_mermaid.go`
- Call `renderBoardMermaid` when `--mermaid` is set
- Print result to stdout

No sanitisation, no `--out`, no mutual exclusion needed for this step.

## Skeleton Status

- [ ] `TestBoardMermaid_WalkingSkeleton` — **ENABLE FIRST** (no `t.Skip`)
- [x] All other tests — skipped pending implementation
