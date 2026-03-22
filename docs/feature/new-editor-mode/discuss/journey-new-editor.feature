Feature: kanban new — editor mode
  As a developer using kanban as my git-native task tracker
  I want to run "kanban new" with no arguments and have my editor open
  So that I can fill in task details interactively without typing the title inline

  Background:
    Given Alex is in a git repository with kanban initialised
    And Alex's git identity is configured (name and email)
    And Alex's $EDITOR environment variable is set to "nvim"

  # -------------------------------------------------------------------------
  # Happy Path Scenarios
  # -------------------------------------------------------------------------

  Scenario: Editor opens when kanban new is invoked with no arguments
    When Alex runs "kanban new" via CLI with no arguments
    Then the editor opens with a blank task template
    And the template contains fields: title, priority, assignee, description
    And the template contains comment lines explaining that title is required
    And all editable fields are initially empty

  Scenario: Task is created after Alex fills in title and saves
    Given the editor has opened with a blank task template
    When Alex sets the title field to "Fix nil pointer in auth handler"
    And Alex saves and quits the editor
    Then the process prints "Created TASK-042: Fix nil pointer in auth handler"
    And the process prints "Hint: reference TASK-042 in your next commit to start tracking"
    And the process exits with code 0
    And a task file exists at ".kanban/tasks/TASK-042.md" with title "Fix nil pointer in auth handler"

  Scenario: Task is created with all optional fields populated
    Given the editor has opened with a blank task template
    When Alex sets title to "Refactor payment gateway client"
    And Alex sets priority to "P2"
    And Alex sets assignee to "jordan"
    And Alex sets description to "Extract HTTP client into its own package"
    And Alex saves and quits the editor
    Then the task file contains priority "P2"
    And the task file contains assignee "jordan"
    And the task file contains description "Extract HTTP client into its own package"
    And the process exits with code 0

  Scenario: Success output format matches existing kanban new <title> behaviour
    Given Alex runs "kanban new" with no arguments and fills in title "Update README examples"
    When the task is successfully created as TASK-007
    Then stdout is exactly:
      """
      Created TASK-007: Update README examples
      Hint: reference TASK-007 in your next commit to start tracking
      """
    And the process exits with code 0

  Scenario: Inline title still works when argument is provided
    When Alex runs "kanban new" with argument "Quick hotfix for prod"
    Then no editor is opened
    And the process prints "Created TASK-043: Quick hotfix for prod"
    And the process exits with code 0

  # -------------------------------------------------------------------------
  # Error Path Scenarios
  # -------------------------------------------------------------------------

  Scenario: Empty title after editor save is rejected
    Given the editor has opened with a blank task template
    When Alex saves and quits the editor without filling in the title field
    Then the process prints to stderr "title cannot be empty"
    And the process exits with code 2
    And no task file is created

  Scenario: Editor aborted with no changes produces no task
    Given the editor has opened with a blank task template
    When Alex quits the editor without saving (e.g. :q! in vim)
    Then the process prints to stderr "title cannot be empty"
    And the process exits with code 2
    And no task file is created

  Scenario: $EDITOR not set and vi is unavailable
    Given Alex's $EDITOR environment variable is unset
    And "vi" is not present in PATH
    When Alex runs "kanban new" via CLI with no arguments
    Then the process prints to stderr a message containing "open editor"
    And the process exits with code 1

  Scenario: Running outside a git repository
    Given Alex is not in a git repository
    When Alex runs "kanban new" via CLI with no arguments
    Then the process prints to stderr "Not a git repository"
    And the process exits with code 1

  Scenario: kanban not yet initialised in the repository
    Given Alex is in a git repository without kanban initialised
    When Alex runs "kanban new" via CLI with no arguments
    Then the process prints to stderr "kanban not initialised — run 'kanban init' first"
    And the process exits with code 1
