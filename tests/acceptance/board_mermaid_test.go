package acceptance

// board_mermaid_test.go — Acceptance tests for the board-mermaid-export feature.
//
// User stories: US-01 through US-07 (board-mermaid-export)
// Driving port: kanban binary invoked as subprocess via DSL.
//
// Implementation order: enable one test at a time, implement until it passes,
// then enable the next. The walking skeleton test (TestBoardMermaid_WalkingSkeleton)
// is enabled first — it gates all subsequent work.

import (
    . "github.com/jmsargent/kanban/tests/acceptance/dsl"

	"testing"

)

// ---------------------------------------------------------------------------
// Walking Skeleton — AC-01
// ---------------------------------------------------------------------------

// TestBoardMermaid_WalkingSkeleton validates AC-01:
// "kanban board --mermaid" exits 0, stdout begins with a fenced mermaid block,
// and stdout contains "kanban" as the diagram type declaration.
// This is the minimum end-to-end slice proving the feature is wired correctly.
func TestBoardMermaid_WalkingSkeleton(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskExists("title: Fix login bug"))

	When(ctx, DeveloperRunsKanbanBoardMermaid())

	Then(ctx, ExitsSuccessfully())
	And(ctx, StdoutStartsWithMermaidFence())
	And(ctx, StdoutContainsMermaidKanbanType())
}

// ---------------------------------------------------------------------------
// Milestone 1 — Core export (US-01, US-02)
// ---------------------------------------------------------------------------

// TestBoardMermaid_AllColumnsAppearAsSections validates AC-02:
// all configured columns appear as Mermaid section headers in the order
// defined by the board configuration.
func TestBoardMermaid_AllColumnsAppearAsSections(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())

	When(ctx, DeveloperRunsKanbanBoardMermaid())

	Then(ctx, ExitsSuccessfully())
	// Default config has three columns: todo, in-progress, done.
	And(ctx, StdoutContainsMermaidSection("To Do"))
	And(ctx, StdoutContainsMermaidSection("In Progress"))
	And(ctx, StdoutContainsMermaidSection("Done"))
}

// TestBoardMermaid_TasksAppearAsNodesUnderTheirColumn validates AC-03:
// each task appears as a Mermaid node under its column section,
// and each node includes the task ID and title.
func TestBoardMermaid_TasksAppearAsNodesUnderTheirColumn(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatus("title: Fix login bug", "status: todo"))
	todoTaskID := ctx.LastTaskID()
	Given(ctx, ATaskWithStatus("title: Write docs", "status: in-progress"))
	inProgressTaskID := ctx.LastTaskID()

	When(ctx, DeveloperRunsKanbanBoardMermaid())

	Then(ctx, ExitsSuccessfully())
	And(ctx, StdoutContainsMermaidNode(todoTaskID))
	And(ctx, StdoutContainsMermaidNode(inProgressTaskID))
	// Both node labels should contain the task titles.
	And(ctx, StdoutContains("text: Fix login bug"))
	And(ctx, StdoutContains("text: Write docs"))
}

// TestBoardMermaid_EmptyBoardProducesValidMermaidBlock validates AC-04:
// when no tasks exist, "kanban board --mermaid" exits 0 and produces a valid
// Mermaid kanban block with column sections and no task nodes.
func TestBoardMermaid_EmptyBoardProducesValidMermaidBlock(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, NoTasksExist())

	When(ctx, DeveloperRunsKanbanBoardMermaid())

	Then(ctx, ExitsSuccessfully())
	And(ctx, StdoutStartsWithMermaidFence())
	And(ctx, StdoutContainsMermaidKanbanType())
	And(ctx, StdoutContainsMermaidSection("To Do"))
	// No task nodes appear (the output should not contain "@{").
	And(ctx, StdoutDoesNotContain("text: @{"))
}

// ---------------------------------------------------------------------------
// Milestone 2 — Composability (US-03, US-04)
// ---------------------------------------------------------------------------

// TestBoardMermaid_MeFilterApplies validates AC-05:
// "kanban board --mermaid --me" shows only tasks assigned to the current
// git user and excludes tasks assigned to others.
func TestBoardMermaid_MeFilterApplies(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	// InAGitRepo sets git user.email = "test@example.com".
	Given(ctx, TaskAssignedTo("Fix OAuth login bug", "test@example.com"))
	myTaskID := ctx.LastTaskID()
	Given(ctx, TaskAssignedTo("Refactor billing module", "colleague@example.com"))
	otherTaskID := ctx.LastTaskID()

	When(ctx, DeveloperRunsKanbanBoardMermaidWithMe())

	Then(ctx, ExitsSuccessfully())
	And(ctx, StdoutContainsMermaidNode(myTaskID))
	And(ctx, StdoutDoesNotContainMermaidNode(otherTaskID))
}

// TestBoardMermaid_MutuallyExclusiveWithJSON validates AC-06:
// "kanban board --mermaid --json" exits 2, writes "mutually exclusive" to stderr,
// and produces no stdout output.
func TestBoardMermaid_MutuallyExclusiveWithJSON(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())

	When(ctx, DeveloperRunsKanbanBoardMermaidAndJSON())

	Then(ctx, ExitsWithCode(2))
	And(ctx, StderrContains("text: mutually exclusive"))
	And(ctx, StdoutIsEmpty())
}

// ---------------------------------------------------------------------------
// Milestone 3 — Sanitisation (US-05, US-06)
// ---------------------------------------------------------------------------

