package acceptance

import (
    . "github.com/jmsargent/kanban/tests/acceptance/dsl"

	"testing"

)

// ---- US-EST-02: kanban ci-done (no commit) ----

// AC-02-1: ci-done identifies task IDs from commit messages and sets status: done.
// AC-02-2: ci-done does not invoke git commit or git add (HEAD SHA unchanged).
func TestCiDone_UpdatesTaskStatusFromCommitMessages(t *testing.T) {
	var sinceSHA, headBefore string
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Fix OAuth login bug", "status: in-progress", "id: TASK-001"))
	Given(ctx, CaptureGitHeadSHA(&sinceSHA))
	Given(ctx, GitCommitReferencingTask("TASK-001"))
	Given(ctx, CaptureGitHeadSHA(&headBefore))
	When(ctx, IRunKanbanCiDoneFrom(&sinceSHA))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, TaskFileStatusIs("TASK-001", "done"))
	Then(ctx, GitHeadSHAIs(&headBefore))
}

// AC-02-3: when no task IDs appear in the commit range, ci-done exits 0 silently.
func TestCiDone_NoTasksInRangeExitsClean(t *testing.T) {
	var sinceSHA string
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, CaptureGitHeadSHA(&sinceSHA))
	When(ctx, IRunKanbanCiDoneFrom(&sinceSHA))
	Then(ctx, ExitCodeIs("code: 0"))
}

// AC-02-4: ci-done is idempotent — a task already at done is skipped.
func TestCiDone_AlreadyDoneTaskIsSkipped(t *testing.T) {
	var sinceSHA string
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Deploy to staging", "status: done", "id: TASK-001"))
	Given(ctx, CaptureGitHeadSHA(&sinceSHA))
	Given(ctx, GitCommitReferencingTask("TASK-001"))
	When(ctx, IRunKanbanCiDoneFrom(&sinceSHA))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, TaskFileStatusIs("TASK-001", "done"))
}

// ---- US-EST-03: kanban board reads from YAML ----

// AC-03-2: board works without transitions.log present.
// AC-03-3: in-progress YAML status places task in IN PROGRESS column.
func TestBoard_ReadsStatusFromYAMLWithNoTransitionsLog(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Implement login page", "status: in-progress", "id: TASK-001"))
	Given(ctx, TransitionsLogAbsent())
	When(ctx, DeveloperRunsKanbanBoard())
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, BoardShowsTaskInColumn("TASK-001", "In Progress"))
}

// AC-03-1: tasks with different YAML statuses appear in correct columns.
func TestBoard_GroupsTasksByYAMLStatus(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	Given(ctx, ATaskWithStatusAs("title: Write docs", "status: todo", "id: TASK-001"))
	Given(ctx, ATaskWithStatusAs("title: Fix bug", "status: in-progress", "id: TASK-002"))
	Given(ctx, ATaskWithStatusAs("title: Deploy service", "status: done", "id: TASK-003"))
	When(ctx, DeveloperRunsKanbanBoard())
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, BoardShowsTaskInColumn("TASK-001", "To Do"))
	Then(ctx, BoardShowsTaskInColumn("TASK-002", "In Progress"))
	Then(ctx, BoardShowsTaskInColumn("TASK-003", "Done"))
}

// AC-03-4: legacy task with no status: field is treated as todo.
func TestBoard_LegacyTaskWithNoStatusFieldTreatedAsTodo(t *testing.T) {
	t.Skip("not yet implemented: requires DSL step to create task without status: field")
}

// ---- US-EST-04: commit-msg hook removed ----

// AC-04-1: install-hook command does not appear in kanban --help output.
func TestHookRemoved_InstallHookAbsentFromHelp(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanban("subcommand: --help"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, OutputDoesNotContain("install-hook"))
}

// AC-04-2: running kanban install-hook exits 1 with a removal notice.
func TestHookRemoved_InstallHookCommandExitsWithError(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanban("subcommand: install-hook"))
	Then(ctx, ExitCodeIs("code: 1"))
	Then(ctx, OutputContains("text: install-hook has been removed"))
}

// AC-04-3: leftover _hook commit-msg invocations (on developer machines that
// still have the old hook installed) exit 0 with no side effects.
func TestHookRemoved_CommitMsgHookDelegationIsNoOp(t *testing.T) {
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, KanbanInitialised())
	When(ctx, IRunKanbanHookCommitMsg("TASK-001: implement feature"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, TransitionsLogAbsent())
}

// ---- Design discovery D7: kanban init no longer auto-commits ----
// (upstream-changes.md AC-init-1; ADR-013 decision 3)

// AC-init-1: kanban init creates .kanban/ but does not invoke git commit or git add.
func TestInit_DoesNotAutoCommit(t *testing.T) {
	var headBefore string
	ctx := NewContext(t)
	Given(ctx, InAGitRepo())
	Given(ctx, NoKanbanSetup())
	Given(ctx, CaptureGitHeadSHA(&headBefore))
	When(ctx, IRunKanban("subcommand: init"))
	Then(ctx, ExitCodeIs("code: 0"))
	Then(ctx, KanbanDotKanbanDirectoryExists())
	Then(ctx, GitHeadSHAIs(&headBefore))
	Then(ctx, InitDidNotAutoCommit())
}
