package backend

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"

// TestAddTask_TitleOnly verifies that a task created with only a title gets sensible
// defaults: status todo and created_by set from the authenticated user's display_name.
func TestAddTask_TitleOnly(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithNoTasks())
	Given(ctx, WithGitHubStub("token: valid-token-123", "login: alice", "display_name: Alice"))
	Given(ctx, AnAuthenticatedUser("token: valid-token-123", "display_name: Alice"))
	When(ctx, IAddTask("title: Fix login bug"))
	When(ctx, IVisitTheBoard())
	Then(ctx, ColumnContainsCards("column: Todo", "title: Fix login bug"))
	Then(ctx, TaskExistsInRepo(
		"title: Fix login bug",
		"status: todo",
		"created_by: Alice",
	))
}

// TestAddTask_TitleRequired verifies that submitting a task with no title fails
// validation: the add-task form is re-rendered with an error and no task file
// is created in the repository.
func TestAddTask_TitleRequired(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithNoTasks())
	Given(ctx, WithGitHubStub("token: valid-token-123", "login: alice", "display_name: Alice"))
	Given(ctx, AnAuthenticatedUser("token: valid-token-123", "display_name: Alice"))
	When(ctx, IAddTask("title: "))
	Then(ctx, TaskCreationFails("error: title is required"))
	Then(ctx, NoNewTaskInRepo())
}

// TestAddTask_AllFields verifies that an authenticated user can add a task with
// all fields (title, description, priority, assignee) via the web form.
// The task appears in the Todo column on the board and its file exists in the repo
// with all submitted fields plus created_by from the session.
func TestAddTask_AllFields(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, ARepoWithNoTasks())
	Given(ctx, WithGitHubStub("token: valid-token-123", "login: alice", "display_name: Alice"))
	Given(ctx, AnAuthenticatedUser("token: valid-token-123", "display_name: Alice"))
	When(ctx, IAddTask(
		"title: Write release notes",
		"description: Describe the v1.0 release",
		"priority: high",
		"assignee: bob",
	))
	When(ctx, IVisitTheBoard())
	Then(ctx, ColumnContainsCards("column: Todo", "title: Write release notes"))
	Then(ctx, TaskExistsInRepo(
		"title: Write release notes",
		"priority: high",
		"assignee: bob",
		"created_by: Alice",
	))
}
