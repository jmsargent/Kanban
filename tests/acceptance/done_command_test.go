package acceptance

import (
    . "github.com/jmsargent/kanban/tests/acceptance/dsl"

	"testing"

)

// US-EST-01: kanban done

// AC-01-1, AC-01-2: in-progress task transitions to done; stdout confirms the move.
func TestDoneCommand_InProgressTaskTransitionsToDone(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Fix OAuth login bug", "status: in-progress", "id: TASK-001"))
	When(ctx, IRunKanbanDone("TASK-001"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, OutputContains("text: moved in-progress -> done"))
	Then(ctx, TaskFileStatusIs("TASK-001", "done"))
}

// AC-01-3: todo task can also be moved directly to done.
func TestDoneCommand_TodoTaskTransitionsToDone(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Refactor auth module", "status: todo", "id: TASK-002"))
	When(ctx, IRunKanbanDone("TASK-002"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, OutputContains("text: moved todo -> done"))
	Then(ctx, TaskFileStatusIs("TASK-002", "done"))
}

// AC-01-4: non-existent task ID exits 1 with a not-found error; no files modified.
func TestDoneCommand_NonexistentTask(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanDone("TASK-999"))
	Then(ctx, ExitCodeIs("code: 1"))
	Then(ctx, OutputContains("text: not found"))
}

// AC-01-6: calling done on an already-done task is idempotent — exits 0.
func TestDoneCommand_AlreadyDoneIsIdempotent(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Deploy to production", "status: done", "id: TASK-003"))
	When(ctx, IRunKanbanDone("TASK-003"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, OutputContains("text: already done"))
	Then(ctx, TaskFileStatusIs("TASK-003", "done"))
}

// AC-01-5: kanban done must not invoke git commit, git add, or any git subprocess.
// Verified by asserting HEAD SHA is unchanged after the command.
func TestDoneCommand_NoAutoCommit(t *testing.T) {
	var headBefore string
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Fix auth bug", "status: in-progress", "id: TASK-001"))
	Given(ctx, CaptureGitHeadSHA(&headBefore))
	When(ctx, IRunKanbanDone("TASK-001"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, GitHeadSHAIs(&headBefore))
}
