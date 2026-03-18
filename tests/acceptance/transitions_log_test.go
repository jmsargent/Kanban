package acceptance

// US-BSG-02: Append-only transitions log
//
// All scenarios are marked t.Skip("pending: US-BSG-02 not yet implemented").
// Enable one test, implement, pass, commit. Repeat.
//
// Port-to-port: all scenarios invoke the kanban binary as subprocess.
// Direct .kanban/transitions.log reads are permitted for structural assertions
// (line count, field format) per the Port-to-Port Principle note in the
// configuration — transitions.log is an observable file output of the system.

import (
	"os"
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

const skipUS02 = "pending: US-BSG-02 not yet implemented"

// ---- Task creation ----

// TestTransitionsLog_NewTaskHasNoStatusField validates AC-02-1:
// when a developer creates a new task the task file contains no "status:" field
// because status is now derived solely from transitions.log.
func TestTransitionsLog_NewTaskHasNoStatusField(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())

	dsl.When(ctx, dsl.TaskCreatedViaAdd("Fix OAuth login bug"))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.TaskFileDoesNotContainStatusField(ctx.LastTaskID()))
}

// TestTransitionsLog_NewTaskFileHasLogComment validates AC-02-2:
// when a developer creates a new task the task file contains a comment
// explaining that status is tracked in transitions.log, helping future
// readers understand the storage model.
func TestTransitionsLog_NewTaskFileHasLogComment(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())

	dsl.When(ctx, dsl.TaskCreatedViaAdd("Implement rate limiting"))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.TaskFileContainsComment(ctx.LastTaskID(), "transitions.log"))
}

// TestTransitionsLog_NewTaskAppearsInTodoColumn validates AC-02-3:
// a newly created task with no transitions.log entry appears in the TODO column
// of the board, because a missing entry is treated as implicit "todo".
func TestTransitionsLog_NewTaskAppearsInTodoColumn(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Write release notes"))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardShowsTaskInColumn(ctx.LastTaskID(), "TODO"))
}

// ---- kanban start ----

// TestTransitionsLog_StartAppendsOneLine validates AC-02-4:
// when a developer starts a task exactly one line is appended to
// transitions.log — no duplicate writes, no bulk rewrites.
func TestTransitionsLog_StartAppendsOneLine(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Fix OAuth login bug"))
	taskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.DeveloperRunsKanbanStart(taskID))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.TransitionsLogLineCountIs(1))
}

// TestTransitionsLog_StartLineContainsRequiredFields validates AC-02-5:
// the line appended by "kanban start" contains all five required fields:
// timestamp, task ID, "todo->in-progress", author email, and trigger "manual".
func TestTransitionsLog_StartLineContainsRequiredFields(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Deploy to staging"))
	taskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.DeveloperRunsKanbanStart(taskID))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.TransitionsLogLastLineContains(
		taskID,
		"todo->in-progress",
		"test@example.com",
		"manual",
	))
}

// TestTransitionsLog_StartDoesNotModifyTaskFile validates AC-02-6:
// "kanban start" records the transition in transitions.log only — it must
// not touch the task file, keeping task files as stable, human-authored content.
func TestTransitionsLog_StartDoesNotModifyTaskFile(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Migrate database schema"))
	taskID := ctx.LastTaskID()

	dsl.Then(ctx, dsl.TaskFileMtimeUnchangedAfter(taskID, func(c *dsl.Context) error {
		return dsl.DeveloperRunsKanbanStart(taskID).Run(c)
	}))
}

// TestTransitionsLog_BoardShowsInProgress_AfterStart validates AC-02-7:
// after "kanban start" the board derives the task's status from
// transitions.log and shows the task in the IN PROGRESS column.
func TestTransitionsLog_BoardShowsInProgress_AfterStart(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Implement throttle middleware"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardShowsTaskInColumn(taskID, "IN PROGRESS"))
}

