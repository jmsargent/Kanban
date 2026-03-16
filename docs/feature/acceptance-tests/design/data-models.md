# Data Models: Internal Go BDD DSL

**Feature**: acceptance-tests
**Wave**: DESIGN
**Date**: 2026-03-16

---

## Core Types

### `Step`

```go
// Step is a named test action returned by step factory functions.
// Description is embedded in Fatalf messages when the step fails.
// Run is the function the orchestrator invokes.
type Step struct {
    Description string
    Run         func(*Context) error
}
```

### `Context`

```go
// Context holds mutable state for a single test scenario.
// Created via NewContext(t). Cleaned up automatically via t.Cleanup.
type Context struct {
    t          *testing.T
    repoDir    string  // absolute path to the per-test temp git repo
    binPath    string  // absolute path to the compiled kanban binary
    lastStdout string
    lastStderr string
    lastOutput string  // lastStdout + lastStderr
    lastExit   int
    lastTaskID string  // most recent TASK-NNN captured from output
    env        []string
}
```

### `NewContext`

```go
// NewContext creates a Context for the given test, resolves the binary path,
// and registers temp directory cleanup with t.Cleanup.
func NewContext(t *testing.T) *Context
```

Binary resolution order:
1. `os.Getenv("KANBAN_BIN")` — if non-empty, use as-is.
2. `filepath.Abs("../../bin/kanban")` resolved relative to the test package directory — the same fallback as the existing `resolveDefaultBin()` in `kanban_steps_test.go`.

### `LastTaskID`

```go
// LastTaskID returns the most recently captured TASK-NNN identifier.
// Used by test files to pass a dynamic task ID to subsequent step factories.
func (ctx *Context) LastTaskID() string
```

---

## Orchestrators

```go
func Given(ctx *Context, step Step)
func When(ctx *Context, step Step)
func Then(ctx *Context, step Step)
func And(ctx *Context, step Step)  // identical to Then
```

Failure message format (where `<Phase>` is the orchestrator name):

```
<Phase>: <step.Description>: <err>
```

Example: `Given: a task "Fix OAuth bug" with status "todo": kanban new failed (exit 1): ...`

---

## Step Factory Signatures

All factories return `Step`. Factories that need parameters accept them as arguments and close over them in `Run`.

### Setup Step Factories

These establish preconditions. They typically run `git init`, `kanban init`, create task files, or configure the context.

```go
// InAGitRepo creates a temporary directory, runs git init, sets repoDir,
// registers t.Cleanup(os.RemoveAll). Also creates an initial commit.
func InAGitRepo() Step

// KanbanInitialised runs "kanban init" in the current repoDir.
func KanbanInitialised() Step

// NoKanbanSetup asserts that .kanban/ does not exist in repoDir.
func NoKanbanSetup() Step

// NotAGitRepo creates a temp directory without git init, sets repoDir.
func NotAGitRepo() Step

// ATaskWithStatus creates a task via "kanban new" and optionally overwrites
// the status field in the task file front matter.
func ATaskWithStatus(title, status string) Step

// ATaskWithStatusAs creates a task and renames the file to the given taskID.
func ATaskWithStatusAs(title, status, taskID string) Step

// ATaskExists creates a task in "todo" status.
func ATaskExists(title string) Step

// NoTasksExist removes all files from .kanban/tasks/.
func NoTasksExist() Step

// TaskFileExistsAs asserts that the given task file is present on disk.
func TaskFileExistsAs(taskID string) Step

// CommitHookInstalled ensures the commit-msg hook is present by running "kanban init".
func CommitHookInstalled() Step

// EnvVarSet appends name=value to ctx.env.
func EnvVarSet(name, value string) Step

// PipelineCommitWith stages all changes and creates a commit referencing taskID.
func PipelineCommitWith(taskID string) Step
```

### Action Step Factories

These invoke the kanban binary or git commands and record output in the context.