// TestBoardMermaid_TaskTitlesWithUnsafeCharsAreSanitised validates AC-07:
// task titles containing Mermaid-unsafe characters (", [, ]) are sanitised
// so the output is syntactically valid Mermaid.
func TestBoardMermaid_TaskTitlesWithUnsafeCharsAreSanitised(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	// Title contains double quotes and square brackets — Mermaid-unsafe.
	Given(ctx, ATaskExists(`Fix "login" bug [urgent]`))
	taskID := ctx.LastTaskID()

	When(ctx, DeveloperRunsKanbanBoardMermaid())

	Then(ctx, ExitsSuccessfully())
	// The task node must appear.
	And(ctx, StdoutContainsMermaidNode(taskID))
	// The unsafe characters must not appear unescaped inside the label value.
	// Double-quoted word from title must not appear verbatim in output.
	And(ctx, StdoutDoesNotContain("text: \"login\""))
	// Square brackets from title must not appear verbatim in output.
	And(ctx, StdoutDoesNotContain("text: [urgent]"))
}

// TestBoardMermaid_ColumnLabelsWithSpecialCharsAreSafe validates AC-08:
// column labels containing characters that could break Mermaid section syntax
// are safely handled — the command exits 0 and produces valid output.
// Note: the default config uses plain labels; this test injects a custom config.
func TestBoardMermaid_ColumnLabelsWithSpecialCharsAreSafe(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	// Write a custom config with a column label containing a colon.
	Given(ctx, FileExistsWithContent(".kanban/config", `task_prefix: TASK
columns:
  - name: todo
    label: "In Progress: Active"
  - name: done
    label: Done
`))

	When(ctx, DeveloperRunsKanbanBoardMermaid())

	Then(ctx, ExitsSuccessfully())
	And(ctx, StdoutStartsWithMermaidFence())
	// Output must not be blank after the fence.
	And(ctx, StdoutContainsMermaidKanbanType())
}

// ---------------------------------------------------------------------------
// Milestone 4 — File output (US-07)
// ---------------------------------------------------------------------------

// TestBoardMermaid_OutCreatesFileWhenNotExists validates AC-09:
// "kanban board --mermaid --out BOARD.md" creates the file when it does not
// exist, writes the Mermaid block to it, and produces no stdout.
func TestBoardMermaid_OutCreatesFileWhenNotExists(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskExists("title: Fix login bug"))
	// BOARD.md does not exist yet (InAGitRepo creates README.md, not BOARD.md).

	When(ctx, DeveloperRunsKanbanBoardMermaidWithOut("BOARD.md"))

	Then(ctx, ExitsSuccessfully())
	And(ctx, FileExists("BOARD.md"))
	And(ctx, FileContainsText("BOARD.md", "```mermaid"))
	And(ctx, FileContainsText("BOARD.md", "kanban"))
	And(ctx, StdoutIsEmpty())
}

// TestBoardMermaid_OutErrorsWhenFileExistsWithNoKanbanBlock validates AC-10:
// "kanban board --mermaid --out README.md" exits 1 and writes a descriptive
// error to stderr when README.md exists but contains no Mermaid kanban block.
// The file must not be modified.
func TestBoardMermaid_OutErrorsWhenFileExistsWithNoKanbanBlock(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	// README.md is created by InAGitRepo with content "# test\n" — no kanban block.
	originalContent := "# test\n"

	When(ctx, DeveloperRunsKanbanBoardMermaidWithOut("README.md"))

	Then(ctx, ExitsWithCode(1))
	And(ctx, StderrContains("text: README.md"))
	// Error must tell the user what placeholder to add.
	And(ctx, StderrContains("text: mermaid"))
	// File must be unchanged.
	And(ctx, FileContentEquals("README.md", originalContent))
}

// TestBoardMermaid_OutReplacesExistingKanbanBlockInPlace validates AC-11:
// "kanban board --mermaid --out README.md" replaces the existing Mermaid kanban
// block in README.md and preserves all other content in the file.
func TestBoardMermaid_OutReplacesExistingKanbanBlockInPlace(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskExists("title: Fix login bug"))
	taskID := ctx.LastTaskID()
	// Write a README with an existing mermaid kanban block.
	Given(ctx, FileExistsWithKanbanBlock("README.md"))

	When(ctx, DeveloperRunsKanbanBoardMermaidWithOut("README.md"))

	Then(ctx, ExitsSuccessfully())
	// New task must appear in the updated block.
	And(ctx, FileContainsText("README.md", taskID))
	// Old placeholder task must be gone.
	And(ctx, FileDoesNotContainText("README.md", "TASK-OLD"))
	// Surrounding content must be preserved.
	And(ctx, FileContainsText("README.md", "# My Project"))
	And(ctx, FileContainsText("README.md", "Some text above."))
	And(ctx, FileContainsText("README.md", "Some text below."))
	And(ctx, StdoutIsEmpty())
}

// TestBoardMermaid_OutWithoutMermaidIsUsageError validates AC-12:
// "kanban board --out README.md" (without --mermaid) exits 2 and writes
// "--out requires --mermaid" to stderr.
func TestBoardMermaid_OutWithoutMermaidIsUsageError(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())

	When(ctx, DeveloperRunsKanbanBoardWithOut("README.md"))

	Then(ctx, ExitsWithCode(2))
	And(ctx, StderrContains("text: --out requires --mermaid"))
}
