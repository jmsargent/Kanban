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

// TestAuth_UserAuthenticatesSuccessfully verifies that a user who submits a
// valid GitHub token and display name receives an encrypted session cookie and
// is redirected to the board with write access (can reach the add-task form).
func TestAuth_UserAuthenticatesSuccessfully(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithNoTasks())
	Given(ctx, WithGitHubStub("token: valid-token-123", "login: alice", "display_name: Alice"))
	When(ctx, IAuthenticate("token: valid-token-123", "display_name: Alice"))
	Then(ctx, IAmOnTheBoard())
	Then(ctx, ICanAddTasks())
}

// TestAuth_InvalidCredentialsRejected verifies that a user who submits an invalid
// GitHub token is rejected with clear feedback and cannot add tasks.
func TestAuth_InvalidCredentialsRejected(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithNoTasks())
	Given(ctx, WithGitHubStub("token: valid-token-123", "login: alice", "display_name: Alice"))
	Given(ctx, AnUnauthorizedUser())
	When(ctx, IAttemptToAuthenticate("token: bad-token-999", "display_name: Hacker"))
	Then(ctx, AuthenticationIsRejected())
	Then(ctx, ICannotAddTasks())
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
