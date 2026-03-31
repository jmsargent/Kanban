package acceptance

// US-BSG-03: kanban board --me
//
// All scenarios are marked t.Skip("pending: US-BSG-03 not yet implemented").
// Enable one test, implement, pass, commit. Repeat.
//
// Port-to-port: all scenarios invoke the kanban binary as subprocess.

import (
	"testing"

	dsl "github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// TestBoardMe_ShowsOnlyCurrentDeveloperTasks validates AC-03-1 and AC-03-2:
// "kanban board --me" shows only tasks assigned to the current developer's
// git author email and hides tasks assigned to other developers. This lets
// a developer on a shared board focus on their own work without noise.
func TestBoardMe_ShowsOnlyCurrentDeveloperTasks(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// The test git identity is test@example.com (set by InAGitRepo).
	dsl.Given(ctx, dsl.TaskAssignedTo("Fix OAuth login bug", "test@example.com"))
	myTaskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskAssignedTo("Refactor billing module", "colleague@example.com"))
	otherTaskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardWithMe())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardOutputContains(myTaskID))
	dsl.And(ctx, dsl.BoardOutputDoesNotContain(otherTaskID))
}

// TestBoardMe_WarnsAboutUnassignedTasks validates AC-03-3 and AC-03-4:
// when tasks exist with no assignee "kanban board --me" displays a warning
// informing the developer that unassigned tasks exist but are excluded from
// the filtered view — so no work is silently hidden.
func TestBoardMe_WarnsAboutUnassignedTasks(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskAssignedTo("My assigned task", "test@example.com"))
	// Create an unassigned task (no --assignee flag, relying on kanban add default).
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Unassigned team task"))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardWithMe())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	// The warning should mention unassigned tasks exist.
	dsl.And(ctx, dsl.OutputContains("text: unassigned"))
}

// TestBoardMe_ShowsEmptyBoardGracefully_WhenNoMatchingTasks validates AC-03-5:
// when no tasks are assigned to the current developer "kanban board --me"
// exits 0 and produces a clean empty-board message rather than an error or
// a blank screen.
func TestBoardMe_ShowsEmptyBoardGracefully_WhenNoMatchingTasks(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// All tasks assigned to someone else.
	dsl.Given(ctx, dsl.TaskAssignedTo("Refactor billing module", "other@example.com"))
	dsl.Given(ctx, dsl.TaskAssignedTo("Update CI pipeline", "other@example.com"))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoardWithMe())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	// Empty board message — exact wording determined by implementation.
	dsl.And(ctx, dsl.OutputContains("text: no tasks"))
}

// TestBoardMe_DoesNotAffectUnfilteredBoard validates AC-03-6:
// "kanban board" (without --me) continues to show all tasks regardless of
// assignee. The --me flag only affects its own invocation.
func TestBoardMe_DoesNotAffectUnfilteredBoard(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskAssignedTo("Fix OAuth login bug", "test@example.com"))
	myTaskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskAssignedTo("Refactor billing module", "colleague@example.com"))
	otherTaskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardOutputContains(myTaskID))
	dsl.And(ctx, dsl.BoardOutputContains(otherTaskID))
}
