# Test Scenarios — task-creator-attribution

## Story Reference

| US | Effort | Test Functions |
|----|--------|---------------|
| US-01: Automatic creator capture | S | TestTaskCreator_CreatorRecordedOnKanbanNew (walking skeleton), TestTaskCreator_FrontMatterContainsCreatorField, TestTaskCreator_AtomicWriteNoTempFiles |
| US-02: Creator visible on board | XS | TestTaskCreator_BoardShowsCreatorName, TestTaskCreator_PreExistingTaskShowsDashOnBoard, TestTaskCreator_JSONOutputHasCreatedByField |
| US-03: Error guidance | XS | TestTaskCreator_MissingIdentityFailsWithGuidance, TestTaskCreator_MissingIdentityWritesNoFile |
| US-04: Creator immutability | XS | TestTaskCreator_EditDoesNotChangeCreator |

## Test File

`tests/acceptance/task_creator_test.go`

## New DSL Step Factories

`tests/acceptance/dsl/creator_steps.go`

### Setup Steps (new)

| Factory | Purpose |
|---------|---------|
| `InAGitRepoWithoutGitIdentity()` | Git repo with HOME isolated → `git config user.name` returns empty → triggers error path |
| `APreExistingTaskWithoutCreator(title)` | Writes a task file without `created_by` to simulate pre-feature tasks |

### Assertion Steps (new)

| Factory | Purpose |
|---------|---------|
| `TaskHasCreator(taskID, creator)` | Reads task file, asserts `created_by: <creator>` in front matter |
| `BoardRowForTaskContains(taskID, text)` | Finds task row in last board output, asserts row contains text |
| `TasksDirIsEmpty()` | Asserts `.kanban/tasks/` has no `.md` files — verifies error path wrote nothing |

### Existing DSL Steps Reused

| Factory | Scenario |
|---------|---------|
| `InAGitRepo()` | All happy-path tests (already configures `user.name = "Test User"`) |
| `KanbanInitialised()` | All tests |
| `ATaskExists(title)` | US-02 board tests (sets up a task for display) |
| `IRunKanbanNew(title)` | US-01, US-03 |
| `IRunKanbanBoard()` | US-02 |
| `IRunKanbanBoardJSON()` | US-02 AC-02-3 |
| `IRunKanbanEditTitle(taskID, newTitle)` | US-04 |
| `ExitCodeIs(code)` | All tests |
| `StdoutContains(text)` | US-01 |
| `StderrContains(text)` | US-03 |
| `TaskHasStatus(taskID, status)` | US-01 AC-01-2 |
| `NoTempFilesRemain()` | US-01 AC-01-3 |
| `OutputIsValidJSON()` | US-02 AC-02-3 |
| `JSONHasFields(fields)` | US-02 AC-02-3 |

## Scenario Inventory

| # | Test Function | AC | Category |
|---|--------------|-----|---------|
| 1 | TestTaskCreator_CreatorRecordedOnKanbanNew | AC-01-1 | Walking Skeleton / Happy Path |
| 2 | TestTaskCreator_FrontMatterContainsCreatorField | AC-01-2 | Happy Path |
| 3 | TestTaskCreator_AtomicWriteNoTempFiles | AC-01-3 | Non-Functional / Safety |
| 4 | TestTaskCreator_BoardShowsCreatorName | AC-02-1 | Happy Path |
| 5 | TestTaskCreator_PreExistingTaskShowsDashOnBoard | AC-02-2 | Edge Case / Backward Compatibility |
| 6 | TestTaskCreator_JSONOutputHasCreatedByField | AC-02-3 | Happy Path |
| 7 | TestTaskCreator_MissingIdentityFailsWithGuidance | AC-03-1 | Error Path |
| 8 | TestTaskCreator_MissingIdentityWritesNoFile | AC-03-2 | Error Path / Safety |
| 9 | TestTaskCreator_EditDoesNotChangeCreator | AC-04-2 | Edge Case / Audit Integrity |

## Coverage Summary

- Total scenarios: 9
- Walking skeleton: 1 (Scenario #1)
- Happy path: 3 (Scenarios #2, #4, #6)
- Error path: 2 (Scenarios #7, #8)
- Edge / backward compat: 2 (Scenarios #5, #9)
- Safety / non-functional: 1 (Scenario #3)
- All 9 acceptance criteria covered (AC-01-1 through AC-04-2)

## Note on AC-04-1

AC-04-1 (edit temp file does not expose `created_by`) is not directly verifiable at
the acceptance level without intercepting the temp file mid-edit. Instead, the behavioral
equivalent (AC-04-2: saved task retains original creator after edit) is tested end-to-end.
The structural guarantee for AC-04-1 is provided by the absence of `created_by` from
`editFields` in the filesystem adapter, which is covered by use case unit tests.
