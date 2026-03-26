package acceptance

import (
	"testing"

	dsl "github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// TestWalkingSkeleton_EditAndDelete ports the non-@skip scenario:
// "Developer edits a task and then removes it when the work is cancelled"
func TestWalkingSkeleton_EditAndDelete(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("Migrate database schema", "todo"))
	dsl.When(ctx, dsl.IRunKanbanEditTitle(ctx.LastTaskID(), "Migrate user table schema"))
	dsl.Then(ctx, dsl.OutputContains("title"))
	dsl.Then(ctx, dsl.BoardShowsTaskUnder("Migrate user table schema", "To Do"))
	dsl.When(ctx, dsl.IRunKanbanDelete(ctx.LastTaskID(), "y"))
	dsl.Then(ctx, dsl.OutputContains("git commit"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.BoardNotListsTask("Migrate user table schema"))
}