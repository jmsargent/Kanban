package acceptance

// US-BSG-03: kanban board --me
//
// All scenarios are marked t.Skip("pending: US-BSG-03 not yet implemented").
// Enable one test, implement, pass, commit. Repeat.
//
// Port-to-port: all scenarios invoke the kanban binary as subprocess.

import (
    . "github.com/jmsargent/kanban/tests/acceptance/dsl"

	"testing"

)

// TestBoardMe_ShowsOnlyCurrentDeveloperTasks validates AC-03-1 and AC-03-2:
// "kanban board --me" shows only tasks assigned to the current developer's
// git author email and hides tasks assigned to other developers. This lets
// a developer on a shared board focus on their own work without noise.
func TestBoardMe_ShowsOnlyCurrentDeveloperTasks(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	// The test git identity is test@example.com (set by InAGitRepo).
	Given(ctx, TaskAssignedTo("Fix OAuth login bug", "test@example.com"))
	myTaskID := ctx.LastTaskID()
	Given(ctx, TaskAssignedTo("Refactor billing module", "colleague@example.com"))
	otherTaskID := ctx.LastTaskID()

	When(ctx, DeveloperRunsKanbanBoardWithMe())

	Then(ctx, ExitsSuccessfully())
	And(ctx, BoardOutputContains(myTaskID))
	And(ctx, BoardOutputDoesNotContain(otherTaskID))
}

// TestBoardMe_WarnsAboutUnassignedTasks validates AC-03-3 and AC-03-4:
// when tasks exist with no assignee "kanban board --me" displays a warning
// informing the developer that unassigned tasks exist but are excluded from
// the filtered view — so no work is silently hidden.
func TestBoardMe_WarnsAboutUnassignedTasks(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, TaskAssignedTo("My assigned task", "test@example.com"))
	// Create an unassigned task (no --assignee flag, relying on kanban add default).
	Given(ctx, TaskCreatedViaAdd("Unassigned team task"))

	When(ctx, DeveloperRunsKanbanBoardWithMe())

	Then(ctx, ExitsSuccessfully())
	// The warning should mention unassigned tasks exist.
	And(ctx, OutputContains("text: unassigned"))
}

// TestBoardMe_ShowsEmptyBoardGracefully_WhenNoMatchingTasks validates AC-03-5:
// when no tasks are assigned to the current developer "kanban board --me"
// exits 0 and produces a clean empty-board message rather than an error or
// a blank screen.
func TestBoardMe_ShowsEmptyBoardGracefully_WhenNoMatchingTasks(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	// All tasks assigned to someone else.
	Given(ctx, TaskAssignedTo("Refactor billing module", "other@example.com"))
	Given(ctx, TaskAssignedTo("Update CI pipeline", "other@example.com"))

	When(ctx, DeveloperRunsKanbanBoardWithMe())

	Then(ctx, ExitsSuccessfully())
	// Empty board message — exact wording determined by implementation.
	And(ctx, OutputContains("text: no tasks"))
}

// TestBoardMe_DoesNotAffectUnfilteredBoard validates AC-03-6:
// "kanban board" (without --me) continues to show all tasks regardless of
// assignee. The --me flag only affects its own invocation.
func TestBoardMe_DoesNotAffectUnfilteredBoard(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, TaskAssignedTo("Fix OAuth login bug", "test@example.com"))
	myTaskID := ctx.LastTaskID()
	Given(ctx, TaskAssignedTo("Refactor billing module", "colleague@example.com"))
	otherTaskID := ctx.LastTaskID()

	When(ctx, DeveloperRunsKanbanBoard())

	Then(ctx, ExitsSuccessfully())
	And(ctx, BoardOutputContains(myTaskID))
	And(ctx, BoardOutputContains(otherTaskID))
}
