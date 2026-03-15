Feature: Developer tracks work end-to-end without leaving the terminal
  As a developer on a shared git repository
  I want to capture tasks, see the board, work through them, and finish them automatically
  So that the team board stays accurate with no extra effort beyond normal git usage

  Background:
    Given I am working in a git repository

  @walking_skeleton
  Scenario: Developer completes the full task lifecycle from creation to done
    Given the repository has no kanban setup
    When I run "kanban init"
    Then the kanban workspace is ready for use
    And I run "kanban add" with title "Fix OAuth login bug"
    Then a new task is created with status "todo"
    And I run "kanban board"
    Then the board shows "Fix OAuth login bug" under TODO
    When I commit with message containing the new task ID
    Then the task status advances to "in-progress"
    And the board shows the task under IN PROGRESS
    When the CI pipeline passes with that task referenced in commits
    Then the task status advances to "done"
    And the board shows the task under DONE

  @walking_skeleton @skip
  Scenario: Developer initialises kanban and immediately creates and views a task
    Given the repository has no kanban setup
    When I run "kanban init"
    Then output confirms "Initialised kanban at .kanban/"
    And the exit code is 0
    When I run "kanban add" with title "Write integration tests"
    Then output shows the created task ID and title
    And output contains a commit tip referencing the task ID
    And the exit code is 0
    When I run "kanban board"
    Then output groups tasks under TODO, IN PROGRESS, and DONE headings
    And "Write integration tests" appears under TODO
    And the exit code is 0

  @walking_skeleton
  Scenario: Developer edits a task and then removes it when the work is cancelled
    Given the repository is initialised with kanban
    And a task "Migrate database schema" exists with status "todo"
    When I run "kanban edit" on that task and update the title to "Migrate user table schema"
    Then output confirms which fields changed
    And the board shows the updated title
    When I run "kanban delete" on that task and confirm with "y"
    Then the task is no longer on the board
    And output suggests a git commit command to record the deletion
    And the exit code is 0
