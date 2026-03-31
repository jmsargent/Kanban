package acceptance

import (
	"testing"

	dsl "github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// ---- US-EST-02: kanban ci-done (no commit) ----

// AC-02-1: ci-done identifies task IDs from commit messages and sets status: done.
// AC-02-2: ci-done does not invoke git commit or git add (HEAD SHA unchanged).
func TestCiDone_UpdatesTaskStatusFromCommitMessages(t *testing.T) {
	var sinceSHA, headBefore string
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Fix OAuth login bug", "status: in-progress", "id: TASK-001"))
	dsl.Given(ctx, dsl.CaptureGitHeadSHA(&sinceSHA))
	dsl.Given(ctx, dsl.GitCommitReferencingTask("TASK-001"))
	dsl.Given(ctx, dsl.CaptureGitHeadSHA(&headBefore))
	dsl.When(ctx, dsl.IRunKanbanCiDoneFrom(&sinceSHA))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.TaskFileStatusIs("TASK-001", "done"))
	dsl.Then(ctx, dsl.GitHeadSHAIs(&headBefore))
}

// AC-02-3: when no task IDs appear in the commit range, ci-done exits 0 silently.
func TestCiDone_NoTasksInRangeExitsClean(t *testing.T) {
	var sinceSHA string
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.CaptureGitHeadSHA(&sinceSHA))
	dsl.When(ctx, dsl.IRunKanbanCiDoneFrom(&sinceSHA))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
}

// AC-02-4: ci-done is idempotent — a task already at done is skipped.
func TestCiDone_AlreadyDoneTaskIsSkipped(t *testing.T) {
	var sinceSHA string
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Deploy to staging", "status: done", "id: TASK-001"))
	dsl.Given(ctx, dsl.CaptureGitHeadSHA(&sinceSHA))
	dsl.Given(ctx, dsl.GitCommitReferencingTask("TASK-001"))
	dsl.When(ctx, dsl.IRunKanbanCiDoneFrom(&sinceSHA))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.TaskFileStatusIs("TASK-001", "done"))
}

// ---- US-EST-03: kanban board reads from YAML ----

// AC-03-2: board works without transitions.log present.
// AC-03-3: in-progress YAML status places task in IN PROGRESS column.
func TestBoard_ReadsStatusFromYAMLWithNoTransitionsLog(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Implement login page", "status: in-progress", "id: TASK-001"))
	dsl.Given(ctx, dsl.TransitionsLogAbsent())
	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.BoardShowsTaskInColumn("TASK-001", "In Progress"))
}

// AC-03-1: tasks with different YAML statuses appear in correct columns.
func TestBoard_GroupsTasksByYAMLStatus(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Write docs", "status: todo", "id: TASK-001"))
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Fix bug", "status: in-progress", "id: TASK-002"))
	dsl.Given(ctx, dsl.ATaskWithStatusAs("title: Deploy service", "status: done", "id: TASK-003"))
	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.BoardShowsTaskInColumn("TASK-001", "To Do"))
	dsl.Then(ctx, dsl.BoardShowsTaskInColumn("TASK-002", "In Progress"))
	dsl.Then(ctx, dsl.BoardShowsTaskInColumn("TASK-003", "Done"))
}

// AC-03-4: legacy task with no status: field is treated as todo.
func TestBoard_LegacyTaskWithNoStatusFieldTreatedAsTodo(t *testing.T) {
	t.Skip("not yet implemented: requires DSL step to create task without status: field")
}

// ---- US-EST-04: commit-msg hook removed ----

// AC-04-1: install-hook command does not appear in kanban --help output.
func TestHookRemoved_InstallHookAbsentFromHelp(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanban("subcommand: --help"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.OutputDoesNotContain("install-hook"))
}

// AC-04-2: running kanban install-hook exits 1 with a removal notice.
func TestHookRemoved_InstallHookCommandExitsWithError(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanban("subcommand: install-hook"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 1"))
	dsl.Then(ctx, dsl.OutputContains("text: install-hook has been removed"))
}

// AC-04-3: leftover _hook commit-msg invocations (on developer machines that
// still have the old hook installed) exit 0 with no side effects.
func TestHookRemoved_CommitMsgHookDelegationIsNoOp(t *testing.T) {
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.When(ctx, dsl.IRunKanbanHookCommitMsg("TASK-001: implement feature"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.TransitionsLogAbsent())
}

// ---- Design discovery D7: kanban init no longer auto-commits ----
// (upstream-changes.md AC-init-1; ADR-013 decision 3)

// AC-init-1: kanban init creates .kanban/ but does not invoke git commit or git add.
func TestInit_DoesNotAutoCommit(t *testing.T) {
	var headBefore string
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.NoKanbanSetup())
	dsl.Given(ctx, dsl.CaptureGitHeadSHA(&headBefore))
	dsl.When(ctx, dsl.IRunKanban("subcommand: init"))
	dsl.Then(ctx, dsl.ExitCodeIs("code: 0"))
	dsl.Then(ctx, dsl.KanbanDotKanbanDirectoryExists())
	dsl.Then(ctx, dsl.GitHeadSHAIs(&headBefore))
	dsl.Then(ctx, dsl.InitDidNotAutoCommit())
}
