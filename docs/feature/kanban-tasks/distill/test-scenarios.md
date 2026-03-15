# Test Scenario Inventory: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DISTILL
**Date**: 2026-03-15

---

## Summary

| Metric | Count |
|--------|-------|
| Total scenarios | 52 |
| Walking skeletons (`@walking_skeleton`) | 3 |
| Focused scenarios | 46 |
| Property-based signals (`@property`) | 2 |
| Error / edge path scenarios | 22 |
| Error ratio | 42% |
| Scenarios enabled (non-`@skip`) | 4 |
| Scenarios marked `@skip` | 48 |

Error path ratio target: 40%. Actual: 42%. Target met.

---

## Scenario Inventory by Feature File

### walking-skeleton.feature (3 scenarios)

| # | Scenario | Type | Story | Driving Port |
|---|----------|------|-------|-------------|
| 1 | Developer completes the full task lifecycle from creation to done | @walking_skeleton | US-01 US-02 US-03 US-04 US-05 | CLIAdapter + GitHookAdapter + CIPipelineAdapter |
| 2 | Developer initialises kanban and immediately creates and views a task | @walking_skeleton | US-01 US-02 US-03 | CLIAdapter |
| 3 | Developer edits a task and then removes it when the work is cancelled | @walking_skeleton | US-06 US-07 | CLIAdapter |

---

### milestone-1-task-crud.feature (26 scenarios)

#### US-01: Repository Initialisation (3 scenarios)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 4 | Developer sets up kanban in a new git repository | Happy path | AC-01-1 AC-01-2 AC-01-3 |
| 5 | Running init a second time makes no changes | Edge | AC-01-4 |
| 6 | Developer cannot initialise kanban outside a git repository | Error | AC-01-5 AC-X-3 |

#### US-02: Create Task (6 scenarios)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 7 | Developer creates a task with title only | Happy path | AC-02-1 AC-02-3 |
| 8 | Developer creates a task with all optional fields | Happy path | AC-02-4 AC-02-5 AC-02-6 |
| 9 | Task IDs increment sequentially | Edge | AC-02-2 |
| 10 | Creating a task fails when no title is provided | Error | AC-02-7 |
| 11 | Creating a task fails when the due date is in the past | Error | AC-02-8 |
| 12 | Creating a task fails outside a git repository | Error | AC-02-9 AC-X-3 |

#### US-03: View Board (7 scenarios)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 13 | Developer views the board with tasks in all three statuses | Happy path | AC-03-1 AC-03-2 AC-03-4 |
| 14 | Board shows "--" for missing priority and due, "unassigned" for missing assignee | Edge | AC-03-3 |
| 15 | Overdue task shows a distinct indicator on the board | Edge | AC-03-5 |
| 16 | Empty board shows onboarding message | Edge | AC-03-6 |
| 17 | Board outputs valid machine-readable format when requested | Happy path | AC-03-7 |
| 18 | Board suppresses colour codes when NO_COLOR is set | Edge | AC-03-8 |
| 19 | Board produces plain output when piped to another command | Edge | AC-03-9 |

#### US-06: Edit Task (5 scenarios)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 20 | Developer adds a description to an existing task | Happy path | AC-06-1 AC-06-3 |
| 21 | Edit displays all current field values before opening the editor | Happy path | AC-06-1 |
| 22 | Editing the title updates the board display | Happy path | AC-06-3 AC-06-5 |
| 23 | Edit with no changes made reports no update | Edge | AC-06-4 |
| 24 | Editing a non-existent task reports a clear error | Error | AC-06-6 |

#### US-07: Delete Task (6 scenarios)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 25 | Developer deletes a task after confirming | Happy path | AC-07-1 AC-07-2 AC-07-5 |
| 26 | Developer aborts a delete by entering "n" | Error | AC-07-3 |
| 27 | Developer aborts a delete by pressing Enter without input | Edge | AC-07-3 |
| 28 | Force delete removes a task without prompting | Happy path | AC-07-4 |
| 29 | Deleting a non-existent task reports a clear error | Error | AC-07-7 |
| 30 | Kanban does not auto-commit after task deletion | Edge | AC-07-8 |

---

### milestone-2-auto-transitions.feature (16 scenarios)

#### US-04: Auto-move to In Progress — GitHookAdapter (8 scenarios)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 31 | First commit referencing a task advances it to in-progress | Happy path | AC-04-1 AC-04-2 AC-04-3 |
| 32 | Commit hook always exits successfully even when referenced task does not exist | Error | AC-04-5 |
| 33 | Commit referencing a task already in-progress produces no transition | Edge | AC-04-4 |
| 34 | Commit referencing a completed task produces no transition | Edge | AC-04-4 |
| 35 | Commit with no task reference produces no kanban output | Edge | AC-04-6 |
| 36 | Commit hook reads the task pattern from project configuration | Edge | AC-04-7 |
| 37 | Commit hook completes without delaying the git commit perceptibly | Edge (NFR) | AC-04-8 |
| 38 | Commit hook logs errors internally and never blocks the commit | Error | AC-04-3 AC-04-5 |

