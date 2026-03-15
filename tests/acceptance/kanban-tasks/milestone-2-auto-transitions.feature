Feature: Automatic task status transitions via git commit hook and CI pipeline
  As a developer working in a git-native workflow
  I want tasks to advance automatically when I commit and when CI passes
  So that the board stays accurate with zero manual status updates

  # US-04, US-05

  Background:
    Given I am working in a git repository
    And the repository is initialised with kanban

  # ---------------------------------------------------------------------------
  # US-04: Auto-move to In Progress (Commit Hook)
  # Driving port: GitHookAdapter (kanban _hook commit-msg)
  # ---------------------------------------------------------------------------

  Scenario: First commit referencing a task advances it to in-progress
    Given a task "Fix OAuth login bug" exists with status "todo" as "TASK-001"
    And the git commit hook is installed
    When I commit with message "TASK-001: reproduce OAuth bug on Chrome"
    Then the task "TASK-001" has status "in-progress"
    And the commit output contains "kanban: TASK-001 moved  todo -> in-progress"
    And "kanban board" shows "TASK-001" under IN PROGRESS
    And the git commit exit code is 0

  @skip
  Scenario: Commit hook always exits successfully even when the referenced task does not exist
    Given the git commit hook is installed
    And no task file exists for "TASK-099"
    When I commit with message "TASK-099: working on this"
    Then the git commit exit code is 0
    And the commit output contains a warning about "TASK-099 not found"

  @skip
  Scenario: Commit referencing a task already in-progress produces no transition
    Given a task "API work" exists with status "in-progress" as "TASK-002"
    And the git commit hook is installed
    When I commit with message "TASK-002: add throttle middleware"
    Then the task "TASK-002" status remains "in-progress"
    And the commit output contains no kanban transition lines

  @skip
  Scenario: Commit referencing a completed task produces no transition
    Given a task "Finished work" exists with status "done" as "TASK-003"
    And the git commit hook is installed
    When I commit with message "TASK-003: minor cleanup"
    Then the task "TASK-003" status remains "done"
    And the commit output contains no kanban transition lines

  @skip
  Scenario: Commit with no task reference produces no kanban output
    Given a task "Silent task" exists with status "todo" as "TASK-001"
    And the git commit hook is installed
    When I commit with message "fix typo in README"
    Then the commit output contains no kanban lines
    And the task "TASK-001" status remains "todo"
    And the git commit exit code is 0

  @skip
  Scenario: Commit hook reads the task pattern from the project configuration
    Given the project configuration sets a custom task pattern "PROJ-[0-9]+"
    And a task "Custom pattern task" exists with status "todo" as "TASK-001"
    And the git commit hook is installed
    When I commit with message "PROJ-001: implement feature"
    Then the task matching "PROJ-001" advances to "in-progress"

  @skip
  Scenario: Commit hook completes without delaying the git commit perceptibly
    Given a task "Performance test task" exists with status "todo" as "TASK-001"
    And the git commit hook is installed
    When I commit with message "TASK-001: first implementation"
    Then the hook completes within 500 milliseconds
    And the git commit exit code is 0

  @skip
  Scenario: Commit hook logs errors internally and never blocks the commit
    Given the repository has a corrupted kanban configuration
    And the git commit hook is installed
    When I commit with message "TASK-001: commit despite error"
    Then the git commit exit code is 0
    And the error is recorded in the hook log file

  # ---------------------------------------------------------------------------
  # US-05: Auto-move to Done (CI Pipeline)
  # Driving port: CIPipelineAdapter (kanban ci-done)
  # ---------------------------------------------------------------------------

  Scenario: CI step advances an in-progress task to done when all tests pass
    Given a task "Implement throttle middleware" exists with status "in-progress" as "TASK-003"
    And the pipeline run includes a commit with "TASK-003" in the message
    When the CI step runs after all tests pass
    Then the task "TASK-003" has status "done"
    And the CI log contains "[kanban] TASK-003 moved  in-progress -> done"
    And the updated task file is committed back to the repository
    And the CI step exit code is 0

  @skip
  Scenario: CI step leaves tasks unchanged when tests fail
    Given a task "Failing work" exists with status "in-progress" as "TASK-003"
    And the pipeline run includes a commit with "TASK-003" in the message
    When the CI step runs after one or more tests fail
    Then the task "TASK-003" status remains "in-progress"
    And the CI log contains no transition lines

  @skip
  Scenario: CI step advances multiple tasks in one pipeline run
    Given a task "OAuth fix" exists with status "in-progress" as "TASK-001"
    And a task "Rate limiting" exists with status "in-progress" as "TASK-003"
    And the pipeline run includes commits referencing both "TASK-001" and "TASK-003"
    When the CI step runs after all tests pass
    Then the task "TASK-001" has status "done"
    And the task "TASK-003" has status "done"
    And the CI log contains a transition line for each moved task

  @skip
  Scenario: CI step skips tasks already in done status
    Given a task "Already completed" exists with status "done" as "TASK-001"
    And the pipeline run includes a commit with "TASK-001" in the message
    When the CI step runs after all tests pass
    Then the task "TASK-001" status remains "done"
    And the CI log contains no transition lines for "TASK-001"

  @skip
  Scenario: CI step only advances tasks that are currently in-progress
    Given a task "Not started yet" exists with status "todo" as "TASK-001"
    And the pipeline run includes a commit with "TASK-001" in the message
    When the CI step runs after all tests pass
    Then the task "TASK-001" status remains "todo"
    And the CI log contains no transition lines for "TASK-001"

  @skip
  Scenario: CI step reads the task pattern from the project configuration
    Given the project configuration sets a custom task pattern "PROJ-[0-9]+"
    And a task "Custom CI task" exists with status "in-progress" as "TASK-001"
    And the pipeline run includes a commit with "PROJ-001" in the message
    When the CI step runs after all tests pass
    Then the task matching "PROJ-001" advances to "done"

  @skip
  Scenario: CI step produces output without colour codes or interactive prompts
    Given a task "CI output test" exists with status "in-progress" as "TASK-001"
    And the pipeline run includes a commit with "TASK-001" in the message
    When the CI step runs after all tests pass
    Then the CI log output contains no ANSI colour escape sequences
    And the CI step runs without requiring any interactive input