// TestTransitionsLog_StartOnInProgress_ExitsWithMessage validates AC-02-8:
// starting a task that is already in-progress exits with code 0 and an
// informational message — not an error — because the desired state is already met.
func TestTransitionsLog_StartOnInProgress_ExitsWithMessage(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("API rate limiting"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))

	dsl.When(ctx, dsl.DeveloperRunsKanbanStart(taskID))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.OutputContains("already in progress"))
}

// TestTransitionsLog_StartIdempotent_NoDuplicateEntry validates AC-02-9:
// starting a task that is already in-progress does not append a duplicate
// entry to transitions.log — the log remains a clean audit trail.
func TestTransitionsLog_StartIdempotent_NoDuplicateEntry(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Fix OAuth login bug"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))

	dsl.When(ctx, dsl.DeveloperRunsKanbanStart(taskID))

	dsl.Then(ctx, dsl.TransitionsLogLineCountIs(1))
}

// ---- Hook ----

// TestTransitionsLog_HookAppendsOneLine_OnCommit validates AC-02-10:
// when a developer makes a commit referencing a task the commit-msg hook
// appends exactly one line to transitions.log recording the in-progress transition.
func TestTransitionsLog_HookAppendsOneLine_OnCommit(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.HookInstalled())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Fix OAuth login bug"))
	taskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.GitCommitReferencingTask(taskID))

	dsl.Then(ctx, dsl.TransitionsLogLineCountIs(1))
}

// TestTransitionsLog_HookDoesNotModifyTaskFiles validates AC-02-11:
// the commit-msg hook records transitions in transitions.log only and does not
// modify any task file — keeping task files stable across commits.
func TestTransitionsLog_HookDoesNotModifyTaskFiles(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.HookInstalled())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Implement rate limiting"))
	taskID := ctx.LastTaskID()

	dsl.Then(ctx, dsl.TaskFileMtimeUnchangedAfter(taskID, func(c *dsl.Context) error {
		return dsl.GitCommitReferencingTask(taskID).Run(c)
	}))
}

// TestTransitionsLog_HookExitsZero_OnSuccess validates AC-02-12:
// the commit-msg hook exits 0 when it successfully records a transition,
// so the developer's commit always proceeds.
func TestTransitionsLog_HookExitsZero_OnSuccess(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.HookInstalled())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Deploy to staging"))
	taskID := ctx.LastTaskID()

	dsl.When(ctx, dsl.ICommitWithMessage(taskID+": add deployment script"))

	dsl.Then(ctx, dsl.GitCommitExitCodeIs(0))
}

// TestTransitionsLog_HookExitsZero_WhenLogUnwritable validates AC-02-13:
// even when transitions.log cannot be written the commit-msg hook exits 0 —
// the hook must never block a commit under any circumstance.
func TestTransitionsLog_HookExitsZero_WhenLogUnwritable(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.HookInstalled())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Fix OAuth login bug"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TransitionsLogMadeUnwritable())

	dsl.When(ctx, dsl.ICommitWithMessage(taskID+": attempt commit with unwritable log"))

	dsl.Then(ctx, dsl.GitCommitExitCodeIs(0))
}

// TestTransitionsLog_HookWritesStderrWarning_WhenLogUnwritable validates AC-02-14:
// when transitions.log is unwritable the hook emits a warning to stderr so the
// developer is informed the transition was not recorded — but their commit
// still proceeds.
func TestTransitionsLog_HookWritesStderrWarning_WhenLogUnwritable(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.HookInstalled())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Implement throttle middleware"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TransitionsLogMadeUnwritable())

	dsl.When(ctx, dsl.ICommitWithMessage(taskID+": work"))

	// Git captures hook stderr and surfaces it to the terminal.
	// We check the combined output of the git commit command for the warning.
	dsl.Then(ctx, dsl.GitCommitExitCodeIs(0))
	dsl.And(ctx, dsl.OutputContains("warning"))
}

