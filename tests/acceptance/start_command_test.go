package acceptance

import (
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

func TestStartCommand_TodoTransitions(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Fix OAuth login bug", "todo", "TASK-001"))
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-001"))
	dsl.Then(ctx, dsl.TaskHasStatus("TASK-001", "in-progress"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.OutputContains("Started TASK-001"))
}

func TestStartCommand_AlreadyInProgress(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("API rate limiting", "in-progress", "TASK-002"))
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-002"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.OutputContains("already in progress"))
	dsl.Then(ctx, dsl.TaskStatusRemains("TASK-002", "in-progress"))
}

func TestStartCommand_DoneTaskRejected(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Deploy to staging", "done", "TASK-003"))
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-003"))
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.Then(ctx, dsl.OutputContains("already finished"))
}

func TestStartCommand_UnknownTaskReportsNotFound(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-099"))
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.Then(ctx, dsl.OutputContains("not found"))
}

func TestStartCommand_NoKanbanSetup(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup())
	dsl.When(ctx, dsl.IRunKanbanStart("TASK-001"))
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.Then(ctx, dsl.OutputContains("kanban not initialised"))
}