```go
// IRunKanban runs "kanban <subcommand>" splitting on spaces.
func IRunKanban(subcommand string) Step

// IRunKanbanNew runs "kanban new <title>".
func IRunKanbanNew(title string) Step

// IRunKanbanNewWithOptions runs "kanban new <title>" with optional metadata flags.
func IRunKanbanNewWithOptions(title, priority, due, assignee string) Step

// IRunKanbanBoard runs "kanban board".
func IRunKanbanBoard() Step

// IRunKanbanBoardJSON runs "kanban board --json".
func IRunKanbanBoardJSON() Step

// IRunKanbanStart runs "kanban start <taskID>".
func IRunKanbanStart(taskID string) Step

// IRunKanbanStartOnThatTask runs "kanban start <ctx.LastTaskID()>".
// The task ID is resolved at run time from ctx, not at factory call time.
func IRunKanbanStartOnThatTask() Step

// IRunKanbanEdit runs "kanban edit <taskID>" with a mock EDITOR script
// that appends the given description text.
func IRunKanbanEditAddDescription(taskID, description string) Step

// IRunKanbanEditTitle runs "kanban edit <taskID>" with a mock EDITOR script
// that replaces the title field with newTitle.
func IRunKanbanEditTitle(taskID, newTitle string) Step

// IRunKanbanDelete runs "kanban delete <taskID>" piping confirmInput to stdin.
func IRunKanbanDelete(taskID, confirmInput string) Step

// IRunKanbanDeleteForce runs "kanban delete <taskID> --force".
func IRunKanbanDeleteForce(taskID string) Step

// ICommitWithMessage runs "git commit --allow-empty -m <message>".
func ICommitWithMessage(message string) Step

// ICommitWithTaskID runs "git commit --allow-empty -m <ctx.LastTaskID()>: start working on this".
// Task ID is resolved at run time.
func ICommitWithTaskID() Step

// CIStepRunsPass runs "kanban ci-done".
func CIStepRunsPass() Step

// CIStepRunsFail runs "kanban ci-done" with KANBAN_TEST_EXIT=1 in env.
func CIStepRunsFail() Step
```

### Assertion Step Factories

These read state from `ctx` (captured output, exit code) or from the filesystem and return an error if the expectation is not met.

```go
// ExitCodeIs asserts ctx.lastExit == code.
func ExitCodeIs(code int) Step

// StdoutContains asserts ctx.lastStdout contains text.
func StdoutContains(text string) Step

// OutputContains asserts ctx.lastOutput contains text.
func OutputContains(text string) Step

// OutputIsValidJSON asserts ctx.lastOutput is parseable as JSON.
func OutputIsValidJSON() Step

// JSONHasFields asserts the JSON array in ctx.lastOutput contains an object
// with all named fields present. fields is a comma-separated list.
func JSONHasFields(fields string) Step

// TaskHasStatus reads the task file for taskID and asserts status == expected.
func TaskHasStatus(taskID, expected string) Step

// TaskStatusRemains is an alias for TaskHasStatus with a distinct description.
func TaskStatusRemains(taskID, expected string) Step

// TaskFileExists asserts the file .kanban/tasks/<taskID>.md is present.
func TaskFilePresent(taskID string) Step

// TaskFileRemoved asserts the file .kanban/tasks/<taskID>.md is absent.
func TaskFileRemoved(taskID string) Step

// BoardShowsTaskUnder runs "kanban board" and asserts that title appears
// after the heading line in the output.
func BoardShowsTaskUnder(title, heading string) Step

// BoardNotListsTask runs "kanban board" and asserts taskID does not appear.
func BoardNotListsTask(taskID string) Step

// GitCommitExitCodeIs asserts ctx.lastExit == code after a git commit action.
func GitCommitExitCodeIs(code int) Step

// WorkspaceReady asserts .kanban/tasks/ exists.
func WorkspaceReady() Step

// ConfigFileHasDefaults asserts .kanban/config contains expected defaults.
func ConfigFileHasDefaults() Step

// HookLogInGitignore asserts .gitignore contains "hook.log".
func HookLogInGitignore() Step

// NoTempFilesRemain asserts no *.tmp files exist in .kanban/tasks/.
func NoTempFilesRemain() Step

// UpdatedTaskCommitted asserts git log contains a kanban commit.
func UpdatedTaskCommitted() Step

// NoAutoCommitFromDelete asserts the most recent git log entry does not
// reference "delete" or "remove task".
func NoAutoCommitFromDelete() Step

// NoKanbanOutputLines asserts no line in ctx.lastOutput starts with "kanban:".
func NoKanbanOutputLines() Step

// NoTransitionLines asserts no line in ctx.lastOutput contains "moved" and "->".
func NoTransitionLines() Step

// NoANSIEscapeCodes asserts ctx.lastOutput contains no ANSI escape sequences.
func NoANSIEscapeCodes() Step

// NoSpinnerChars asserts ctx.lastOutput contains no Unicode spinner characters.
func NoSpinnerChars() Step
```

---

## Representative Usage Example

```go
func TestStartTask_TodoTransitions(t *testing.T) {
    ctx := dsl.NewContext(t)
    Given(ctx, dsl.InAGitRepo())
    Given(ctx, dsl.KanbanInitialised())
    Given(ctx, dsl.ATaskWithStatusAs("Fix OAuth bug", "todo", "TASK-001"))
    When(ctx, dsl.IRunKanbanStart("TASK-001"))
    Then(ctx, dsl.StdoutContains("Started"))
    Then(ctx, dsl.ExitCodeIs(0))
    Then(ctx, dsl.TaskHasStatus("TASK-001", "in-progress"))
}

func TestStartTask_AlreadyInProgress(t *testing.T) {
    ctx := dsl.NewContext(t)
    Given(ctx, dsl.InAGitRepo())
    Given(ctx, dsl.KanbanInitialised())
    Given(ctx, dsl.ATaskWithStatusAs("API rate limiting", "in-progress", "TASK-002"))
    When(ctx, dsl.IRunKanbanStart("TASK-002"))
    Then(ctx, dsl.ExitCodeIs(0))
    Then(ctx, dsl.OutputContains("already in progress"))
    Then(ctx, dsl.TaskStatusRemains("TASK-002", "in-progress"))
}
```

