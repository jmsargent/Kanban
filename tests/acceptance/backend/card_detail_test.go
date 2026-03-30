package backend

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"

// TestCardDetail_ShowsFullInfo verifies that the card detail page renders all
// populated task fields: title, description, priority, assignee, created_by,
// and status.
func TestCardDetail_ShowsFullInfo(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Fix login bug",
		"status: in-progress",
		"assignee: alice",
		"description: The login button fails on mobile devices.",
		"priority: high",
		"created_by: bob",
	))
	When(ctx, IVisitTheBoard())
	When(ctx, IViewCard("title: Fix login bug"))
	Then(ctx, CardShows(
		"title: Fix login bug",
		"status: in-progress",
		"assignee: alice",
		"description: The login button fails on mobile devices.",
		"priority: high",
		"created_by: bob",
	))
}

// TestCardDetail_MissingOptionalFields verifies that optional fields (assignee,
// priority) are not rendered in the HTML when they are empty.
func TestCardDetail_MissingOptionalFields(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Simple task",
		"status: todo",
	))
	When(ctx, IVisitTheBoard())
	When(ctx, IViewCard("title: Simple task"))
	Then(ctx, CardDoesNotShow("field: assignee"))
	Then(ctx, CardDoesNotShow("field: priority"))
}

// TestCardDetail_TaskIDVisible verifies that the task ID is always visible on
// the card detail page.
func TestCardDetail_TaskIDVisible(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Track this task",
		"status: todo",
	))
	When(ctx, IVisitTheBoard())
	When(ctx, IViewCard("title: Track this task"))
	Then(ctx, CardShows("id: TASK-001"))
}
