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
		"task: Fix login bug", "status: in-progress", "assignee: alice@example.com",
		"task: Write docs",
		"task: Deploy v1", "status: done",
	))
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Write docs"))
	dsl.Then(ctx, dsl.ColumnContainsCards("column: Doing", "title: Fix login bug"))
	dsl.Then(ctx, dsl.ColumnContainsCards("column: Done", "title: Deploy v1"))
}

// TestViewBoard_CardShowsSummary verifies that clicking a card shows its
// summary information: title, assignee, and status.
func TestViewBoard_CardShowsSummary(t *testing.T) {
	ctx := dsl.NewWebContext(t)
	dsl.Given(ctx, dsl.ARepoWithTasks(
		"task: Fix login bug", "status: in-progress", "assignee: alice",
	))
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.When(ctx, dsl.IViewCard("title: Fix login bug"))
	dsl.Then(ctx, dsl.CardShows("title: Fix login bug", "assignee: alice", "status: in-progress"))
}
