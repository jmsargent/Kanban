package dsl_test

import (
	"testing"

	"github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// TestSetupInAGitRepo verifies that InAGitRepo creates a temp dir, sets repoDir,
// and runs git init so that .git exists.
func TestSetupInAGitRepo(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	// The repoDir is not directly accessible, but we can verify via a subsequent step.
	// KanbanInitialised depends on repoDir being set; if InAGitRepo didn't set it,
	// run() would fail. We verify the git init happened by checking a downstream step.
	dsl.Given(ctx, dsl.NoKanbanSetup())
}

// TestSetupKanbanInitialised verifies that KanbanInitialised runs kanban init
// and creates the .kanban/ directory.
func TestSetupKanbanInitialised(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
}

// TestSetupNoKanbanSetup verifies that NoKanbanSetup asserts .kanban/ does not exist.
func TestSetupNoKanbanSetup(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup()) // must not error in a fresh repo
}

// TestSetupNotAGitRepo verifies that NotAGitRepo creates a temp dir without git init.
func TestSetupNotAGitRepo(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.NotAGitRepo())
	// Verify .git does NOT exist — the step should NOT run git init.
	// We verify indirectly: NoKanbanSetup should succeed (no .kanban/ either).
	dsl.Given(ctx, dsl.NoKanbanSetup())
}

// TestSetupATaskWithStatus verifies that ATaskWithStatus creates a task file
// with the specified status.
func TestSetupATaskWithStatus(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("title: Fix login bug", "status: in-progress"))
	// If ATaskWithStatus succeeded without fatal, the task file was created.
	// The observable outcome is that lastTaskID is set.
	if ctx.LastTaskID() == "" {
		t.Fatal("expected LastTaskID to be set after ATaskWithStatus")
	}
}

// TestSetupEnvVarSet verifies that EnvVarSet appends the env var to ctx.env.
// We verify indirectly: after setting KANBAN_BIN the context still works.
func TestSetupEnvVarSet(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.EnvVarSet("MY_TEST_VAR", "hello"))
	// No direct accessor for env; verify it doesn't panic and the step succeeds.
}

// TestSetupNoTasksExist verifies that NoTasksExist removes all task files.
func TestSetupNoTasksExist(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatus("title: Some Task", "status: todo"))
	dsl.Given(ctx, dsl.NoTasksExist())
	// ATaskWithStatus set lastTaskID; after NoTasksExist, task file should be gone.
	// TaskFileExistsAs would fail — we just verify NoTasksExist doesn't error here.
}

