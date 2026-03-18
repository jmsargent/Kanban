# Test Scenarios — board-state-in-git

Full mapping of all 44 acceptance criteria to test function names.
Organized by story and milestone.

---

## US-BSG-01: kanban log (Walking Skeleton)

File: `tests/acceptance/kanban_log_test.go`

| AC | Description | Test Function | Status |
|----|-------------|---------------|--------|
| AC-01-1 | Output identifies task by ID and title | `TestKanbanLog_ShowsHeader_WhenTaskHasHistory` | Enabled |
| AC-01-2 | Each entry shows timestamp, from→to, author email, trigger | `TestKanbanLog_ShowsTransitionFields_InEachEntry` | Skip (pending) |
| AC-01-3 | Entries sorted oldest-first | `TestKanbanLog_SortsEntries_OldestFirst` | Skip (pending) |
| AC-01-4 | No transitions recorded message when task has no history | `TestKanbanLog_ShowsNoTransitions_WhenTaskHasNoCommits` | Enabled |
| AC-01-5 | Exit 1 with "not found" when task ID unknown | `TestKanbanLog_ExitsOne_WhenTaskNotFound` | Enabled |
| AC-01-6 | Output suggests "kanban board" when task not found | `TestKanbanLog_SuggestsKanbanBoard_WhenTaskNotFound` | Enabled |
| AC-01-7 | Exit 1 when kanban not initialised | `TestKanbanLog_ExitsOne_WhenNotInitialised` | Enabled |
| AC-01-8 | Output suggests "kanban init" when not initialised | `TestKanbanLog_ExitsOne_WhenNotInitialised` | Enabled (combined with AC-01-7) |
| AC-01-9 | Domain language used ("todo", "in-progress", "done") not raw git messages | `TestKanbanLog_UsesDomainLanguage_NotRawGitMessages` | Skip (pending) |
| AC-01-10 | Commit SHA appears as supplementary context | `TestKanbanLog_ShowsCommitSHA_AsSupplementaryContext` | Skip (pending) |
| AC-01-11 | Performance: completes within 2s for 10,000+ commit repos | **Excluded — see note below** | N/A |

### Walking skeleton passing criteria

The walking skeleton for US-BSG-01 is done when these 5 tests pass:

1. `TestKanbanLog_ShowsHeader_WhenTaskHasHistory` — task identified by ID and title
2. `TestKanbanLog_ShowsNoTransitions_WhenTaskHasNoCommits` — graceful empty state
3. `TestKanbanLog_ExitsOne_WhenTaskNotFound` — error on unknown task
4. `TestKanbanLog_SuggestsKanbanBoard_WhenTaskNotFound` — actionable error message
5. `TestKanbanLog_ExitsOne_WhenNotInitialised` — error when not set up

### AC-01-11 Performance Benchmark: Exclusion Rationale

AC-01-11 specifies that `kanban log` must complete within 2 seconds for a repository
with 10,000+ commits. This criterion is excluded from the automated acceptance suite
because:

1. Creating a 10,000-commit repository in CI would be prohibitively slow (minutes of
   git operations per test run).
2. Performance benchmarks are environment-sensitive and produce false failures when
   CI agents are under load.
3. The criterion is a non-functional quality bar, not a user-observable outcome that
   can be demonstrated to a stakeholder from test output.

Recommended approach: implement as a local-only Go benchmark:

```go
// BenchmarkKanbanLog_LargeRepo is a local-only benchmark.
// Run with: go test -bench=BenchmarkKanbanLog_LargeRepo -benchtime=1x ./tests/acceptance/
func BenchmarkKanbanLog_LargeRepo(b *testing.B) {
    // Create repo with 10,000 commits referencing a task...
    // b.ResetTimer()
    // run(ctx, "log", taskID)
    // Assert elapsed < 2 seconds
}
```

This benchmark should be run manually before releases and is not part of CI.

---

## US-BSG-02: Append-only transitions log (Milestone 1)

File: `tests/acceptance/transitions_log_test.go`

All 26 tests are `t.Skip("pending: US-BSG-02 not yet implemented")`.

### Task creation (AC-02-1 through AC-02-3)

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-02-1 | New task file has no status field | `TestTransitionsLog_NewTaskHasNoStatusField` |
| AC-02-2 | New task file contains comment about transitions.log | `TestTransitionsLog_NewTaskFileHasLogComment` |
| AC-02-3 | New task appears in TODO column | `TestTransitionsLog_NewTaskAppearsInTodoColumn` |

### kanban start (AC-02-4 through AC-02-9)

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-02-4 | Start appends exactly one line | `TestTransitionsLog_StartAppendsOneLine` |
| AC-02-5 | Start line contains all 5 required fields | `TestTransitionsLog_StartLineContainsRequiredFields` |
| AC-02-6 | Start does not modify task file | `TestTransitionsLog_StartDoesNotModifyTaskFile` |
| AC-02-7 | Board shows IN PROGRESS after start | `TestTransitionsLog_BoardShowsInProgress_AfterStart` |
| AC-02-8 | Start on in-progress task exits 0 with message | `TestTransitionsLog_StartOnInProgress_ExitsWithMessage` |
| AC-02-9 | Start idempotent — no duplicate log entry | `TestTransitionsLog_StartIdempotent_NoDuplicateEntry` |

