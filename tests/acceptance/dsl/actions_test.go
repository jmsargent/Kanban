package dsl_test

import (
	"testing"

	"github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// TestActionIRunKanban verifies that IRunKanban("board") exits 0 after init.
func TestActionIRunKanban(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanban("subcommand: board"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}

// TestActionIRunKanbanNew verifies that IRunKanbanNew creates a task and sets lastTaskID.
func TestActionIRunKanbanNew(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNew("title: Test task title"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	if ctx.LastTaskID() == "" {
		t.Fatal("expected LastTaskID to be non-empty after IRunKanbanNew")
	}
}

// TestActionIRunKanbanNewWithOptions verifies that optional flags are accepted.
func TestActionIRunKanbanNewWithOptions(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNewWithOptions("title: Flagged task", "priority: high", "due: 2026-12-31", "assignee: alice"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}

// TestActionIRunKanbanBoard verifies that IRunKanbanBoard exits 0.
func TestActionIRunKanbanBoard(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanBoard())
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}

// TestActionIRunKanbanBoardJSON verifies that IRunKanbanBoardJSON produces valid JSON.
func TestActionIRunKanbanBoardJSON(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanBoardJSON())
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.OutputIsValidJSON())
}

// TestActionIRunKanbanStart verifies that IRunKanbanStart transitions a task.
func TestActionIRunKanbanStart(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Start test task", "status: todo", "id: TASK-001"))
	dsl.When(ctx, dsl.IRunKanbanStart("task: TASK-001"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}

// TestActionIRunKanbanStartOnThatTask verifies that IRunKanbanStartOnThatTask
// resolves the task ID at run time from ctx.lastTaskID.
func TestActionIRunKanbanStartOnThatTask(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("title: Dynamic start task", "status: todo"))
	dsl.When(ctx, dsl.IRunKanbanStartOnThatTask())
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}

// TestActionIRunKanbanDeleteForce verifies that IRunKanbanDeleteForce removes the task.
func TestActionIRunKanbanDeleteForce(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Force delete task", "status: todo", "id: TASK-099"))
	dsl.When(ctx, dsl.IRunKanbanDeleteForce("task: TASK-099"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.TaskFileRemoved("task: TASK-099"))
}

// TestActionIRunKanbanDelete verifies that IRunKanbanDelete pipes confirm input.
func TestActionIRunKanbanDelete(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Delete confirm task", "status: todo", "id: TASK-098"))
	dsl.When(ctx, dsl.IRunKanbanDelete("task: TASK-098", "confirm: y"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.TaskFileRemoved("task: TASK-098"))
}

// TestActionICommitWithMessage verifies that ICommitWithMessage records exit code.
func TestActionICommitWithMessage(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.ICommitWithMessage("message: test: plain commit"))
	dsl.Then(ctx, dsl.GitCommitExitCodeIs("code: 0"))
}

// TestActionICommitWithTaskID verifies that ICommitWithTaskID resolves task ID at run time.
func TestActionICommitWithTaskID(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("title: Commit task", "status: todo"))
	dsl.When(ctx, dsl.ICommitWithTaskID())
	dsl.Then(ctx, dsl.GitCommitExitCodeIs("code: 0"))
}

// TestActionCIStepRunsPass verifies that CIStepRunsPass runs kanban ci-done.
func TestActionCIStepRunsPass(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: CI pass task", "status: in-progress", "id: TASK-010"))
	dsl.Given(ctx, dsl.PipelineCommitWith("TASK-010"))
	dsl.When(ctx, dsl.CIStepRunsPass())
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}

// TestActionCIStepRunsFail verifies that CIStepRunsFail runs ci-done with failure env.
func TestActionCIStepRunsFail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: CI fail task", "status: in-progress", "id: TASK-011"))
	dsl.Given(ctx, dsl.PipelineCommitWith("TASK-011"))
	dsl.When(ctx, dsl.CIStepRunsFail())
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.TaskHasStatus("task: TASK-011", "status: done"))
}

// TestActionIRunKanbanEditTitle verifies that IRunKanbanEditTitle runs kanban edit
// with a mock EDITOR and updates the task title.
func TestActionIRunKanbanEditTitle(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Old title", "status: todo", "id: TASK-020"))
	dsl.When(ctx, dsl.IRunKanbanEditTitle("task: TASK-020", "title: New title"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}
