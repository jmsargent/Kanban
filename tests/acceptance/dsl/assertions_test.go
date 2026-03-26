package dsl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmsargent/kanban/tests/acceptance/dsl"
)


// ─── ExitCodeIs ───────────────────────────────────────────────────────────────

// TestAssertionExitCodeIs_Pass verifies ExitCodeIs(0) passes when lastExit is 0.
func TestAssertionExitCodeIs_Pass(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanBoard())
	dsl.Then(ctx, dsl.ExitCodeIs(0))
}

// TestAssertionExitCodeIs_Fail verifies ExitCodeIs(1) returns an error when lastExit is 0.
func TestAssertionExitCodeIs_Fail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanBoard()) // exits 0
	step := dsl.ExitCodeIs(1)
	if err := step.Run(ctx); err == nil {
		t.Fatal("ExitCodeIs(1) should return an error when lastExit is 0")
	}
}

// ─── OutputContains ───────────────────────────────────────────────────────────

// TestAssertionOutputContains_Pass verifies OutputContains passes when text is in output.
func TestAssertionOutputContains_Pass(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("Search target task", "todo"))
	dsl.When(ctx, dsl.IRunKanbanBoard())
	dsl.Then(ctx, dsl.OutputContains("Search target task"))
}

// TestAssertionOutputContains_Fail verifies OutputContains returns error when text is absent.
func TestAssertionOutputContains_Fail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanBoard())
	step := dsl.OutputContains("NOTPRESENT_XYZ_12345")
	if err := step.Run(ctx); err == nil {
		t.Fatal("OutputContains should return an error when text is absent from output")
	}
}

// ─── StderrContains ───────────────────────────────────────────────────────────

// TestAssertionStderrContains_Fail verifies StderrContains returns error when text is absent.
func TestAssertionStderrContains_Fail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanBoard())
	step := dsl.StderrContains("NOTPRESENT_STDERR_XYZ")
	if err := step.Run(ctx); err == nil {
		t.Fatal("StderrContains should return an error when text is absent from stderr")
	}
}

// ─── TaskHasStatus ────────────────────────────────────────────────────────────

// TestAssertionTaskHasStatus_Pass verifies TaskHasStatus reads the status from the task file.
func TestAssertionTaskHasStatus_Pass(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Status check task", "in-progress", "TASK-050"))
	dsl.Then(ctx, dsl.TaskHasStatus("TASK-050", "in-progress"))
}

// TestAssertionTaskHasStatus_Fail verifies TaskHasStatus returns error on status mismatch.
func TestAssertionTaskHasStatus_Fail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Mismatch status task", "todo", "TASK-051"))
	step := dsl.TaskHasStatus("TASK-051", "done")
	if err := step.Run(ctx); err == nil {
		t.Fatal("TaskHasStatus should return an error when status does not match")
	}
}

// ─── TaskFilePresent / TaskFileRemoved ───────────────────────────────────────

// TestAssertionTaskFilePresent verifies TaskFilePresent passes when the file exists.
func TestAssertionTaskFilePresent(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Present task", "todo", "TASK-060"))
	dsl.Then(ctx, dsl.TaskFilePresent("TASK-060"))
}

// TestAssertionTaskFileRemoved verifies TaskFileRemoved returns error when file exists.
func TestAssertionTaskFileRemoved(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("Still present task", "todo", "TASK-061"))
	step := dsl.TaskFileRemoved("TASK-061")
	if err := step.Run(ctx); err == nil {
		t.Fatal("TaskFileRemoved should return an error when the file still exists")
	}
}

// ─── WorkspaceReady ───────────────────────────────────────────────────────────

// TestAssertionWorkspaceReady verifies that .kanban/tasks/ exists after kanban init.
func TestAssertionWorkspaceReady(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Then(ctx, dsl.WorkspaceReady())
}

// ─── NoANSIEscapeCodes ───────────────────────────────────────────────────────

// TestAssertionNoANSIEscapeCodes_Pass verifies that plain board output has no ANSI codes.
func TestAssertionNoANSIEscapeCodes_Pass(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanBoard())
	dsl.Then(ctx, dsl.NoANSIEscapeCodes())
}

// TestAssertionNoANSIEscapeCodes_Fail verifies that output containing ANSI escape
// sequences causes NoANSIEscapeCodes to return an error.
func TestAssertionNoANSIEscapeCodes_Fail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.WithLastOutput("\033[31mred text\033[0m"))
	step := dsl.NoANSIEscapeCodes()
	if err := step.Run(ctx); err == nil {
		t.Fatal("NoANSIEscapeCodes should return an error when output contains ANSI escape sequences")
	}
}

// ─── NoTempFilesRemain ───────────────────────────────────────────────────────

// TestAssertionNoTempFilesRemain verifies that no .tmp files remain in tasks dir.
func TestAssertionNoTempFilesRemain(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("Atomicity task", "todo"))
	dsl.Then(ctx, dsl.NoTempFilesRemain())
}

// TestAssertionNoTempFilesRemain_Fail verifies an error is returned when .tmp files exist.
func TestAssertionNoTempFilesRemain_Fail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// Manually create a .tmp file in the tasks directory.
	plantTmpStep := dsl.Step{
		Description: "plant a .tmp file in tasks dir",
		Run: func(c *dsl.Context) error {
			tasksDir := filepath.Join(c.RepoDir(), ".kanban", "tasks")
			return os.WriteFile(filepath.Join(tasksDir, "partial.tmp"), []byte("partial"), 0644)
		},
	}
	dsl.Given(ctx, plantTmpStep)
	step := dsl.NoTempFilesRemain()
	if err := step.Run(ctx); err == nil {
		t.Fatal("NoTempFilesRemain should return an error when .tmp files exist")
	}
}

// ─── ConfigFileHasDefaults ───────────────────────────────────────────────────

// TestAssertionConfigFileHasDefaults verifies the .kanban/config has expected defaults.
func TestAssertionConfigFileHasDefaults(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Then(ctx, dsl.ConfigFileHasDefaults())
}

// ─── NoKanbanOutputLines ─────────────────────────────────────────────────────

// TestAssertionNoKanbanOutputLines_Fail verifies that output containing a "kanban:"
// prefix line causes NoKanbanOutputLines to return an error.
func TestAssertionNoKanbanOutputLines_Fail(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.WithLastOutput("kanban: error: something went wrong\n"))
	step := dsl.NoKanbanOutputLines()
	if err := step.Run(ctx); err == nil {
		t.Fatal("NoKanbanOutputLines should return an error when output contains kanban: prefix lines")
	}
}
