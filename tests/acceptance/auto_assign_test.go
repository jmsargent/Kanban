package acceptance

import (
    . "github.com/jmsargent/kanban/tests/acceptance/dsl"

	"testing"

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
	ctx := NewContext(t)
	Given(ctx, InAGitRepo()) // sets user.name = "Test User"
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Fix login bug", "status: todo", "id: TASK-001"))
	When(ctx, IRunKanbanStart("task: TASK-001"))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, StdoutContains("text: Started TASK-001:"))
	And(ctx, TaskHasStatus("task: TASK-001", "status: in-progress"))
	// Assignee is now tracked in transitions.log (Author field), not in task YAML.
	And(ctx, StdoutDoesNotContain("text: previously assigned"))
}

// AC-09-2
// Starting a task with a different YAML assignee still warns (PreviousAssignee
// is read from whatever name was in the task YAML before the transition).
func TestAutoAssign_PreviouslyAssignedTaskWarnsAndReassigns(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, GitIdentityConfiguredAs("name: Bob"))
	Given(ctx, ATaskWithStatusAs("title: Update docs", "status: todo", "id: TASK-002"))
	Given(ctx, ATaskAssigneeSetTo("task: TASK-002", "assignee: Alice"))
	When(ctx, IRunKanbanStart("task: TASK-002"))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, StdoutContains("text: Started TASK-002:"))
	And(ctx, StdoutContains("text: Note: task was previously assigned to Alice"))
	And(ctx, TaskHasStatus("task: TASK-002", "status: in-progress"))
	// Assignee is no longer written to task YAML — it's in TransitionEntry.Author.
}

// AC-09-3
// Starting a task where the YAML assignee matches the git email produces no warning.
func TestAutoAssign_SameAssigneeNoWarning(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo()) // user.email = "test@example.com"
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Refactor cache", "status: todo", "id: TASK-003"))
	// Set the YAML assignee to the git email so the same-assignee check matches.
	Given(ctx, ATaskAssigneeSetTo("task: TASK-003", "assignee: test@example.com"))
	When(ctx, IRunKanbanStart("task: TASK-003"))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, StdoutContains("text: Started TASK-003:"))
	And(ctx, StdoutDoesNotContain("text: previously assigned"))
	And(ctx, TaskHasStatus("task: TASK-003", "status: in-progress"))
}

// AC-09-4
// When git identity is not configured, kanban start exits 1 with a clear error
// and leaves the task unchanged.
func TestAutoAssign_MissingIdentityFailsAndLeavesTaskUnchanged(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo()) // user.name = "Test User" initially
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Add tests", "status: todo", "id: TASK-004"))
	Given(ctx, GitIdentityUnset()) // strip identity before the When step
	When(ctx, IRunKanbanStart("task: TASK-004"))
	Then(ctx, ExitCodeIs("code: 1"))
	And(ctx, StderrContains("text: git identity not configured"))
	And(ctx, TaskHasStatus("task: TASK-004", "status: todo"))
	And(ctx, TaskHasNoAssignee("task: TASK-004"))
}

// AC-09-5
// Starting an already-in-progress task is idempotent — returns "already in progress".
func TestAutoAssign_AlreadyInProgressTaskPreservesExistingAssignee(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, GitIdentityConfiguredAs("name: Bob"))
	Given(ctx, ATaskWithStatusAs("title: Write tests", "status: in-progress", "id: TASK-005"))
	Given(ctx, ATaskAssigneeSetTo("task: TASK-005", "assignee: Alice"))
	When(ctx, IRunKanbanStart("task: TASK-005"))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, StdoutContains("text: already in progress"))
	And(ctx, TaskAssigneeRemains("task: TASK-005", "assignee: Alice"))
	And(ctx, TaskStatusRemains("task: TASK-005", "status: in-progress"))
}