### Commit-msg hook (AC-02-10 through AC-02-15)

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-02-10 | Hook appends one line on commit | `TestTransitionsLog_HookAppendsOneLine_OnCommit` |
| AC-02-11 | Hook does not modify task files | `TestTransitionsLog_HookDoesNotModifyTaskFiles` |
| AC-02-12 | Hook exits 0 on success | `TestTransitionsLog_HookExitsZero_OnSuccess` |
| AC-02-13 | Hook exits 0 when log unwritable (never blocks commit) | `TestTransitionsLog_HookExitsZero_WhenLogUnwritable` |
| AC-02-14 | Hook writes stderr warning when log unwritable | `TestTransitionsLog_HookWritesStderrWarning_WhenLogUnwritable` |
| AC-02-15 | Hook does not modify commit message when log fails | `TestTransitionsLog_HookDoesNotModifyCommitMessage_WhenLogFails` |

### Board (AC-02-16 through AC-02-17)

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-02-16 | Board derives status from log, not YAML | `TestTransitionsLog_BoardDerivesStatus_FromLogNotYAML` |
| AC-02-17 | Board shows TODO when no log entries | `TestTransitionsLog_BoardShowsTodo_WhenNoLogEntries` |

### ci-done (AC-02-18 through AC-02-20)

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-02-18 | ci-done appends done entry | `TestTransitionsLog_CiDoneAppendsDoneEntry` |
| AC-02-19 | ci-done commit contains only transitions.log | `TestTransitionsLog_CiDoneCommitsOnlyTransitionsLog` |
| AC-02-20 | ci-done commit excludes task files | `TestTransitionsLog_CiDoneCommitExcludesTaskFiles` |

### Rebase safety (AC-02-21 through AC-02-22)

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-02-21 | Entries preserved after rebase | `TestTransitionsLog_RebaseSafe_EntriesPreserved` |
| AC-02-22 | Board correct after rebase | `TestTransitionsLog_RebaseSafe_BoardCorrectAfterRebase` |

### Concurrency and edge cases (AC-02-23 through AC-02-26)

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-02-23 | Concurrent appends produce correct line count | `TestTransitionsLog_ConcurrentAppends_CorrectLineCount` |
| AC-02-24 | Concurrent appends produce no truncated lines | `TestTransitionsLog_ConcurrentAppends_NoTruncatedLines` |
| AC-02-25 | Deleted task excluded from board | `TestTransitionsLog_DeletedTask_ExcludedFromBoard` |
| AC-02-26 | ci-done with no matching tasks exits 0 with message | `TestTransitionsLog_CiDoneWithNoMatchingTasks_ExitsZeroWithMessage` |

---

## US-BSG-03: kanban board --me (Milestone 2)

File: `tests/acceptance/board_me_test.go`

All 5 tests are `t.Skip("pending: US-BSG-03 not yet implemented")`.

| AC | Description | Test Function |
|----|-------------|---------------|
| AC-03-1 | board --me shows only current developer's tasks | `TestBoardMe_ShowsOnlyCurrentDeveloperTasks` |
| AC-03-2 | board --me hides tasks assigned to others | `TestBoardMe_ShowsOnlyCurrentDeveloperTasks` (combined) |
| AC-03-3 | board --me warns about unassigned tasks | `TestBoardMe_WarnsAboutUnassignedTasks` |
| AC-03-4 | Warning message identifies unassigned tasks exist | `TestBoardMe_WarnsAboutUnassignedTasks` (combined) |
| AC-03-5 | board --me shows empty board gracefully | `TestBoardMe_ShowsEmptyBoardGracefully_WhenNoMatchingTasks` |
| AC-03-6 | board (no flag) still shows all tasks | `TestBoardMe_DoesNotAffectUnfilteredBoard` |
| AC-03-7 | board --me works with transitions.log status storage | `TestBoardMe_WorksWithTransitionsLogStatusStorage` |

---

## Coverage summary

| Story | ACs | Test functions | Error/edge scenarios | Ratio |
|-------|-----|----------------|---------------------|-------|
| US-BSG-01 | 11 (10 automated + 1 excluded) | 9 | AC-01-5, AC-01-6, AC-01-7/8, AC-01-9 = 4 | 44% |
| US-BSG-02 | 26 | 26 | AC-02-8, AC-02-9, AC-02-13, AC-02-14, AC-02-15, AC-02-23, AC-02-24, AC-02-25, AC-02-26 = 9 | 35% |
| US-BSG-03 | 7 | 5 | AC-03-3, AC-03-4, AC-03-5 = 3 | 43% |
| **Total** | **44** | **40** | **16** | **40%** |

Error path ratio meets the 40% target across the full feature.
