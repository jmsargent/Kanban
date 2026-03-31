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
	"testing"

	dsl "github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// ---------------------------------------------------------------------------
// Walking Skeleton — AC-01
// ---------------------------------------------------------------------------

// TestBoardMermaid_WalkingSkeleton validates AC-01:
// "kanban board --mermaid" exits 0, stdout begins with a fenced mermaid block,
// and stdout contains "kanban" as the diagram type declaration.
// This is the minimum end-to-end slice proving the feature is wired correctly.
func TestBoardMermaid_WalkingSkeleton(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskExists("title: Fix login bug"))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaid())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.StdoutStartsWithMermaidFence())
	dsl.And(ctx, dsl.StdoutContainsMermaidKanbanType())
}

// ---------------------------------------------------------------------------
// Milestone 1 — Core export (US-01, US-02)
// ---------------------------------------------------------------------------

// TestBoardMermaid_AllColumnsAppearAsSections validates AC-02:
// all configured columns appear as Mermaid section headers in the order
// defined by the board configuration.
func TestBoardMermaid_AllColumnsAppearAsSections(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaid())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	// Default config has three columns: todo, in-progress, done.
	dsl.And(ctx, dsl.StdoutContainsMermaidSection("To Do"))
	dsl.And(ctx, dsl.StdoutContainsMermaidSection("In Progress"))
	dsl.And(ctx, dsl.StdoutContainsMermaidSection("Done"))
}

// TestBoardMermaid_TasksAppearAsNodesUnderTheirColumn validates AC-03:
// each task appears as a Mermaid node under its column section,
// and each node includes the task ID and title.
func TestBoardMermaid_TasksAppearAsNodesUnderTheirColumn(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("title: Fix login bug", "status: todo"))
	todoTaskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.ATaskWithStatus("title: Write docs", "status: in-progress"))
	inProgressTaskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaid())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.StdoutContainsMermaidNode(todoTaskID))
	dsl.And(ctx, dsl.StdoutContainsMermaidNode(inProgressTaskID))
	// Both node labels should contain the task titles.
	dsl.And(ctx, dsl.StdoutContains("text: Fix login bug"))
	dsl.And(ctx, dsl.StdoutContains("text: Write docs"))
}

// TestBoardMermaid_EmptyBoardProducesValidMermaidBlock validates AC-04:
// when no tasks exist, "kanban board --mermaid" exits 0 and produces a valid
// Mermaid kanban block with column sections and no task nodes.
func TestBoardMermaid_EmptyBoardProducesValidMermaidBlock(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.NoTasksExist())

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaid())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.StdoutStartsWithMermaidFence())
	dsl.And(ctx, dsl.StdoutContainsMermaidKanbanType())
	dsl.And(ctx, dsl.StdoutContainsMermaidSection("To Do"))
	// No task nodes appear (the output should not contain "@{").
	dsl.And(ctx, dsl.StdoutDoesNotContain("text: @{"))
}

// ---------------------------------------------------------------------------
// Milestone 2 — Composability (US-03, US-04)
// ---------------------------------------------------------------------------

// TestBoardMermaid_MeFilterApplies validates AC-05:
// "kanban board --mermaid --me" shows only tasks assigned to the current
// git user and excludes tasks assigned to others.
func TestBoardMermaid_MeFilterApplies(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// InAGitRepo sets git user.email = "test@example.com".
	dsl.Given(ctx, dsl.TaskAssignedTo("Fix OAuth login bug", "test@example.com"))
	myTaskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskAssignedTo("Refactor billing module", "colleague@example.com"))
	otherTaskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaidWithMe())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.StdoutContainsMermaidNode(myTaskID))
	dsl.And(ctx, dsl.StdoutDoesNotContainMermaidNode(otherTaskID))
}

// TestBoardMermaid_MutuallyExclusiveWithJSON validates AC-06:
// "kanban board --mermaid --json" exits 2, writes "mutually exclusive" to stderr,
// and produces no stdout output.
func TestBoardMermaid_MutuallyExclusiveWithJSON(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaidAndJSON())

	dsl.Then(ctx, dsl.ExitsWithCode(2))
	dsl.And(ctx, dsl.StderrContains("text: mutually exclusive"))
	dsl.And(ctx, dsl.StdoutIsEmpty())
}

// ---------------------------------------------------------------------------
// Milestone 3 — Sanitisation (US-05, US-06)
// ---------------------------------------------------------------------------

