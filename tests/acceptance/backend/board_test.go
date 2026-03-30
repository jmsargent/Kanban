package backend

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"


// TestViewBoard_ThreeColumnsWithTasks verifies that the board renders three
// columns (Todo, Doing, Done) and places each task in the correct column
// based on its status.
func TestViewBoard_ThreeColumnsWithTasks(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Fix login bug", "status: in-progress", "assignee: alice@example.com",
		"task: Write docs",
		"task: Deploy v1", "status: done",
	))
	When(ctx, IVisitTheBoard())
	Then(ctx, ColumnContainsCards("column: Todo", "title: Write docs"))
	Then(ctx, ColumnContainsCards("column: Doing", "title: Fix login bug"))
	Then(ctx, ColumnContainsCards("column: Done", "title: Deploy v1"))
}

// TestViewBoard_CardsSortedByDate verifies that cards within a column are
// rendered in ascending creation-date order (oldest first).
func TestViewBoard_CardsSortedByDate(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Newer Task", "status: todo", "created_at: 2024-06-01T00:00:00Z",
		"task: Older Task", "status: todo", "created_at: 2024-01-01T00:00:00Z",
	))
	When(ctx, IVisitTheBoard())
	Then(ctx, CardAppearsBeforeInColumn("first: Older Task", "second: Newer Task", "column: Todo"))
}

// TestViewBoard_EmptyBoard verifies that a board with no tasks still renders
// three empty columns: Todo, Doing, and Done.
func TestViewBoard_EmptyBoard(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithNoTasks())
	When(ctx, IVisitTheBoard())
	Then(ctx, BoardHasColumns("columns: Todo, Doing, Done"))
	Then(ctx, ColumnIsEmpty("column: Todo"))
	Then(ctx, ColumnIsEmpty("column: Doing"))
	Then(ctx, ColumnIsEmpty("column: Done"))
}

// TestViewBoard_ReflectsChangesFromAnotherUser verifies that when another user
// pushes a new task to the remote, a subsequent board view shows that task.
// This exercises the background git pull sync path.
func TestViewBoard_ReflectsChangesFromAnotherUser(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithRemote())
	When(ctx, AUserPushesTask("task: Remote Task", "status: todo"))
	When(ctx, IVisitTheBoard())
	Then(ctx, ColumnContainsCards("column: Todo", "title: Remote Task"))
}

// TestViewBoard_UnauthenticatedUserCanView verifies that a first-time visitor
// with no authentication can view the public board and see its tasks.
func TestViewBoard_UnauthenticatedUserCanView(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Fix login bug", "status: in-progress",
	))
	Given(ctx, AFirstTimeVisitor())
	When(ctx, IVisitTheBoard())
	Then(ctx, ColumnContainsCards("column: Doing", "title: Fix login bug"))
}

// TestViewBoard_CardShowsSummary verifies that clicking a card shows its
// summary information: title, assignee, and status.
func TestViewBoard_CardShowsSummary(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Fix login bug", "status: in-progress", "assignee: alice",
	))
	When(ctx, IVisitTheBoard())
	When(ctx, IViewCard("title: Fix login bug"))
	Then(ctx, CardShows("title: Fix login bug", "assignee: alice", "status: in-progress"))
}
