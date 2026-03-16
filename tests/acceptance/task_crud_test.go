package acceptance

import (
	"testing"
)

// US-02: Create Task

func TestTaskCRUD_DeveloperCreatesTaskWithTitleOnly(t *testing.T) {
	t.Skip("not yet implemented: Developer creates a task with title only")
}

func TestTaskCRUD_DeveloperCreatesTaskWithAllOptionalFields(t *testing.T) {
	t.Skip("not yet implemented: Developer creates a task with all optional fields")
}

func TestTaskCRUD_TaskIDsIncrementSequentially(t *testing.T) {
	t.Skip("not yet implemented: Task IDs increment sequentially")
}

func TestTaskCRUD_CreatingTaskFailsWhenNoTitleProvided(t *testing.T) {
	t.Skip("not yet implemented: Creating a task fails when no title is provided")
}

func TestTaskCRUD_CreatingTaskFailsWhenDueDateIsInPast(t *testing.T) {
	t.Skip("not yet implemented: Creating a task fails when the due date is in the past")
}

func TestTaskCRUD_CreatingTaskFailsOutsideGitRepository(t *testing.T) {
	t.Skip("not yet implemented: Creating a task fails outside a git repository")
}

// US-03: View Board

func TestTaskCRUD_DeveloperViewsBoardWithTasksInAllThreeStatuses(t *testing.T) {
	t.Skip("not yet implemented: Developer views the board with tasks in all three statuses")
}

func TestTaskCRUD_BoardShowsDashForMissingPriorityAndDueUnassignedForAssignee(t *testing.T) {
	t.Skip("not yet implemented: Board shows \"--\" for missing priority and due, \"unassigned\" for missing assignee")
}

func TestTaskCRUD_OverdueTaskShowsDistinctIndicatorOnBoard(t *testing.T) {
	t.Skip("not yet implemented: Overdue task shows a distinct indicator on the board")
}

func TestTaskCRUD_EmptyBoardShowsOnboardingMessage(t *testing.T) {
	t.Skip("not yet implemented: Empty board shows onboarding message")
}

func TestTaskCRUD_BoardOutputsValidMachineReadableFormatWhenRequested(t *testing.T) {
	t.Skip("not yet implemented: Board outputs valid machine-readable format when requested")
}

func TestTaskCRUD_BoardSuppressesColourCodesWhenNOCOLORIsSet(t *testing.T) {
	t.Skip("not yet implemented: Board suppresses colour codes when NO_COLOR is set in the environment")
}

func TestTaskCRUD_BoardProducesPlainOutputWhenPiped(t *testing.T) {
	t.Skip("not yet implemented: Board produces plain output when piped to another command")
}

// US-06: Edit Task

func TestTaskCRUD_DeveloperAddsDescriptionToExistingTask(t *testing.T) {
	t.Skip("not yet implemented: Developer adds a description to an existing task")
}

func TestTaskCRUD_EditDisplaysAllCurrentFieldValuesBeforeOpeningEditor(t *testing.T) {
	t.Skip("not yet implemented: Edit displays all current field values before opening the editor")
}

func TestTaskCRUD_EditingTitleUpdatesTheBoardDisplay(t *testing.T) {
	t.Skip("not yet implemented: Editing the title updates the board display")
}

func TestTaskCRUD_EditWithNoChangesMadeReportsNoUpdate(t *testing.T) {
	t.Skip("not yet implemented: Edit with no changes made reports no update")
}

func TestTaskCRUD_EditingNonExistentTaskReportsClearError(t *testing.T) {
	t.Skip("not yet implemented: Editing a non-existent task reports a clear error")
}

// US-07: Delete Task

func TestTaskCRUD_DeveloperDeletesTaskAfterConfirming(t *testing.T) {
	t.Skip("not yet implemented: Developer deletes a task after confirming")
}

func TestTaskCRUD_DeveloperAbortsDeleteByEnteringN(t *testing.T) {
	t.Skip("not yet implemented: Developer aborts a delete by entering \"n\"")
}

func TestTaskCRUD_DeveloperAbortsDeleteByPressingEnterWithoutInput(t *testing.T) {
	t.Skip("not yet implemented: Developer aborts a delete by pressing Enter without input")
}

func TestTaskCRUD_ForceDeleteRemovesTaskWithoutPrompting(t *testing.T) {
	t.Skip("not yet implemented: Force delete removes a task without prompting")
}

func TestTaskCRUD_DeletingNonExistentTaskReportsClearError(t *testing.T) {
	t.Skip("not yet implemented: Deleting a non-existent task reports a clear error")
}

func TestTaskCRUD_KanbanDoesNotAutoCommitAfterTaskDeletion(t *testing.T) {
	t.Skip("not yet implemented: Kanban does not auto-commit after task deletion")
}