// TestBoardMermaid_TaskTitlesWithUnsafeCharsAreSanitised validates AC-07:
// task titles containing Mermaid-unsafe characters (", [, ]) are sanitised
// so the output is syntactically valid Mermaid.
func TestBoardMermaid_TaskTitlesWithUnsafeCharsAreSanitised(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// Title contains double quotes and square brackets — Mermaid-unsafe.
	dsl.Given(ctx, dsl.ATaskExists(`Fix "login" bug [urgent]`))
	taskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaid())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	// The task node must appear.
	dsl.And(ctx, dsl.StdoutContainsMermaidNode(taskID))
	// The unsafe characters must not appear unescaped inside the label value.
	// Double-quoted word from title must not appear verbatim in output.
	dsl.And(ctx, dsl.StdoutDoesNotContain("text: \"login\""))
	// Square brackets from title must not appear verbatim in output.
	dsl.And(ctx, dsl.StdoutDoesNotContain("text: [urgent]"))
}

// TestBoardMermaid_ColumnLabelsWithSpecialCharsAreSafe validates AC-08:
// column labels containing characters that could break Mermaid section syntax
// are safely handled — the command exits 0 and produces valid output.
// Note: the default config uses plain labels; this test injects a custom config.
func TestBoardMermaid_ColumnLabelsWithSpecialCharsAreSafe(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	// Write a custom config with a column label containing a colon.
	dsl.Given(ctx, dsl.FileExistsWithContent(".kanban/config", `task_prefix: TASK
columns:
  - name: todo
    label: "In Progress: Active"
  - name: done
    label: Done
`))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaid())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.StdoutStartsWithMermaidFence())
	// Output must not be blank after the fence.
	dsl.And(ctx, dsl.StdoutContainsMermaidKanbanType())
}

// ---------------------------------------------------------------------------
// Milestone 4 — File output (US-07)
// ---------------------------------------------------------------------------

// TestBoardMermaid_OutCreatesFileWhenNotExists validates AC-09:
// "kanban board --mermaid --out BOARD.md" creates the file when it does not
// exist, writes the Mermaid block to it, and produces no stdout.
func TestBoardMermaid_OutCreatesFileWhenNotExists(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskExists("title: Fix login bug"))
	// BOARD.md does not exist yet (InAGitRepo creates README.md, not BOARD.md).

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaidWithOut("BOARD.md"))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.FileExists("BOARD.md"))
	dsl.And(ctx, dsl.FileContainsText("BOARD.md", "```mermaid"))
	dsl.And(ctx, dsl.FileContainsText("BOARD.md", "kanban"))
	dsl.And(ctx, dsl.StdoutIsEmpty())
}

// TestBoardMermaid_OutErrorsWhenFileExistsWithNoKanbanBlock validates AC-10:
// "kanban board --mermaid --out README.md" exits 1 and writes a descriptive
// error to stderr when README.md exists but contains no Mermaid kanban block.
// The file must not be modified.
func TestBoardMermaid_OutErrorsWhenFileExistsWithNoKanbanBlock(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// README.md is created by InAGitRepo with content "# test\n" — no kanban block.
	originalContent := "# test\n"

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaidWithOut("README.md"))

	dsl.Then(ctx, dsl.ExitsWithCode(1))
	dsl.And(ctx, dsl.StderrContains("text: README.md"))
	// Error must tell the user what placeholder to add.
	dsl.And(ctx, dsl.StderrContains("text: mermaid"))
	// File must be unchanged.
	dsl.And(ctx, dsl.FileContentEquals("README.md", originalContent))
}

// TestBoardMermaid_OutReplacesExistingKanbanBlockInPlace validates AC-11:
// "kanban board --mermaid --out README.md" replaces the existing Mermaid kanban
// block in README.md and preserves all other content in the file.
func TestBoardMermaid_OutReplacesExistingKanbanBlockInPlace(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskExists("title: Fix login bug"))
	taskID := ctx.LastTaskID()
	// Write a README with an existing mermaid kanban block.
	dsl.Given(ctx, dsl.FileExistsWithKanbanBlock("README.md"))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardMermaidWithOut("README.md"))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	// New task must appear in the updated block.
	dsl.And(ctx, dsl.FileContainsText("README.md", taskID))
	// Old placeholder task must be gone.
	dsl.And(ctx, dsl.FileDoesNotContainText("README.md", "TASK-OLD"))
	// Surrounding content must be preserved.
	dsl.And(ctx, dsl.FileContainsText("README.md", "# My Project"))
	dsl.And(ctx, dsl.FileContainsText("README.md", "Some text above."))
	dsl.And(ctx, dsl.FileContainsText("README.md", "Some text below."))
	dsl.And(ctx, dsl.StdoutIsEmpty())
}

// TestBoardMermaid_OutWithoutMermaidIsUsageError validates AC-12:
// "kanban board --out README.md" (without --mermaid) exits 2 and writes
// "--out requires --mermaid" to stderr.
func TestBoardMermaid_OutWithoutMermaidIsUsageError(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardWithOut("README.md"))

	dsl.Then(ctx, dsl.ExitsWithCode(2))
	dsl.And(ctx, dsl.StderrContains("text: --out requires --mermaid"))
}
