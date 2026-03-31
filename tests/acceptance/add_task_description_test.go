package acceptance

// add_task_description_test.go — Acceptance tests for the add-task-description feature.
//
// Feature scope:
//   US-01: Add description: "" to blankTaskTemplate — editor template field affordance
//   US-02: Wire snapshot.Description → task.Description in NewEditorTask.Execute()
//   US-03: Add --description flag to kanban new for non-interactive use
//
// Driving port: kanban binary invoked as subprocess.
// No mocks at this level — real filesystem, real git.
//
// Description is stored as the Markdown body of the task file (text below the
// closing "---" front matter delimiter), NOT as a "description:" YAML key.
//
// Implementation sequence: remove t.Skip from tests one at a time. The walking
// skeleton (TestAddTaskDescription_WalkingSkeleton_DescriptionFieldInTemplate)
// has NO t.Skip — it is the first RED test that drives implementation.

import (
    . "github.com/jmsargent/kanban/tests/acceptance/dsl"

	"testing"

)

// ---------------------------------------------------------------------------
// Walking skeleton — US-01
// ---------------------------------------------------------------------------

// TestAddTaskDescription_WalkingSkeleton_DescriptionFieldInTemplate is the
// walking skeleton for this feature. It validates the most visible user outcome:
// when a developer opens "kanban new" interactively, the editor template already
// contains a description field they can fill in without knowing the internal
// field name.
//
// This is the first RED test. It has NO t.Skip — it drives the first
// implementation step (adding description: "" to blankTaskTemplate).
//
// Covers: AC-01-1, AC-01-2, AC-01-4
func TestAddTaskDescription_WalkingSkeleton_DescriptionFieldInTemplate(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	editorScript, capturePath, err := EditorScriptThatCapturesTemplate(ctx)
	if err != nil {
		t.Fatalf("setup capture script: %v", err)
	}

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	// Editor captures the template then leaves title blank — binary exits 2.
	// We assert only on template structure, not exit code.
	When(ctx, IRunKanbanNewInteractive(editorScript))
	Then(ctx, TemplateHasBlankDescriptionField("path: "+capturePath))
	And(ctx, TemplateHasBlankTitleField(capturePath))
	And(ctx, TemplateHasBlankPriorityField(capturePath))
	And(ctx, TemplateHasBlankAssigneeField(capturePath))
}

// ---------------------------------------------------------------------------
// US-01: Template field
// ---------------------------------------------------------------------------

// TestAddTaskDescription_EmptyDescriptionNoError validates that a developer who
// leaves the description field blank in the editor still sees a successful task
// creation — an empty description is not an error.
//
// Covers: AC-01-3
func TestAddTaskDescription_EmptyDescriptionNoError(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	editorScript, err := EditorScriptThatSetsTitle(ctx, "Add pagination to results")
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewInteractive(editorScript))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, SuccessMessageMatchesNewWithTitle("Add pagination to results"))
}

// TestAddTaskDescription_TemplateFieldAppearsInEditorSession validates that
// the description field appears in the template when a full editor session runs
// (not just when the template is captured with no edits). This confirms the
// field is part of blankTaskTemplate, not injected after editor exit.
//
// Covers: AC-01-4
func TestAddTaskDescription_TemplateFieldAppearsInEditorSession(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	editorScript, capturePath, err := EditorScriptThatCapturesTemplate(ctx)
	if err != nil {
		t.Fatalf("setup capture script: %v", err)
	}

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewInteractive(editorScript))
	// The template must contain description before the editor touches it.
	Then(ctx, TemplateHasBlankDescriptionField("path: "+capturePath))
	And(ctx, TemplateHasNoDueField(capturePath))
}

// ---------------------------------------------------------------------------
// US-02: Editor wiring
// ---------------------------------------------------------------------------