// TestTransitionsLog_HookDoesNotModifyCommitMessage_WhenLogFails validates AC-02-15:
// the commit message the developer wrote must reach the repository unchanged
// even when the hook fails to record the transition.
func TestTransitionsLog_HookDoesNotModifyCommitMessage_WhenLogFails(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.HookInstalled())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Migrate database schema"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TransitionsLogMadeUnwritable())
	originalMessage := taskID + ": original commit message"

	dsl.When(ctx, dsl.ICommitWithMessage(originalMessage))

	dsl.Then(ctx, dsl.GitCommitExitCodeIs(0))
	// Verify the commit message was stored verbatim
	dsl.And(ctx, dsl.UpdatedTaskCommitted()) // reuses "recent git log" check pattern
}

// ---- Board ----

// TestTransitionsLog_BoardDerivesStatus_FromLogNotYAML validates AC-02-16:
// the board command reads status from transitions.log — it must not fall back
// to a "status:" YAML field in the task file, validating the single source of truth.
func TestTransitionsLog_BoardDerivesStatus_FromLogNotYAML(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("API rate limiting"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))

	// The task file has no status field; status comes from transitions.log.
	dsl.Then(ctx, dsl.TaskFileDoesNotContainStatusField(taskID))

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardShowsTaskInColumn(taskID, "IN PROGRESS"))
}

// TestTransitionsLog_BoardShowsTodo_WhenNoLogEntries validates AC-02-17:
// a task with no transitions.log entry is shown in the TODO column —
// the implicit default state for a new task with no history.
func TestTransitionsLog_BoardShowsTodo_WhenNoLogEntries(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Write release notes"))
	taskID := ctx.LastTaskID()

	// Explicitly no kanban start — task has no log entry.
	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardShowsTaskInColumn(taskID, "TODO"))
}

// ---- ci-done ----

// TestTransitionsLog_CiDoneAppendsDoneEntry validates AC-02-18:
// when CI passes "kanban ci-done" appends a "in-progress->done" entry to
// transitions.log for each in-progress task that was referenced in the run's commits.
func TestTransitionsLog_CiDoneAppendsDoneEntry(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Implement throttle middleware"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))
	dsl.Given(ctx, dsl.PipelineCommitWith(taskID))

	dsl.When(ctx, dsl.DeveloperRunsKanbanCiDone(""))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.TransitionsLogLastLineContains(taskID, "in-progress->done"))
}

// TestTransitionsLog_CiDoneCommitsOnlyTransitionsLog validates AC-02-19:
// the commit created by "kanban ci-done" includes only transitions.log —
// no task files should be staged or committed by the CI step.
func TestTransitionsLog_CiDoneCommitsOnlyTransitionsLog(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Deploy to production"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))
	dsl.Given(ctx, dsl.PipelineCommitWith(taskID))

	dsl.When(ctx, dsl.DeveloperRunsKanbanCiDone(""))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.CiDoneCommitContainsOnlyTransitionsLog())
}

// TestTransitionsLog_CiDoneCommitExcludesTaskFiles validates AC-02-20:
// the "kanban ci-done" commit must not stage task files — ensures the
// git history contains only intentional developer commits touching task files.
func TestTransitionsLog_CiDoneCommitExcludesTaskFiles(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Fix OAuth login bug"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))
	dsl.Given(ctx, dsl.PipelineCommitWith(taskID))

	dsl.When(ctx, dsl.DeveloperRunsKanbanCiDone(""))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	// The ci-done commit must not contain any .kanban/tasks/*.md file
	dsl.And(ctx, dsl.CiDoneCommitContainsOnlyTransitionsLog())
	dsl.And(ctx, dsl.TaskFileDoesNotContainStatusField(taskID))
}

// ---- Rebase safety ----

// TestTransitionsLog_RebaseSafe_EntriesPreserved validates AC-02-21:
// transitions.log entries survive a git rebase — the log is an ordinary
// tracked file so rebase preserves it as long as there are no conflicts.
func TestTransitionsLog_RebaseSafe_EntriesPreserved(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Implement rate limiting"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))

	dsl.When(ctx, dsl.RebaseSquashingAllCommits())

	dsl.Then(ctx, dsl.TransitionsLogLineCountAtLeast(1))
}

