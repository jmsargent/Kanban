package acceptance

import (
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

// TestNewEditorMode_WalkingSkeleton_TaskCreated is the walking skeleton for the
// new-editor-mode feature. It validates the complete user journey: a developer
// invokes "kanban new" with no arguments, an editor opens with a blank task
// template, the developer fills in a title, saves, and sees the standard
// creation confirmation.
//
// This test covers AC-01 (no-argument invocation routes to editor mode) and
// AC-03 (task created and confirmed after valid editor session).
//
// This is the first RED test. It is NOT skipped — it drives the first
// implementation step in the DELIVER wave.
func TestNewEditorMode_WalkingSkeleton_TaskCreated(t *testing.T) {
	ctx := dsl.NewContext(t)

	editorScript, err := dsl.EditorScriptThatSetsTitle(ctx, "Implement user authentication")
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNewInteractive(editorScript))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.SuccessMessageMatchesNewWithTitle("Implement user authentication"))
	dsl.Then(ctx, dsl.HintMessagePresent())
	dsl.And(ctx, dsl.TaskCreatedWithTitle("Implement user authentication"))
	dsl.And(ctx, dsl.NoTempFileFromNewEditor())
}

// TestNewEditorMode_BlankTemplate_StructureCorrect validates that the blank
// task template presented in the editor contains all required fields and the
// "title is required" comment, and omits the due field.
//
// This test covers AC-02.
func TestNewEditorMode_BlankTemplate_StructureCorrect(t *testing.T) {

	ctx := dsl.NewContext(t)

	editorScript, capturePath, err := dsl.EditorScriptThatCapturesTemplate(ctx)
	if err != nil {
		t.Fatalf("setup capture script: %v", err)
	}

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// Editor captures template then leaves title blank → binary exits 2.
	// We do not assert the exit code here — the template structure is what matters.
	dsl.When(ctx, dsl.IRunKanbanNewInteractive(editorScript))
	dsl.Then(ctx, dsl.TemplateHasBlankTitleField(capturePath))
	dsl.Then(ctx, dsl.TemplateHasBlankPriorityField(capturePath))
	dsl.Then(ctx, dsl.TemplateHasBlankAssigneeField(capturePath))
	dsl.Then(ctx, dsl.TemplateHasTitleRequiredComment(capturePath))
	dsl.And(ctx, dsl.TemplateHasNoDueField(capturePath))
}

// TestNewEditorMode_OptionalFields_Persisted validates that when the developer
// sets priority and assignee in the editor, those values appear in the created
// task file.
//
// This test covers AC-04.
func TestNewEditorMode_OptionalFields_Persisted(t *testing.T) {
	t.Skip("pending: new-editor-mode not yet implemented")

	ctx := dsl.NewContext(t)

	editorScript, err := dsl.EditorScriptThatSetsFields(ctx,
		"Add rate limiting to payment endpoint",
		"high",
		"dana@example.com",
	)
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNewInteractive(editorScript))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.TaskCreatedWithFields(
		"Add rate limiting to payment endpoint",
		"high",
		"dana@example.com",
	))
}

// TestNewEditorMode_EmptyTitle_RejectedWithExitCode2 validates that when the
// developer saves the editor without filling in a title, the command exits with
// code 2, prints "title cannot be empty" to stderr, and creates no task file.
//
// This test covers AC-05.
func TestNewEditorMode_EmptyTitle_RejectedWithExitCode2(t *testing.T) {
	ctx := dsl.NewContext(t)

	editorScript, err := dsl.EditorScriptThatLeavesBlankTitle(ctx)
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNewInteractive(editorScript))
	dsl.Then(ctx, dsl.ExitCodeIs(2))
	dsl.Then(ctx, dsl.StderrContains("title cannot be empty"))
	dsl.And(ctx, dsl.NoTaskFileCreated())
	dsl.And(ctx, dsl.NoTempFileFromNewEditor())
}

// TestNewEditorMode_TitleArgument_NoEditorOpened validates that invoking
// "kanban new <title>" with a positional argument preserves the existing
// behaviour: no editor is launched, the task is created immediately, and the
// command exits 0 with the standard success message.
//
// This test covers AC-06.
func TestNewEditorMode_TitleArgument_NoEditorOpened(t *testing.T) {
	t.Skip("pending: new-editor-mode not yet implemented")

	ctx := dsl.NewContext(t)

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNew("Fix pagination bug in task list"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.SuccessMessageMatchesNewWithTitle("Fix pagination bug in task list"))
	dsl.And(ctx, dsl.TaskCreatedWithTitle("Fix pagination bug in task list"))
}

// TestNewEditorMode_EditorUnavailable_ExitsWithRuntimeError validates that when
// neither $EDITOR nor vi is available, the command exits with code 1 and prints
// an error message containing "open editor" to stderr.
//
// This test covers AC-07.
func TestNewEditorMode_EditorUnavailable_ExitsWithRuntimeError(t *testing.T) {
	ctx := dsl.NewContext(t)

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNewInteractiveNoEditor())
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.Then(ctx, dsl.StderrContains("open editor"))
}

// TestNewEditorMode_KanbanNotInitialised_PreflightBlocksEditor validates that
// when kanban has not been initialised in the repository, the command exits
// before opening an editor and prints the standard "not initialised" message.
//
// This test covers AC-08.
func TestNewEditorMode_KanbanNotInitialised_PreflightBlocksEditor(t *testing.T) {
	ctx := dsl.NewContext(t)

	// A script that fails loudly if called — its presence on disk is enough;
	// it should never be invoked because the pre-flight check runs first.
	editorScript, err := dsl.EditorScriptThatSetsTitle(ctx, "should never be created")
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup())
	dsl.When(ctx, dsl.IRunKanbanNewInteractive(editorScript))
	dsl.Then(ctx, dsl.ExitCodeIs(1))
	dsl.Then(ctx, dsl.StderrContains("kanban not initialised"))
	dsl.And(ctx, dsl.NoTaskFileCreated())
}

// TestNewEditorMode_TempFileCleanedUpOnSuccess validates that after a
// successful editor session the temp file created for the blank template is
// removed. The binary uses a deferred os.Remove, so the absence of .tmp files
// in the tasks directory is the observable proxy.
//
// This test covers AC-09 (success path).
func TestNewEditorMode_TempFileCleanedUpOnSuccess(t *testing.T) {
	t.Skip("pending: new-editor-mode not yet implemented")

	ctx := dsl.NewContext(t)

	editorScript, err := dsl.EditorScriptThatSetsTitle(ctx, "Refactor config loading")
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNewInteractive(editorScript))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.And(ctx, dsl.NoTempFileFromNewEditor())
}

// TestNewEditorMode_TempFileCleanedUpOnEmptyTitle validates that when the
// developer saves with a blank title (exit code 2), the temp file is still
// removed. The deferred cleanup must run on all exit paths.
//
// This test covers AC-09 (error path).
func TestNewEditorMode_TempFileCleanedUpOnEmptyTitle(t *testing.T) {
	t.Skip("pending: new-editor-mode not yet implemented")

	ctx := dsl.NewContext(t)

	editorScript, err := dsl.EditorScriptThatLeavesBlankTitle(ctx)
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanNewInteractive(editorScript))
	dsl.Then(ctx, dsl.ExitCodeIs(2))
	dsl.And(ctx, dsl.NoTempFileFromNewEditor())
}
