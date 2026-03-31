package acceptance

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/dsl"

func TestStartCommand_TodoTransitions(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Fix OAuth login bug", "status: todo", "id: TASK-001"))
	When(ctx, IRunKanbanStart("task: TASK-001"))
	Then(ctx, TaskHasStatus("task: TASK-001", "status: in-progress"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, OutputContains("text: Started TASK-001"))
}

func TestStartCommand_AlreadyInProgress(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: API rate limiting", "status: in-progress", "id: TASK-002"))
	When(ctx, IRunKanbanStart("task: TASK-002"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, OutputContains("text: already in progress"))
	Then(ctx, TaskStatusRemains("task: TASK-002", "status: in-progress"))
}

func TestStartCommand_DoneTaskRejected(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Deploy to staging", "status: done", "id: TASK-003"))
	When(ctx, IRunKanbanStart("task: TASK-003"))
	Then(ctx, ExitCodeIs("code: 1"))
	Then(ctx, OutputContains("text: already finished"))
}

func TestStartCommand_UnknownTaskReportsNotFound(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanStart("task: TASK-099"))
	Then(ctx, ExitCodeIs("code: 1"))
	Then(ctx, OutputContains("text: not found"))
}

func TestStartCommand_NoKanbanSetup(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, NoKanbanSetup())
	When(ctx, IRunKanbanStart("task: TASK-001"))
	Then(ctx, ExitCodeIs("code: 1"))
	Then(ctx, OutputContains("text: kanban not initialised"))
}