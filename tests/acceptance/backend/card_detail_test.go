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
