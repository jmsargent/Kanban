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

// TestViewBoard_CardsSortedByDate verifies that cards within a column are
// rendered in ascending creation-date order (oldest first).
func TestViewBoard_CardsSortedByDate(t *testing.T) {
	ctx := dsl.NewWebContext(t)
	dsl.Given(ctx, dsl.ARepoWithTasks(
		"task: Newer Task", "status: todo", "created_at: 2024-06-01T00:00:00Z",
		"task: Older Task", "status: todo", "created_at: 2024-01-01T00:00:00Z",
	))
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.Then(ctx, dsl.CardAppearsBeforeInColumn("first: Older Task", "second: Newer Task", "column: Todo"))
}

// TestViewBoard_EmptyBoard verifies that a board with no tasks still renders
// three empty columns: Todo, Doing, and Done.
func TestViewBoard_EmptyBoard(t *testing.T) {
	ctx := dsl.NewWebContext(t)
	dsl.Given(ctx, dsl.ARepoWithNoTasks())
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.Then(ctx, dsl.BoardHasColumns("columns: Todo, Doing, Done"))
	dsl.Then(ctx, dsl.ColumnIsEmpty("column: Todo"))
	dsl.Then(ctx, dsl.ColumnIsEmpty("column: Doing"))
	dsl.Then(ctx, dsl.ColumnIsEmpty("column: Done"))
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