---

## Step Factory Coverage — Gherkin Scenario Mapping

The following table maps each distinct Gherkin step pattern from the existing feature files to the DSL factory that replaces it.

| Gherkin step | DSL factory |
|---|---|
| `I am working in a git repository` | `InAGitRepo()` |
| `the repository is initialised with kanban` | `KanbanInitialised()` |
| `the repository has no kanban setup` | `NoKanbanSetup()` |
| `the current directory is not a git repository` | `NotAGitRepo()` |
| `a task "T" exists with status "S"` | `ATaskWithStatus("T", "S")` |
| `a task "T" exists with status "S" as "ID"` | `ATaskWithStatusAs("T", "S", "ID")` |
| `a task "T" exists` | `ATaskExists("T")` |
| `no tasks exist yet` | `NoTasksExist()` |
| `a task file "ID" exists` | `TaskFileExistsAs("ID")` |
| `the git commit hook is installed` | `CommitHookInstalled()` |
| `the environment variable "VAR" is set` | `EnvVarSet("VAR", "1")` |
| `the pipeline run includes a commit with "ID"` | `PipelineCommitWith("ID")` |
| `I run "kanban <subcmd>"` | `IRunKanban("<subcmd>")` |
| `I run "kanban new" with title "T"` | `IRunKanbanNew("T")` |
| `I run "kanban new" with title/priority/due/assignee` | `IRunKanbanNewWithOptions(...)` |
| `I run "kanban board"` | `IRunKanbanBoard()` |
| `I run "kanban board" with machine output flag` | `IRunKanbanBoardJSON()` |
| `I run "kanban start" on task "ID"` | `IRunKanbanStart("ID")` |
| `I run "kanban start" on that task` | `IRunKanbanStartOnThatTask()` |
| `I run "kanban edit" on that task and add a description` | `IRunKanbanEditAddDescription(ctx.LastTaskID(), "...")` |
| `I run "kanban edit" on that task and update title to "T"` | `IRunKanbanEditTitle(ctx.LastTaskID(), "T")` |
| `I run "kanban delete" on "ID" and enter "y"` | `IRunKanbanDelete("ID", "y")` |
| `I run "kanban delete" on "ID" and enter "n"` | `IRunKanbanDelete("ID", "n")` |
| `I run "kanban delete" on "ID" with force flag` | `IRunKanbanDeleteForce("ID")` |
| `I commit with message "M"` | `ICommitWithMessage("M")` |
| `I commit with message containing the new task ID` | `ICommitWithTaskID()` |
| `the CI step runs after all tests pass` | `CIStepRunsPass()` |
| `the CI step runs after one or more tests fail` | `CIStepRunsFail()` |
| `the exit code is N` | `ExitCodeIs(N)` |
| `output contains "T"` / `output confirms "T"` / `output shows "T"` | `OutputContains("T")` |
| `output is valid JSON` | `OutputIsValidJSON()` |
| `the JSON array contains an object with fields F` | `JSONHasFields("F")` |
| `the task "ID" has status "S"` | `TaskHasStatus("ID", "S")` |
| `the task "ID" status remains "S"` | `TaskStatusRemains("ID", "S")` |
| `the task file for "ID" still exists` | `TaskFilePresent("ID")` |
| `the task file for "ID" is removed` | `TaskFileRemoved("ID")` |
| `the board shows "T" under <heading>` | `BoardShowsTaskUnder("T", "heading")` |
| `"kanban board" no longer lists "ID"` | `BoardNotListsTask("ID")` |
| `the git commit exit code is N` | `GitCommitExitCodeIs(N)` |
| `the kanban workspace is ready for use` | `WorkspaceReady()` |
| `the kanban workspace directory is created at "P"` | `WorkspaceReady()` (same assertion) |
| `the configuration file is created with default task pattern` | `ConfigFileHasDefaults()` |
| `the hook log file path is added to ".gitignore"` | `HookLogInGitignore()` |
| `no partial or temporary files remain` | `NoTempFilesRemain()` |
| `the updated task file is committed back to the repository` | `UpdatedTaskCommitted()` |
| `the git repository has no new commits from the delete operation` | `NoAutoCommitFromDelete()` |
| `output contains no kanban lines` | `NoKanbanOutputLines()` |
| `output contains no kanban transition lines` | `NoTransitionLines()` |
| `output contains no ANSI colour escape sequences` | `NoANSIEscapeCodes()` |
| `output contains no spinner characters` | `NoSpinnerChars()` |
