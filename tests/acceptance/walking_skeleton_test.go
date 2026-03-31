package acceptance

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/dsl"

// TestWalkingSkeleton_EditAndDelete ports the non-@skip scenario:
// "Developer edits a task and then removes it when the work is cancelled"
func TestWalkingSkeleton_EditAndDelete(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatus("title: Migrate database schema", "status: todo"))
	When(ctx, IRunKanbanEditTitle("task: "+ctx.LastTaskID(), "title: Migrate user table schema"))
	Then(ctx, OutputContains("text: title"))
	Then(ctx, BoardShowsTaskUnder("title: Migrate user table schema", "column: To Do"))
	When(ctx, IRunKanbanDelete("task: "+ctx.LastTaskID(), "confirm: y"))
	Then(ctx, OutputContains("text: git commit"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, BoardNotListsTask("title: Migrate user table schema"))
}