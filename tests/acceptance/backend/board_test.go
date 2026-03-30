package backend

import (
	"testing"

	"github.com/jmsargent/kanban/tests/acceptance/backend/dsl"
)

// TestViewBoard_ThreeColumnsWithTasks verifies that the board renders three
// columns (Todo, Doing, Done) and places each task in the correct column
// based on its status.
func TestViewBoard_ThreeColumnsWithTasks(t *testing.T) {
	ctx := dsl.NewWebContext(t)
	dsl.Given(ctx, dsl.ARepoWithTasks(
		dsl.Task("Fix login bug", "status: in-progress", "assignee: alice@example.com"),
		dsl.Task("Write docs"),
		dsl.Task("Deploy v1", "status: done"),
	))
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Write docs"))
	dsl.Then(ctx, dsl.ColumnContainsCards("column: Doing", "title: Fix login bug"))
	dsl.Then(ctx, dsl.ColumnContainsCards("column: Done", "title: Deploy v1"))
}
