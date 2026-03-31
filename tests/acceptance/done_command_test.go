package acceptance

import (
	"testing"

	dsl "github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// US-EST-01: kanban done

// AC-01-1, AC-01-2: in-progress task transitions to done; stdout confirms the move.
func TestDoneCommand_InProgressTaskTransitionsToDone(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Fix OAuth login bug", "status: in-progress", "id: TASK-001"))
	dsl.When(ctx, dsl.IRunKanbanDone("TASK-001"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.OutputContains("text: moved in-progress -> done"))
	dsl.Then(ctx, dsl.TaskFileStatusIs("TASK-001", "done"))
}

// AC-01-3: todo task can also be moved directly to done.
func TestDoneCommand_TodoTaskTransitionsToDone(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Refactor auth module", "status: todo", "id: TASK-002"))
	dsl.When(ctx, dsl.IRunKanbanDone("TASK-002"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.OutputContains("text: moved todo -> done"))
	dsl.Then(ctx, dsl.TaskFileStatusIs("TASK-002", "done"))
}

// AC-01-4: non-existent task ID exits 1 with a not-found error; no files modified.
func TestDoneCommand_NonexistentTask(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanDone("TASK-999"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 1"))
	dsl.Then(ctx, dsl.OutputContains("text: not found"))
}

// AC-01-6: calling done on an already-done task is idempotent — exits 0.
func TestDoneCommand_AlreadyDoneIsIdempotent(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Deploy to production", "status: done", "id: TASK-003"))
	dsl.When(ctx, dsl.IRunKanbanDone("TASK-003"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.OutputContains("text: already done"))
	dsl.Then(ctx, dsl.TaskFileStatusIs("TASK-003", "done"))
}

// AC-01-5: kanban done must not invoke git commit, git add, or any git subprocess.
// Verified by asserting HEAD SHA is unchanged after the command.
func TestDoneCommand_NoAutoCommit(t *testing.T) {
	var headBefore string
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Fix auth bug", "status: in-progress", "id: TASK-001"))
	dsl.Given(ctx, dsl.CaptureGitHeadSHA(&headBefore))
	dsl.When(ctx, dsl.IRunKanbanDone("TASK-001"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.GitHeadSHAIs(&headBefore))
}
