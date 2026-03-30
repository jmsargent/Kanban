package backend

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"

// TestAuth_FirstTimeUserPromptedOnAddTask verifies that an unauthenticated user
// who attempts to navigate to the add-task page is redirected to the token
// entry form rather than seeing the task creation UI.
func TestAuth_FirstTimeUserPromptedOnAddTask(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithNoTasks())
	Given(ctx, AFirstTimeVisitor())
	When(ctx, IVisitTheBoard())
	When(ctx, ITryToAddTask())
	Then(ctx, PromptedToAuthenticate())
}

// TestAuth_UnauthenticatedCanViewNotAdd verifies that an unauthenticated user
// can view the board (read-only access is public) but sees an authentication
// prompt when attempting to add a task.
func TestAuth_UnauthenticatedCanViewNotAdd(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithTasks(
		"task: Fix login bug", "status: in-progress",
	))
	Given(ctx, AFirstTimeVisitor())
	When(ctx, IVisitTheBoard())
	Then(ctx, BoardIsVisible())
	Then(ctx, AddTaskOptionIsVisible())
	When(ctx, ITryToAddTask())
	Then(ctx, PromptedToAuthenticate())
}
