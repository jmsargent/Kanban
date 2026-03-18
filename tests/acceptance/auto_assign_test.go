package acceptance

import (
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
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
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Fix login bug", "todo", "TASK-001"))
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-001"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.StdoutContains("Started TASK-001:"))
	dsl.And(ctx, dsl.TaskHasStatus("TASK-001", "in-progress"))
	// Assignee is now tracked in transitions.log (Author field), not in task YAML.
	dsl.And(ctx, dsl.StdoutDoesNotContain("previously assigned"))
}

// AC-09-2
// Starting a task with a different YAML assignee still warns (PreviousAssignee
// is read from whatever name was in the task YAML before the transition).
func TestAutoAssign_PreviouslyAssignedTaskWarnsAndReassigns(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.GitIdentityConfiguredAs("Bob"))
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Update docs", "todo", "TASK-002"))
	dsl.Given(ctx, dsl.ATaskAssigneeSetTo("TASK-002", "Alice"))
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-002"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.StdoutContains("Started TASK-002:"))
	dsl.And(ctx, dsl.StdoutContains("Note: task was previously assigned to Alice"))
	dsl.And(ctx, dsl.TaskHasStatus("TASK-002", "in-progress"))
	// Assignee is no longer written to task YAML — it's in TransitionEntry.Author.
}

// AC-09-3
// Starting a task where the YAML assignee matches the git email produces no warning.
func TestAutoAssign_SameAssigneeNoWarning(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo()) // user.email = "test@example.com"
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Refactor cache", "todo", "TASK-003"))
	// Set the YAML assignee to the git email so the same-assignee check matches.
	dsl.Given(ctx, dsl.ATaskAssigneeSetTo("TASK-003", "test@example.com"))
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-003"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.StdoutContains("Started TASK-003:"))
	dsl.And(ctx, dsl.StdoutDoesNotContain("previously assigned"))
	dsl.And(ctx, dsl.TaskHasStatus("TASK-003", "in-progress"))
}

// AC-09-4
// When git identity is not configured, kanban start exits 1 with a clear error
// and leaves the task unchanged.
func TestAutoAssign_MissingIdentityFailsAndLeavesTaskUnchanged(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo()) // user.name = "Test User" initially
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Add tests", "todo", "TASK-004"))
	dsl.Given(ctx, dsl.GitIdentityUnset()) // strip identity before the When step
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-004"))
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.And(ctx, dsl.StderrContains("git identity not configured"))
	dsl.And(ctx, dsl.TaskHasStatus("TASK-004", "todo"))
	dsl.And(ctx, dsl.TaskHasNoAssignee("TASK-004"))
}

// AC-09-5
// Starting an already-in-progress task is idempotent — returns "already in progress".
func TestAutoAssign_AlreadyInProgressTaskPreservesExistingAssignee(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.GitIdentityConfiguredAs("Bob"))
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Write tests", "in-progress", "TASK-005"))
	dsl.Given(ctx, dsl.ATaskAssigneeSetTo("TASK-005", "Alice"))
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-005"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.StdoutContains("already in progress"))
	dsl.And(ctx, dsl.TaskAssigneeRemains("TASK-005", "Alice"))
	dsl.And(ctx, dsl.TaskStatusRemains("TASK-005", "in-progress"))
}
