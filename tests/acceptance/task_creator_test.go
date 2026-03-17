package acceptance

import (
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

// --- US-01: Automatic creator capture ---
// Driving port: kanban new (CLI command)

// AC-01-1 — Walking Skeleton
// Creator name is automatically written to the task file on kanban new.
func TestTaskCreator_CreatorRecordedOnKanbanNew(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo()) // configures git user.name = "Test User"
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNew("Fix login bug"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.StdoutContains("Created TASK-"))
	dsl.And(ctx, dsl.TaskHasCreator(ctx.LastTaskID(), "Test User"))
}

// AC-01-2
// All required front matter fields are present after task creation,
// including the new created_by field.
func TestTaskCreator_FrontMatterContainsCreatorField(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNew("Fix login bug"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	taskID := ctx.LastTaskID()
	dsl.And(ctx, dsl.TaskHasCreator(taskID, "Test User"))
	dsl.And(ctx, dsl.TaskHasStatus(taskID, "todo"))
}

// AC-01-3
// No .tmp files remain in .kanban/tasks/ after a successful task creation.
func TestTaskCreator_AtomicWriteNoTempFiles(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNew("Fix login bug"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.NoTempFilesRemain())
	dsl.And(ctx, dsl.TaskHasCreator(ctx.LastTaskID(), "Test User"))
}

// --- US-02: Creator visible on the board ---
// Driving port: kanban board (CLI command)

// AC-02-1
// The board shows the creator name in the row for a newly created task.
func TestTaskCreator_BoardShowsCreatorName(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskExists("Fix login bug"))
	taskID := ctx.LastTaskID()
	dsl.When(ctx, dsl.IRunKanbanBoard())
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.BoardRowForTaskContains(taskID, "Test User"))
}

// AC-02-2
// A task file without a created_by field (pre-existing task) is displayed
// on the board without errors, showing "--" in the creator column.
func TestTaskCreator_PreExistingTaskShowsDashOnBoard(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.APreExistingTaskWithoutCreator("Old task from before creator feature"))
	taskID := ctx.LastTaskID()
	dsl.When(ctx, dsl.IRunKanbanBoard())
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.BoardRowForTaskContains(taskID, "--"))
}

// AC-02-3
// The JSON output from kanban board --json includes a created_by field
// on every task object.
func TestTaskCreator_JSONOutputHasCreatedByField(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskExists("Fix login bug"))
	dsl.When(ctx, dsl.IRunKanbanBoardJSON())
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.OutputIsValidJSON())
	dsl.And(ctx, dsl.JSONHasFields("created_by"))
}

// --- US-03: Error guidance when git identity is not configured ---
// Driving port: kanban new (CLI command)

// AC-03-1
// When git user.name is not configured, kanban new exits with code 1
// and prints a self-contained error message with setup instructions.
func TestTaskCreator_MissingIdentityFailsWithGuidance(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepoWithoutGitIdentity())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNew("Fix login bug"))
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.And(ctx, dsl.StderrContains("git identity not configured"))
	dsl.And(ctx, dsl.StderrContains("git config --global user.name"))
}

// AC-03-2
// When git user.name is not configured, kanban new writes no task file.
func TestTaskCreator_MissingIdentityWritesNoFile(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepoWithoutGitIdentity())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNew("Fix login bug"))
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.And(ctx, dsl.TasksDirIsEmpty())
}

// --- US-04: Creator immutability ---
// Driving port: kanban edit (CLI command)

// AC-04-2
// Editing a task's title does not change the created_by field.
// (AC-04-1 — edit temp file does not expose created_by — is verified by
//  the absence of created_by in the editor's input; that invariant is
//  enforced structurally by editFields exclusion and covered by the use
//  case unit tests. The behavioral outcome — AC-04-2 — is the end-to-end proof.)
func TestTaskCreator_EditDoesNotChangeCreator(t *testing.T) {
	t.Skip("not yet implemented — step 05-01")
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskExists("Fix login bug"))
	taskID := ctx.LastTaskID()
	dsl.When(ctx, dsl.IRunKanbanEditTitle(taskID, "Fix login bug (updated)"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.TaskHasCreator(taskID, "Test User"))
}
