package acceptance

import (
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

// TestAutoTransitions_CommitHookTodoToInProgress ports the non-@skip scenario:
// "First commit referencing a task advances it to in-progress"
func TestAutoTransitions_CommitHookTodoToInProgress(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Fix OAuth login bug", "todo", "TASK-001"))
	dsl.Given(ctx, dsl.CommitHookInstalled())
	dsl.When(ctx, dsl.ICommitWithMessage("TASK-001: reproduce OAuth bug on Chrome"))
	dsl.Then(ctx, dsl.TaskHasStatus("TASK-001", "in-progress"))
	dsl.Then(ctx, dsl.OutputContains("kanban: TASK-001 moved  todo -> in-progress"))
	dsl.Then(ctx, dsl.BoardShowsTaskUnder("TASK-001", "IN PROGRESS"))
	dsl.Then(ctx, dsl.GitCommitExitCodeIs(0))
}

// TestAutoTransitions_CIStepInProgressToDone ports the non-@skip scenario:
// "CI step advances an in-progress task to done when all tests pass"
func TestAutoTransitions_CIStepInProgressToDone(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Implement throttle middleware", "in-progress", "TASK-003"))
	dsl.Given(ctx, dsl.PipelineCommitWith("TASK-003"))
	dsl.When(ctx, dsl.CIStepRunsPass())
	dsl.Then(ctx, dsl.TaskHasStatus("TASK-003", "done"))
	dsl.Then(ctx, dsl.OutputContains("[kanban] TASK-003 moved  in-progress -> done"))
	dsl.Then(ctx, dsl.UpdatedTaskCommitted())
	dsl.Then(ctx, dsl.ExitCodeIs(0))
}
