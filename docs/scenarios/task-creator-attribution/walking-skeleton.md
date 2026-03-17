# Walking Skeleton — task-creator-attribution

## Identified Skeleton

**Test function**: `TestTaskCreator_CreatorRecordedOnKanbanNew`
(`tests/acceptance/task_creator_test.go`)

## Litmus Test

1. **Title describes a user goal** — "Creator Recorded On Kanban New" — not a technical flow.
   The test name tells a non-engineer stakeholder what observable thing should happen. Pass.

2. **Given/When describe user context and action** — `InAGitRepo()` (developer's environment),
   `KanbanInitialised()` (prerequisite), `IRunKanbanNew("Fix login bug")` (a single user action).
   No internal mocks, no port calls. Pass.

3. **Then describes what the user observes** — exit code 0, stdout contains "Created TASK-",
   and the task file's front matter contains `created_by: Test User`. All three are observable
   without reading any internal code. Pass.

4. **A non-technical stakeholder can confirm** — "Yes, when a developer creates a task,
   their name should be recorded." Pass.

## What It Proves

A developer can run `kanban new` and their git identity is automatically captured as the task
creator. The full path is exercised: git identity resolution → use case → filesystem serialization.
No part of the feature's core loop is mocked at the acceptance level.

## Implementation Sequence

Implement tests one at a time, using `t.Skip("not yet implemented")` for tests that exercise
code not yet written. Recommended order:

1. **TestTaskCreator_CreatorRecordedOnKanbanNew** — walking skeleton. Establishes the core path.
   Requires: `domain.Task.CreatedBy`, `ports.Identity` + `GitPort.GetIdentity()`,
   `AddTaskInput.CreatedBy`, `GitAdapter.GetIdentity()`, `taskFrontMatter.CreatedBy`,
   `cli/new.go` identity call.

2. **TestTaskCreator_FrontMatterContainsCreatorField** — verifies schema completeness.

3. **TestTaskCreator_BoardShowsCreatorName** — requires board renderer update.

4. **TestTaskCreator_PreExistingTaskShowsDashOnBoard** — verifies backward compat (should work
   automatically if `CreatedBy = ""` renders as `--`).

5. **TestTaskCreator_JSONOutputHasCreatedByField** — requires `boardTaskJSON.CreatedBy`.

6. **TestTaskCreator_AtomicWriteNoTempFiles** — should pass without code changes if atomic
   write is preserved (it is, per NFR-01).

7. **TestTaskCreator_MissingIdentityFailsWithGuidance** — requires error guard in `cli/new.go`
   and `ErrGitIdentityNotConfigured` sentinel.

8. **TestTaskCreator_MissingIdentityWritesNoFile** — side-effect of #7, should pass together.

9. **TestTaskCreator_EditDoesNotChangeCreator** — requires `editFields` to exclude `created_by`
   (structural exclusion — likely already correct if `editFields` is not modified).
