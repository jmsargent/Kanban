package acceptance

import (
	"testing"

	dsl "github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// Driving port: kanban start <task-id> (CLI command)
// Feature: auto-assign-on-start (US-09, US-09a, US-09b)
//
// Updated for board-state-in-git (US-BSG-02): assignee tracking now goes into
// TransitionEntry.Author only. The task YAML file is no longer modified by
// kanban start. Assertions on task file assignee fields have been removed.

// AC-09-1 — Happy Path
// Starting an unassigned task transitions it to in-progress and exits 0.
func TestAutoAssign_UnassignedTaskGetsAssignedOnStart(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo()) // sets user.name = "Test User"
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Fix login bug", "status: todo", "id: TASK-001"))
	dsl.When(ctx, dsl.IRunKanbanStart("task: TASK-001"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.And(ctx, dsl.StdoutContains("text: Started TASK-001:"))
	dsl.And(ctx, dsl.TaskHasStatus("task: TASK-001", "status: in-progress"))
	// Assignee is now tracked in transitions.log (Author field), not in task YAML.
	dsl.And(ctx, dsl.StdoutDoesNotContain("text: previously assigned"))
}

// AC-09-2
// Starting a task with a different YAML assignee still warns (PreviousAssignee
// is read from whatever name was in the task YAML before the transition).
func TestAutoAssign_PreviouslyAssignedTaskWarnsAndReassigns(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.GitIdentityConfiguredAs("name: Bob"))
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Update docs", "status: todo", "id: TASK-002"))
	dsl.Given(ctx, dsl.ATaskAssigneeSetTo("task: TASK-002", "assignee: Alice"))
	dsl.When(ctx, dsl.IRunKanbanStart("task: TASK-002"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.And(ctx, dsl.StdoutContains("text: Started TASK-002:"))
	dsl.And(ctx, dsl.StdoutContains("text: Note: task was previously assigned to Alice"))
	dsl.And(ctx, dsl.TaskHasStatus("task: TASK-002", "status: in-progress"))
	// Assignee is no longer written to task YAML — it's in TransitionEntry.Author.
}

// AC-09-3
// Starting a task where the YAML assignee matches the git email produces no warning.
func TestAutoAssign_SameAssigneeNoWarning(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo()) // user.email = "test@example.com"
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Refactor cache", "status: todo", "id: TASK-003"))
	// Set the YAML assignee to the git email so the same-assignee check matches.
	dsl.Given(ctx, dsl.ATaskAssigneeSetTo("task: TASK-003", "assignee: test@example.com"))
	dsl.When(ctx, dsl.IRunKanbanStart("task: TASK-003"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.And(ctx, dsl.StdoutContains("text: Started TASK-003:"))
	dsl.And(ctx, dsl.StdoutDoesNotContain("text: previously assigned"))
	dsl.And(ctx, dsl.TaskHasStatus("task: TASK-003", "status: in-progress"))
}

// AC-09-4
// When git identity is not configured, kanban start exits 1 with a clear error
// and leaves the task unchanged.
func TestAutoAssign_MissingIdentityFailsAndLeavesTaskUnchanged(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo()) // user.name = "Test User" initially
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Add tests", "status: todo", "id: TASK-004"))
	dsl.Given(ctx, dsl.GitIdentityUnset()) // strip identity before the When step
	dsl.When(ctx, dsl.IRunKanbanStart("task: TASK-004"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 1"))
	dsl.And(ctx, dsl.StderrContains("text: git identity not configured"))
	dsl.And(ctx, dsl.TaskHasStatus("task: TASK-004", "status: todo"))
	dsl.And(ctx, dsl.TaskHasNoAssignee("task: TASK-004"))
}

// AC-09-5
// Starting an already-in-progress task is idempotent — returns "already in progress".
func TestAutoAssign_AlreadyInProgressTaskPreservesExistingAssignee(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.GitIdentityConfiguredAs("name: Bob"))
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Write tests", "status: in-progress", "id: TASK-005"))
	dsl.Given(ctx, dsl.ATaskAssigneeSetTo("task: TASK-005", "assignee: Alice"))
	dsl.When(ctx, dsl.IRunKanbanStart("task: TASK-005"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.And(ctx, dsl.StdoutContains("text: already in progress"))
	dsl.And(ctx, dsl.TaskAssigneeRemains("task: TASK-005", "assignee: Alice"))
	dsl.And(ctx, dsl.TaskStatusRemains("task: TASK-005", "status: in-progress"))
}
