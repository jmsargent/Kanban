Feature: Explicit State Transitions
  As a developer or CI pipeline
  I want kanban to update task state explicitly without auto-committing
  So that git history is fully owned by the developer or CI config

  Background:
    Given a kanban-initialised repository
    And a task "TASK-001" exists with status "in-progress"
    And a task "TASK-002" exists with status "todo"
    And a task "TASK-003" exists with status "done"

  # --- kanban done ---

  Scenario: Developer marks an in-progress task as done
    When the developer runs "kanban done TASK-001"
    Then the process exits with code 0
    And stdout contains "kanban: TASK-001 moved in-progress -> done"
    And the task file for TASK-001 has YAML field "status" equal to "done"
    And no git subprocess was invoked

  Scenario: Developer marks a todo task as done directly
    When the developer runs "kanban done TASK-002"
    Then the process exits with code 0
    And stdout contains "kanban: TASK-002 moved todo -> done"
    And the task file for TASK-002 has YAML field "status" equal to "done"

  Scenario: Marking an already-done task is idempotent
    When the developer runs "kanban done TASK-003"
    Then the process exits with code 0
    And stdout contains "kanban: TASK-003 already done"
    And the task file for TASK-003 is unchanged

  Scenario: Marking a non-existent task exits with error
    When the developer runs "kanban done TASK-999"
    Then the process exits with code 1
    And stderr contains "task not found"
    And no task files are modified

  # --- kanban ci-done (no commit) ---

  Scenario: CI marks tasks done after pipeline passes
    Given commits in the pipeline range reference "TASK-001"
    When the CI runs "kanban ci-done --since=HEAD^"
    Then the process exits with code 0
    And stdout contains "kanban: TASK-001 moved in-progress -> done"
    And the task file for TASK-001 has YAML field "status" equal to "done"
    And no "git add" subprocess was invoked
    And no "git commit" subprocess was invoked

  Scenario: CI finds no tasks in commit range — silent success
    Given no commits in the pipeline range reference any task IDs
    When the CI runs "kanban ci-done --since=HEAD^"
    Then the process exits with code 0
    And stdout is empty

  Scenario: CI ci-done is idempotent for already-done tasks
    Given commits in the pipeline range reference "TASK-003"
    When the CI runs "kanban ci-done --since=HEAD^"
    Then the process exits with code 0
    And the task file for TASK-003 is unchanged

  Scenario: CI output is plain text when NO_COLOR is set
    Given the environment variable "NO_COLOR" is set
    When the CI runs "kanban ci-done --since=HEAD^"
    Then stdout contains no ANSI escape codes

  # --- kanban board reads from YAML ---

  Scenario: Board displays tasks from YAML status fields
    When the developer runs "kanban board"
    Then TASK-001 appears in the "IN PROGRESS" column
    And TASK-002 appears in the "TODO" column
    And TASK-003 appears in the "DONE" column
    And no transitions.log file is read

  Scenario: Board works when no transitions.log exists
    Given no ".kanban/transitions.log" file exists
    When the developer runs "kanban board"
    Then the process exits with code 0

  # --- hook removed ---

  Scenario: Leftover commit-msg hook is a safe no-op
    Given a file ".git/hooks/commit-msg" that delegates to "kanban _hook commit-msg"
    When a git commit is made
    Then "kanban _hook commit-msg" exits with code 0
    And no task files are modified
    And no transitions.log is written

  Scenario: install-hook command is removed
    When the developer runs "kanban install-hook"
    Then the process exits with code 1
    And stderr contains "install-hook has been removed"
