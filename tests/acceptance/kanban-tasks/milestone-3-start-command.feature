Feature: Explicitly start work on a task
  As a developer managing work in the terminal
  I want to mark a task as in-progress with a single command
  So that the board reflects my active work without requiring a git commit first

  # US-08 (start command)
  # Driving port: CLI (kanban start)

  Background:
    Given I am working in a git repository

  # ---------------------------------------------------------------------------
  # Walking Skeleton: Developer begins work on a task
  # Driving port: CLI (kanban start)
  # ---------------------------------------------------------------------------

  @walking_skeleton @ported
  Scenario: Developer starts a todo task and sees it move to in-progress
    Given the repository is initialised with kanban
    And a task "Fix OAuth login bug" exists with status "todo" as "TASK-001"
    When I run "kanban start" on task "TASK-001"
    Then the task "TASK-001" has status "in-progress"
    And the exit code is 0
    And output contains "Started TASK-001"

  # ---------------------------------------------------------------------------
  # Focused scenarios
  # Driving port: CLI (kanban start)
  # ---------------------------------------------------------------------------

  @ported
  Scenario: Starting a task that is already in-progress reports no change needed
    Given the repository is initialised with kanban
    And a task "API rate limiting" exists with status "in-progress" as "TASK-002"
    When I run "kanban start" on task "TASK-002"
    Then the exit code is 0
    And output contains "already in progress"
    And the task "TASK-002" status remains "in-progress"

  @ported
  Scenario: Starting a completed task is rejected as an error
    Given the repository is initialised with kanban
    And a task "Deploy to staging" exists with status "done" as "TASK-003"
    When I run "kanban start" on task "TASK-003"
    Then the exit code is 1
    And output contains "already finished"

  @ported
  Scenario: Starting a task that does not exist reports a clear error
    Given the repository is initialised with kanban
    When I run "kanban start" on task "TASK-099"
    Then the exit code is 1
    And output contains "not found"

  @ported
  Scenario: Starting a task without kanban initialised reports setup guidance
    Given the repository has no kanban setup
    When I run "kanban start" on task "TASK-001"
    Then the exit code is 1
    And output contains "kanban not initialised"
