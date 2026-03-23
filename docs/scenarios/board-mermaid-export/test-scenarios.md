# Test Scenarios — board-mermaid-export

## Driving Port

All scenarios invoke `kanban board` as a compiled binary subprocess. No internal package calls. Exit code, stdout, stderr, and named file contents are the observable outcomes.

## Framework

Custom Go DSL (`tests/acceptance/dsl/`) matching the project's existing acceptance test pattern. New DSL steps added in `tests/acceptance/dsl/mermaid_steps.go`.

## Scenario Inventory

| Test Function | AC | Story | Status |
|---------------|----|-------|--------|
| `TestBoardMermaid_WalkingSkeleton` | AC-01 | US-01 | **enabled** (walking skeleton) |
| `TestBoardMermaid_AllColumnsAppearAsSections` | AC-02 | US-01 | t.Skip |
| `TestBoardMermaid_TasksAppearAsNodesUnderTheirColumn` | AC-03 | US-01 | t.Skip |
| `TestBoardMermaid_EmptyBoardProducesValidMermaidBlock` | AC-04 | US-02 | t.Skip |
| `TestBoardMermaid_MeFilterApplies` | AC-05 | US-03 | t.Skip |
| `TestBoardMermaid_MutuallyExclusiveWithJSON` | AC-06 | US-04 | t.Skip |
| `TestBoardMermaid_TaskTitlesWithUnsafeCharsAreSanitised` | AC-07 | US-05 | t.Skip |
| `TestBoardMermaid_ColumnLabelsWithSpecialCharsAreSafe` | AC-08 | US-06 | t.Skip |
| `TestBoardMermaid_OutCreatesFileWhenNotExists` | AC-09 | US-07 | t.Skip |
| `TestBoardMermaid_OutErrorsWhenFileExistsWithNoKanbanBlock` | AC-10 | US-07 | t.Skip |
| `TestBoardMermaid_OutReplacesExistingKanbanBlockInPlace` | AC-11 | US-07 | t.Skip |
| `TestBoardMermaid_OutWithoutMermaidIsUsageError` | AC-12 | US-07 | t.Skip |

Total: 12 scenarios | Walking skeleton: 1 | Pending: 11

## Implementation Order (recommended)

1. **AC-01** — Walking skeleton: wires the flag and basic renderer ← enable first
2. **AC-02** — Column sections: validates column ordering
3. **AC-03** — Task nodes: validates task rendering with ID and title
4. **AC-04** — Empty board: validates empty-board guard
5. **AC-06** — Mutual exclusion: small, isolated flag validation
6. **AC-05** — `--me` filter: composability with existing flag
7. **AC-07** — Title sanitisation: correctness of unsafe char handling
8. **AC-08** — Column label sanitisation: similar to AC-07
9. **AC-12** — `--out` without `--mermaid`: small flag validation
10. **AC-09** — `--out` new file creation: file I/O
11. **AC-10** — `--out` error on missing block: block detection error path
12. **AC-11** — `--out` in-place replace: full block replacement

## DSL New Steps

File: `tests/acceptance/dsl/mermaid_steps.go`

**Actions** (invoke binary):
- `DeveloperRunsKanbanBoardMermaid()` — `kanban board --mermaid`
- `DeveloperRunsKanbanBoardMermaidWithMe()` — `kanban board --mermaid --me`
- `DeveloperRunsKanbanBoardMermaidWithOut(filename)` — `kanban board --mermaid --out <abs-path>`
- `DeveloperRunsKanbanBoardWithOut(filename)` — `kanban board --out <abs-path>`
- `DeveloperRunsKanbanBoardMermaidAndJSON()` — `kanban board --mermaid --json`

**Setup** (filesystem):
- `FileExistsWithContent(filename, content)` — write file at repoDir/filename
- `FileExistsWithKanbanBlock(filename)` — write file with mermaid kanban block + surrounding content

**Assertions** (observable outcomes):
- `StdoutStartsWithMermaidFence()` — stdout begins with ` ```mermaid`
- `StdoutContainsMermaidKanbanType()` — stdout has `kanban` as a standalone line
- `StdoutContainsMermaidSection(label)` — stdout has `section <label>`
- `StdoutContainsMermaidNode(taskID)` — stdout has `TASK-NNN@{`
- `StdoutDoesNotContainMermaidNode(taskID)` — negation of above
- `StdoutIsEmpty()` — lastStdout is empty
- `FileExists(filename)` — file exists at repoDir/filename
- `FileContainsText(filename, text)` — file contains substring
- `FileDoesNotContainText(filename, text)` — file does not contain substring
- `FileContentEquals(filename, expected)` — file has exact content

## Notes on AC-10 Precondition

`InAGitRepo()` creates `README.md` with content `# test\n`. This makes AC-10 naturally exercisable: `README.md` exists but has no Mermaid kanban block, so `--out README.md` should exit 1 with a descriptive error.

## Notes on AC-09 Filename

AC-09 uses `BOARD.md` rather than `README.md` to avoid collision with the `README.md` that `InAGitRepo()` creates. The file does not exist, so `--out BOARD.md` should create it.
