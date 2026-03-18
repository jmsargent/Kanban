# NOTE: Exact transitions.log line format is subject to RQ-1 resolution in DESIGN wave.
# This file uses an illustrative format. The requirement is the FIELDS present,
# not the exact delimiter or quoting. "Concept B enabled" in scenarios means
# "the kanban repository uses transitions log for state storage."
@feature:board-state-in-git
Feature: Task State Transitions with Audit Trail

  As Jon Santos, a developer using kanban-tasks in a git-native workflow,
  I want task state transitions to be recorded in an auditable, readable log
  so that I can reconstruct the history of any task without git expertise.

  Background:
    Given Jon's git repository is initialized at /repos/kanban-tasks
    And kanban is initialized in the repository
    And Jon's git identity is configured as "jon@kanbandev.io"

  # ─────────────────────────────────────────────────────────────
  # EPIC: kanban log — Quick Win (Phase 1)
  # ─────────────────────────────────────────────────────────────

  @phase:1 @story:US-BSG-01
  Scenario: Developer views transition history for a task with prior commits
    Given TASK-007 "Add retry logic to CI step" exists in .kanban/tasks/
    And the git log contains three commits that touched .kanban/tasks/TASK-007.md
    And the commits are dated 2026-03-15, 2026-03-15, and 2026-03-17
    When Jon runs: kanban log TASK-007
    Then the output header shows: "TASK-007: Add retry logic to CI step"
    And the output shows exactly 3 entries sorted oldest-first
    And each entry shows: timestamp, transition, author, trigger
    And the exit code is 0

  @phase:1 @story:US-BSG-01
  Scenario: Developer views history for a task with no commits yet
    Given TASK-003 "Add dark mode" exists in .kanban/tasks/
    And no commits have touched .kanban/tasks/TASK-003.md
    When Jon runs: kanban log TASK-003
    Then the output shows: "TASK-003: Add dark mode"
    And the output shows: "No transitions recorded yet."
    And the current status is shown as: "todo"
    And the exit code is 0

  @phase:1 @story:US-BSG-01
  Scenario: Developer runs kanban log for a task that does not exist
    Given no task with id TASK-999 exists
    When Jon runs: kanban log TASK-999
    Then the output shows: "Error: task TASK-999 not found"
    And the output suggests: "kanban board" to see all tasks
    And the exit code is 1

  @phase:1 @story:US-BSG-01
  Scenario: Developer runs kanban log outside a kanban repository
    Given the current directory is not inside a git repository with kanban initialized
    When Jon runs: kanban log TASK-001
    Then the output shows an error about the repository not being initialized
    And the output suggests: "kanban init" as next step
    And the exit code is 1

  @phase:1 @story:US-BSG-01
  Scenario: kanban log output uses domain language not raw git messages
    Given TASK-007 has a commit with message "kanban: start TASK-007"
    And TASK-007 has a commit with message "refactor: add retry logic TASK-007"
    When Jon runs: kanban log TASK-007
    Then the output shows "todo → in-progress" not the raw git message
    And the output shows the commit SHA alongside the human-readable transition
    And no raw git plumbing output is shown

  # ─────────────────────────────────────────────────────────────
  # EPIC: Concept B — Append-Only Transition Log (Phase 2)
  # ─────────────────────────────────────────────────────────────

  @phase:2 @story:US-BSG-02
  Scenario: Task created without status field in YAML
    Given kanban is initialized with Concept B enabled
    When Jon runs: kanban add -t "Add retry logic to CI step" -p high
    Then .kanban/tasks/TASK-007.md is created
    And the file contains: id, title, priority fields
    And the file does NOT contain a "status:" field
    And .kanban/tasks/TASK-007.md contains a comment pointing to transitions.log
    And kanban board shows TASK-007 in the TODO column

  @phase:2 @story:US-BSG-02
  Scenario: kanban start appends to transitions log without modifying task file
    Given TASK-007 exists with no entries in .kanban/transitions.log
    When Jon runs: kanban start TASK-007
    Then .kanban/transitions.log contains one new entry
    And the entry contains: TASK-007, todo->in-progress, jon@kanbandev.io, trigger:manual
    And .kanban/tasks/TASK-007.md is NOT modified
    And kanban board shows TASK-007 in the IN PROGRESS column
    And the exit code is 0

  @phase:2 @story:US-BSG-02
  Scenario: kanban start is idempotent — running twice does not double-record
    Given TASK-007 is already in-progress per .kanban/transitions.log
    When Jon runs: kanban start TASK-007 again
    Then the command exits with code 1
    And the error message says TASK-007 is already in-progress
    And .kanban/transitions.log has no duplicate entry for TASK-007

  @phase:2 @story:US-BSG-02
  Scenario: commit-msg hook appends to transitions log and does not write task files
    Given TASK-007 is in-progress per .kanban/transitions.log
    And Jon is about to commit with message "refactor: add retry logic TASK-007"
    When the git commit executes
    Then the commit-msg hook appends one entry to .kanban/transitions.log
    And .kanban/tasks/TASK-007.md is NOT modified by the hook
    And the commit exits with code 0

  @phase:2 @story:US-BSG-02
  Scenario: commit-msg hook exits 0 even when transitions log write fails
    Given the .kanban/transitions.log file is not writable
    When Jon commits with message "fix: something TASK-007"
    Then the commit succeeds (exit code 0)
    And a warning is printed to stderr containing "kanban: could not record transition"
    And the commit message is preserved as-is

  @phase:2 @story:US-BSG-02
  Scenario: kanban ci-done commits only transitions.log (not task files)
    Given TASK-007 is in-progress per .kanban/transitions.log
    And the CI commit range abc123..HEAD contains commits referencing TASK-007
    When kanban ci-done runs with --from=abc123 --to=HEAD
    Then .kanban/transitions.log has a new entry: TASK-007 in-progress->done
    And the git commit staged by ci-done contains only .kanban/transitions.log
    And no .kanban/tasks/*.md files are in the ci-done commit
    And the exit code is 0

  @phase:2 @story:US-BSG-02
  Scenario: kanban board derives status from transitions log
    Given .kanban/transitions.log contains:
      """
      2026-03-15T09:14:23Z TASK-007 todo->in-progress jon@kanbandev.io manual
      2026-03-17T16:42:11Z TASK-007 in-progress->done ci@kanbandev.io ci-done:e91c44d
      2026-03-16T11:00:00Z TASK-005 todo->in-progress jon@kanbandev.io manual
      """
    And TASK-003 exists with no transitions in the log
    When Jon runs: kanban board
    Then TASK-007 appears in the DONE column
    And TASK-005 appears in the IN PROGRESS column
    And TASK-003 appears in the TODO column

  @phase:2 @story:US-BSG-02
  Scenario: transitions log is rebase-safe
    Given TASK-007 has a transition entry in .kanban/transitions.log from a commit
    When Jon performs: git rebase -i HEAD~3 (squashing the commits)
    Then .kanban/transitions.log still contains the original transition entry
    And kanban board still shows TASK-007 in the correct column
    And kanban log TASK-007 still shows the transition history

  @phase:2 @story:US-BSG-02
  Scenario: transitions log append is atomic under concurrent writes
    Given two processes attempt to append to .kanban/transitions.log simultaneously
    When both appends complete
    Then .kanban/transitions.log contains exactly two new entries
    And neither entry is corrupted or truncated
    And the file is a valid line-separated log

  # ─────────────────────────────────────────────────────────────
  # EPIC: kanban board --me (Phase 3 — Independent)
  # ─────────────────────────────────────────────────────────────

  @phase:3 @story:US-BSG-03
  Scenario: Developer filters board to their own tasks using --me
    Given the board has 5 tasks total
    And TASK-005 and TASK-007 have assignee matching Jon's git identity (jon@kanbandev.io)
    And TASK-001, TASK-003, TASK-006 are assigned to other contributors
    When Jon runs: kanban board --me
    Then the output shows only TASK-005 and TASK-007
    And TASK-001, TASK-003, TASK-006 are not shown
    And the exit code is 0

  @phase:3 @story:US-BSG-03
  Scenario: --me filter warns about unassigned tasks
    Given the board has TASK-008 with no assignee field set
    When Jon runs: kanban board --me
    Then the output shows a warning: "1 unassigned task hidden — use 'kanban board' to see all"
    And TASK-008 is not shown in the filtered view

  @phase:3 @story:US-BSG-03
  Scenario: --me filter works without Concept B (Phase 1 compatible)
    Given the current system uses YAML status fields (pre-Concept B)
    And TASK-005 has assignee: jon@kanbandev.io in its YAML
    When Jon runs: kanban board --me
    Then the output shows only tasks assigned to jon@kanbandev.io
    And the filtering uses the assignee field from task YAML

  # ─────────────────────────────────────────────────────────────
  # PROPERTY: System-wide quality guarantees
  # ─────────────────────────────────────────────────────────────

  @property @story:US-BSG-01 @story:US-BSG-02
  Scenario: kanban log responds within acceptable time on large transition log
    Given .kanban/transitions.log contains 1000 transition entries across 50 tasks
    When Jon runs: kanban log TASK-007
    Then the output appears within 500ms
    And the exit code is 0

  @property @story:US-BSG-02
  Scenario: transitions log append preserves existing entries
    Given .kanban/transitions.log contains 50 existing entries
    When a new transition is appended by any mechanism (kanban start, hook, ci-done)
    Then the log contains exactly 51 entries
    And all 50 original entries are byte-for-byte identical
    And the new entry is the last line in the file
