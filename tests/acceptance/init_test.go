package acceptance

import (
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

// @ported
func TestInit_DeveloperSetsUpKanbanInNewGitRepository(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup())
	dsl.When(ctx, dsl.IRunKanban("init"))
	dsl.Then(ctx, dsl.WorkspaceReady())
	dsl.Then(ctx, dsl.ConfigFileHasDefaults())
	dsl.Then(ctx, dsl.HookLogInGitignore())
	dsl.Then(ctx, dsl.OutputContains("Initialised kanban at .kanban/"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
}

func TestInit_PreExistingStagedFilesArePreservedAfterInit(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup())
	dsl.Given(ctx, dsl.WithStagedFile("my-work.go", "package main\n"))
	dsl.When(ctx, dsl.IRunKanban("init"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.StagedFilesContain("my-work.go"))
}

func TestInit_CommitOnlyContainsKanbanConfigFile(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup())
	dsl.Given(ctx, dsl.WithStagedFile("my-work.go", "package main\n"))
	dsl.When(ctx, dsl.IRunKanban("init"))
	dsl.Then(ctx, dsl.ExitCodeIs(0))
	dsl.Then(ctx, dsl.LastCommitChangedOnly(".kanban/config"))
}

func TestInit_RunningInitSecondTimeMakesNoChanges(t *testing.T) {
	t.Skip("not yet implemented: Running init a second time makes no changes")
}

func TestInit_DeveloperCannotInitialiseKanbanOutsideGitRepository(t *testing.T) {
	t.Skip("not yet implemented: Developer cannot initialise kanban outside a git repository")
}