// TestAddTaskDescription_EditorDescription_SavedToTaskBody validates the
// primary wiring: a description typed in the editor is persisted to the task
// file as the Markdown body (below the front matter closing delimiter).
//
// Covers: AC-02-1, AC-02-2
func TestAddTaskDescription_EditorDescription_SavedToTaskBody(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	editorScript, err := EditorScriptThatSetsTitleAndDescription(
		ctx,
		"Fix JWT refresh auth bug",
		"JWT expiry not checked on refresh path",
	)
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewInteractive(editorScript))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, SuccessMessageMatchesNewWithTitle("Fix JWT refresh auth bug"))
	And(ctx, TaskBodyContains("text: JWT expiry not checked on refresh path"))
}

// TestAddTaskDescription_EmptyEditorDescription_NoBodyContent validates that
// when a developer leaves the description field empty in the editor, the task
// file body is empty — no data loss, no error.
//
// Covers: AC-02-3
func TestAddTaskDescription_EmptyEditorDescription_NoBodyContent(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	editorScript, err := EditorScriptThatSetsTitle(ctx, "Add pagination")
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewInteractive(editorScript))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, TaskBodyIsEmpty())
}

// TestAddTaskDescription_EmptyTitle_DescriptionNotPersisted validates that
// title-empty validation still applies even when a description is filled in —
// no task is created, and the description value is not persisted.
//
// Covers: AC-02-4
func TestAddTaskDescription_EmptyTitle_DescriptionNotPersisted(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	// Editor sets description but leaves title blank.
	editorScript, err := EditorScriptThatSetsTitleAndDescription(ctx, `""`, "Some context")
	if err != nil {
		t.Fatalf("setup editor script: %v", err)
	}

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewInteractive(editorScript))
	Then(ctx, ExitCodeIs("code: 2"))
	And(ctx, StderrContains("text: title cannot be empty"))
	And(ctx, NoTaskFileCreated())
}

// ---------------------------------------------------------------------------
// US-03: --description flag
// ---------------------------------------------------------------------------

// TestAddTaskDescription_FlagSavesDescriptionToTaskBody validates that passing
// --description on the command line persists the value to the task file body.
//
// Covers: AC-03-1, AC-03-2
func TestAddTaskDescription_FlagSavesDescriptionToTaskBody(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewWithDescription(
		"title: Deploy hotfix to prod",
		"description: CVE-2025-1234 patch, must ship before 17:00 UTC",
	))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, StdoutContains("text: Created TASK-"))
	And(ctx, StdoutContains("text: Deploy hotfix to prod"))
	And(ctx, TaskBodyContains("text: CVE-2025-1234 patch, must ship before 17:00 UTC"))
}

// TestAddTaskDescription_FlagEmpty_TaskCreatedWithNoBody validates that an
// empty --description flag value is accepted and results in a task with no
// Markdown body content.
//
// Covers: AC-03-4, AC-03-6
func TestAddTaskDescription_FlagEmpty_TaskCreatedWithNoBody(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewWithDescription("Add pagination", ""))
	Then(ctx, ExitCodeIs("code: 0"))
	And(ctx, TaskBodyIsEmpty())
}

// TestAddTaskDescription_FlagWithEmptyTitle_ExitsWithCode2 validates that the
// --description flag does not bypass title validation — an empty title with a
// non-empty description still exits 2.
//
// Covers: AC-03-5
func TestAddTaskDescription_FlagWithEmptyTitle_ExitsWithCode2(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanNewWithDescription("", "Some context"))
	Then(ctx, ExitCodeIs("code: 2"))
	And(ctx, StderrContains("text: title cannot be empty"))
}

// TestAddTaskDescription_FlagAppearsInHelpOutput validates that the new flag
// is discoverable by developers via "kanban new --help".
//
// Covers: AC-03-3
func TestAddTaskDescription_FlagAppearsInHelpOutput(t *testing.T) {
	t.Skip("TODO: implement")

	ctx := NewContext(t)

	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanban("subcommand: new --help"))
	Then(ctx, StdoutContains("text: --description"))
}
