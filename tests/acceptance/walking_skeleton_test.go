package acceptance

import (
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

// TestWalkingSkeleton_FullTaskLifecycle ports the non-@skip scenario:
// "Developer completes the full task lifecycle from creation to done"
func TestWalkingSkeleton_FullTaskLifecycle(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup())
	dsl.When(ctx, dsl.IRunKanban("init"))
	dsl.Then(ctx, dsl.WorkspaceReady())
	dsl.When(ctx, dsl.IRunKanbanNew("Fix OAuth login bug"))
	dsl.Then(ctx, dsl.TaskHasStatus(ctx.LastTaskID(), "todo"))
	dsl.When(ctx, dsl.IRunKanbanBoard())
	dsl.Then(ctx, dsl.BoardShowsTaskUnder("Fix OAuth login bug", "TODO"))
	dsl.When(ctx, dsl.ICommitWithTaskID())
	dsl.Then(ctx, dsl.TaskHasStatus(ctx.LastTaskID(), "in-progress"))
	dsl.Then(ctx, dsl.BoardShowsTaskUnder(ctx.LastTaskID(), "IN PROGRESS"))
	dsl.When(ctx, dsl.CIStepRunsPass())
	dsl.Then(ctx, dsl.TaskHasStatus(ctx.LastTaskID(), "done"))
	dsl.Then(ctx, dsl.BoardShowsTaskUnder(ctx.LastTaskID(), "DONE"))
}

// TestWalkingSkeleton_EditAndDelete ports the non-@skip scenario:
// "Developer edits a task and then removes it when the work is cancelled"
func TestWalkingSkeleton_EditAndDelete(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("Migrate database schema", "todo"))
	dsl.When(ctx, dsl.IRunKanbanEditTitle(ctx.LastTaskID(), "Migrate user table schema"))
	dsl.Then(ctx, dsl.OutputContains("title"))
	dsl.Then(ctx, dsl.BoardShowsTaskUnder("Migrate user table schema", "TODO"))
	dsl.When(ctx, dsl.IRunKanbanDelete(ctx.LastTaskID(), "y"))
	dsl.Then(ctx, dsl.OutputContains("git commit"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.BoardNotListsTask("Migrate user table schema"))
}