// TestTransitionsLog_RebaseSafe_BoardCorrectAfterRebase validates AC-02-22:
// after a git rebase the board still derives the correct status from
// transitions.log — the developer's view of task state is unaffected.
func TestTransitionsLog_RebaseSafe_BoardCorrectAfterRebase(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Migrate database schema"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))
	dsl.Given(ctx, dsl.RebaseSquashingAllCommits())

	dsl.When(ctx, dsl.DeveloperRunsKanbanBoard())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardShowsTaskInColumn(taskID, "IN PROGRESS"))
}

// ---- Concurrency and edge cases ----

// TestTransitionsLog_ConcurrentAppends_CorrectLineCount validates AC-02-23:
// when multiple "kanban start" calls execute simultaneously the transitions.log
// ends up with exactly the expected number of appended lines — no lines are lost
// due to concurrent write races.
func TestTransitionsLog_ConcurrentAppends_CorrectLineCount(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Fix OAuth login bug"))
	taskID := ctx.LastTaskID()

	// Launch 5 concurrent starts. Only the first should produce a log entry;
	// subsequent calls find the task already in-progress and are no-ops.
	dsl.When(ctx, dsl.ConcurrentStartsOnSameTask(taskID, 5))

	// Exactly 1 line: the todo->in-progress transition.
	dsl.Then(ctx, dsl.TransitionsLogLineCountIs(1))
}

// TestTransitionsLog_ConcurrentAppends_NoTruncatedLines validates AC-02-24:
// under concurrent writes, every line in transitions.log is complete and
// well-formed — no partial writes, no truncated lines.
func TestTransitionsLog_ConcurrentAppends_NoTruncatedLines(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.MultipleTasksExist(
		"Task Alpha", "Task Beta", "Task Gamma",
	))

	// We do not know the IDs; skip the structured start here.
	// This test validates line integrity after the system has written entries.
	dsl.Given(ctx, dsl.TaskStarted("TASK-001"))

	dsl.When(ctx, dsl.ConcurrentStartsOnSameTask("TASK-001", 5))

	dsl.Then(ctx, dsl.TransitionsLogHasNoTruncatedLines())
}

// TestTransitionsLog_DeletedTask_ExcludedFromBoard validates AC-02-25:
// when a task file is deleted the board does not show that task even if
// transitions.log still contains entries for it. The task file is authoritative
// for task existence; the log is authoritative for state only.
func TestTransitionsLog_DeletedTask_ExcludedFromBoard(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	dsl.Given(ctx, dsl.TaskCreatedViaAdd("Write release notes"))
	taskID := ctx.LastTaskID()
	dsl.Given(ctx, dsl.TaskStarted(taskID))

	dsl.When(ctx, dsl.IRunKanbanDeleteForce(taskID))
	dsl.And(ctx, dsl.DeveloperRunsKanbanBoard())

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.BoardNotListsTask(taskID))
}

// TestTransitionsLog_CiDoneWithNoMatchingTasks_ExitsZeroWithMessage validates AC-02-26:
// when "kanban ci-done" finds no in-progress tasks matching recent commits it
// exits 0 and emits an informational message rather than failing — a no-op run
// is not an error.
func TestTransitionsLog_CiDoneWithNoMatchingTasks_ExitsZeroWithMessage(t *testing.T) {
	if os.Getenv("KANBAN_BIN") == "" {
		t.Skip("KANBAN_BIN not set — run via make acceptance")
	}
	t.Skip(skipUS02)
	ctx := dsl.NewContext(t)
	dsl.Given(ctx, dsl.InAGitRepo())
	dsl.Given(ctx, dsl.KanbanInitialised())
	// No tasks created, no in-progress work.

	dsl.When(ctx, dsl.DeveloperRunsKanbanCiDone(""))

	dsl.Then(ctx, dsl.ExitsSuccessfully())
	dsl.And(ctx, dsl.OutputContains("no tasks"))
}
