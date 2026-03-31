package acceptance

import (
    . "github.com/jmsargent/kanban/tests/acceptance/dsl"

	"testing"

)

// --- US-01: Automatic creator capture ---
// Driving port: kanban new (CLI command)

// AC-01-1 — Walking Skeleton
// Creator name is automatically written to the task file on kanban new.
func TestTaskCreator_CreatorRecordedOnKanbanNew(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo()) // configures git user.name = "Test User"
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNew("title: Fix login bug"))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, StdoutContains("text: Created TASK-"))
	And(ctx, TaskHasCreator("task: "+ctx.LastTaskID(), "creator: Test User"))
}

// AC-01-2
// All required front matter fields are present after task creation,
// including the new created_by field.
func TestTaskCreator_FrontMatterContainsCreatorField(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNew("title: Fix login bug"))
	Then(ctx, ExitCodeIs("code: 0"))
	taskID := ctx.LastTaskID()
	And(ctx, TaskHasCreator("task: "+taskID, "creator: Test User"))
	And(ctx, TaskHasStatus("task: "+taskID, "status: todo"))
}

// AC-01-3
// No .tmp files remain in .kanban/tasks/ after a successful task creation.
func TestTaskCreator_AtomicWriteNoTempFiles(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNew("title: Fix login bug"))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, NoTempFilesRemain())
	And(ctx, TaskHasCreator("task: "+ctx.LastTaskID(), "creator: Test User"))
}

// --- US-02: Creator visible on the board ---
// Driving port: kanban board (CLI command)

// AC-02-1
// The board shows the creator name in the row for a newly created task.
func TestTaskCreator_BoardShowsCreatorName(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskExists("title: Fix login bug"))
	taskID := ctx.LastTaskID()
	When(ctx, IRunKanbanBoard())
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, BoardRowForTaskContains("task: "+taskID, "text: Test User"))
}

// AC-02-2
// A task file without a created_by field (pre-existing task) is displayed
// on the board without errors, showing "--" in the creator column.
func TestTaskCreator_PreExistingTaskShowsDashOnBoard(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, APreExistingTaskWithoutCreator("title: Old task from before creator feature"))
	taskID := ctx.LastTaskID()
	When(ctx, IRunKanbanBoard())
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, BoardRowForTaskContains("task: "+taskID, "text: --"))
}

// AC-02-3
// The JSON output from kanban board --json includes a created_by field
// on every task object.
func TestTaskCreator_JSONOutputHasCreatedByField(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskExists("title: Fix login bug"))
	When(ctx, IRunKanbanBoardJSON())
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, OutputIsValidJSON())
	And(ctx, JSONHasFields("fields: created_by"))
}

// --- US-03: Error guidance when git identity is not configured ---
// Driving port: kanban new (CLI command)

// AC-03-1
// When git user.name is not configured, kanban new exits with code 1
// and prints a self-contained error message with setup instructions.
func TestTaskCreator_MissingIdentityFailsWithGuidance(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepoWithoutGitIdentity())
	Given(ctx, KanbanInitialised())
	Given(ctx, GitIdentityUnconfigured())
	When(ctx, IRunKanbanNew("title: Fix login bug"))
	Then(ctx, ExitCodeIs("code: 1"))
	And(ctx, StderrContains("text: git identity not configured"))
	And(ctx, StderrContains("text: git config --global user.name"))
}

// AC-03-2
// When git user.name is not configured, kanban new writes no task file.
func TestTaskCreator_MissingIdentityWritesNoFile(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepoWithoutGitIdentity())
	Given(ctx, KanbanInitialised())
	Given(ctx, GitIdentityUnconfigured())
	When(ctx, IRunKanbanNew("title: Fix login bug"))
	Then(ctx, ExitCodeIs("code: 1"))
	And(ctx, TasksDirIsEmpty())
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
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskExists("title: Fix login bug"))
	taskID := ctx.LastTaskID()
	When(ctx, IRunKanbanEditTitle("task: "+taskID, "title: Fix login bug (updated)"))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, TaskHasCreator("task: "+taskID, "creator: Test User"))
}