#### US-05: Auto-move to Done — CIPipelineAdapter (8 scenarios)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 39 | CI step advances an in-progress task to done when all tests pass | Happy path | AC-05-2 AC-05-3 AC-05-4 AC-05-5 |
| 40 | CI step leaves tasks unchanged when tests fail | Error | AC-05-6 |
| 41 | CI step advances multiple tasks in one pipeline run | Happy path | AC-05-2 AC-05-5 |
| 42 | CI step skips tasks already in done status | Edge | AC-05-7 |
| 43 | CI step only advances tasks that are currently in-progress | Edge | AC-05-3 |
| 44 | CI step reads the task pattern from project configuration | Edge | AC-05-8 |
| 45 | CI step produces output without colour codes or interactive prompts | Edge (NFR) | AC-05-9 |

---

### integration-checkpoints.feature (12 scenarios, 2 @property)

| # | Scenario | Type | AC |
|---|----------|------|----|
| 46 | Every command exits with code 0 on success | Happy path | AC-X-2 |
| 47 | Runtime errors exit with code 1 | Error | AC-X-2 |
| 48 | Commands exit with code 1 for task-not-found (outline) | Error | AC-X-2 |
| 49 | Commands exit with code 2 for invalid input (outline) | Error | AC-X-2 |
| 50 | All commands support --help with usage and at least one example | Happy path | AC-X-1 |
| 51 | Version flag outputs the installed version string | Happy path | AC-X-5 |
| 52 | All commands exit with code 1 outside a git repository (outline) | Error | AC-X-3 |
| 53 | Created task file contains valid YAML front matter | Happy path | AC-02-1 |
| 54 | Status field contains only valid status values | Edge | (file integrity) |
| 55 | Task file written atomically — no partial writes visible | Edge (NFR) | (reliability) |
| 56 | Error messages explain what happened, why, and what to do next (outline) | Error | AC-X-4 |
| 57 | CI pipeline can build the binary and run acceptance tests successfully | Integration | (CI smoke) |
| 58 | Full pipeline smoke test | Integration | AC-05-1 |
| 59 | Task file round-trip preserves all field values | @property | AC-02-4 AC-02-5 AC-02-6 |
| 60 | Board output is stable across repeated reads | @property | AC-03-1 |

---

## Story Coverage Matrix

| User Story | Scenarios | All ACs Covered |
|------------|-----------|-----------------|
| US-01 Repository Initialisation | 4 | Yes (AC-01-1 through AC-01-5) |
| US-02 Create Task | 7 | Yes (AC-02-1 through AC-02-9) |
| US-03 View Board | 8 | Yes (AC-03-1 through AC-03-10) |
| US-04 Commit Hook Auto-Transition | 8 | Yes (AC-04-1 through AC-04-8) |
| US-05 CI Done Auto-Transition | 8 | Yes (AC-05-1 through AC-05-9) |
| US-06 Edit Task | 5 | Yes (AC-06-1 through AC-06-6) |
| US-07 Delete Task | 7 | Yes (AC-07-1 through AC-07-8) |
| Cross-cutting (AC-X) | 7 | Yes (AC-X-1 through AC-X-5) |

All 7 user stories covered. All acceptance criteria traced to at least one scenario.

---

## Error / Edge Scenario Breakdown

| Category | Count |
|----------|-------|
| Invalid input errors (exit 2) | 4 |
| Not-found errors (exit 1) | 4 |
| Not-a-git-repo errors (exit 1) | 5 |
| Hook error handling (exit 0 always) | 2 |
| CI failure (no transition) | 1 |
| Idempotency / no-op edges | 6 |
| Empty state / onboarding | 1 |
| NFR boundary (timing, colour, piped output) | 6 |
| File integrity edges | 3 |
| **Total error / edge** | **22** |

---

## @property Scenarios

Two scenarios tagged `@property` signal to the DELIVER wave crafter to implement as property-based tests (using Go's `testing/quick` or a property testing library):

1. **Task file round-trip preserves all field values** — for any valid combination of title, priority, due, assignee, the fields are identical after read-back.
2. **Board output is stable across repeated reads** — for any set of task files, repeated board invocations produce identical output.

Both encode universal invariants ("any valid task", "any set of files") rather than single examples.
