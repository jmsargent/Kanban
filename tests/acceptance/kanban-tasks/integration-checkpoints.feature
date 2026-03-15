Feature: Exit codes, error handling, file format integrity, and CI/CD smoke tests
  As a developer and CI operator
  I want consistent, predictable exit codes and error messages across all commands
  So that scripts, hooks, and pipelines can rely on kanban's behaviour

  # Cross-cutting ACs: AC-X-1 through AC-X-5
  # Exit code contract: 0=success, 1=runtime error, 2=usage error

  Background:
    Given I am working in a git repository

  # ---------------------------------------------------------------------------
  # Exit Code Contract
  # Driving port: CLI (subprocess invocation)
  # ---------------------------------------------------------------------------

  @skip
  Scenario: Every command exits with code 0 on success
    Given the repository is initialised with kanban
    When I run each of the following commands successfully:
      | command                         |
      | kanban board                    |
      | kanban add "Verify exit codes"  |
    Then each command exits with code 0

  @skip
  Scenario: Runtime errors exit with code 1
    Given the repository is initialised with kanban
    When I run "kanban board" in a directory that is not a git repository
    Then the exit code is 1
    And output contains an actionable error message

  @skip
  Scenario Outline: Commands exit with code 1 for task-not-found errors
    Given the repository is initialised with kanban
    When I run <command> for a non-existent task "TASK-099"
    Then the exit code is 1
    And output contains an actionable error message

    Examples:
      | command        |
      | kanban edit    |
      | kanban delete  |

  @skip
  Scenario Outline: Commands exit with code 2 for invalid input
    Given the repository is initialised with kanban
    When I run <command>
    Then the exit code is 2
    And output describes what went wrong and what to do next

    Examples:
      | command                                    |
      | kanban add with empty title                |
      | kanban add with past due date 2025-01-01   |

  @skip
  Scenario: All commands support --help with usage and at least one example
    When I run each kanban command with the help flag:
      | command        |
      | kanban init    |
      | kanban add     |
      | kanban board   |
      | kanban edit    |
      | kanban delete  |
      | kanban ci-done |
    Then each help output contains a usage description, available flags, and at least one example

  @skip
  Scenario: Version flag outputs the installed version string
    When I run "kanban --version"
    Then the exit code is 0
    And output contains a version string

  # ---------------------------------------------------------------------------
  # Commands run outside a git repository
  # Driving port: CLI (subprocess invocation)
  # ---------------------------------------------------------------------------

  @skip
  Scenario Outline: All commands exit with code 1 outside a git repository
    Given the current directory is not a git repository
    When I run <command>
    Then the exit code is 1
    And output contains "Not a git repository"

    Examples:
      | command                    |
      | kanban init                |
      | kanban add "Test task"     |
      | kanban board               |
      | kanban edit TASK-001       |
      | kanban delete TASK-001     |

  # ---------------------------------------------------------------------------
  # Task file format integrity
  # Driving port: CLI (kanban add, then file content assertions)
  # ---------------------------------------------------------------------------

  @skip
  Scenario: Created task file contains valid YAML front matter
    Given the repository is initialised with kanban
    When I run "kanban add" with title "Format validation task"
    Then the created task file begins with a YAML front matter block
    And the front matter contains the fields: id, title, status, priority, due, assignee

  @skip
  Scenario: Status field in task file contains only valid status values
    Given the repository is initialised with kanban
    And a task "Status field test" exists with status "todo"
    When I run the commit hook referencing that task
    Then the task file status field contains "in-progress"
    And the file remains parseable as valid YAML front matter

  @skip
  Scenario: Task file written atomically -- no partial writes visible
    Given the repository is initialised with kanban
    When I run "kanban add" with title "Atomic write test"
    Then the task file is complete and parseable immediately after the command exits
    And no partial or temporary files remain in the tasks directory

  # ---------------------------------------------------------------------------
  # Error message quality
  # Driving port: CLI (subprocess invocation)
  # ---------------------------------------------------------------------------

  @skip
  Scenario Outline: Error messages explain what happened, why, and what to do next
    Given the repository is initialised with kanban
    When I trigger the error condition for <scenario>
    Then the error message states what happened
    And the error message states why it happened
    And the error message states what to do next

    Examples:
      | scenario                              |
      | running add with no title             |
      | running add with a past due date      |
      | running edit on a missing task        |
      | running delete on a missing task      |
      | running any command outside a git repo |

  # ---------------------------------------------------------------------------
  # CI/CD integration smoke tests
  # Driving port: CIPipelineAdapter (kanban ci-done) and CLI (kanban board)
  # ---------------------------------------------------------------------------

  @skip
  Scenario: CI pipeline can build the binary and run acceptance tests successfully
    Given the kanban source code is present
    When the CI pipeline builds the binary
    Then the binary is produced without errors
    And the binary responds to "kanban --version" with exit code 0

  @skip
  Scenario: Full pipeline smoke test: init, add, hook transition, CI done, board verify
    Given a clean git repository with kanban installed
    When the CI pipeline executes the following sequence:
      | step                                              |
      | kanban init                                       |
      | kanban add "Pipeline smoke test task"             |
      | commit with message referencing the task ID       |
      | kanban ci-done after simulated test pass          |
    Then "kanban board" shows the task under DONE
    And all steps exit with code 0

  @property @skip
  Scenario: Task file round-trip preserves all field values
    Given any task created with valid title, priority, due date, and assignee
    When the task file is read back by "kanban board"
    Then all field values match what was provided at creation time exactly

  @property @skip
  Scenario: Board output is stable across repeated reads of the same task files
    Given any set of task files in the tasks directory
    When "kanban board" is run multiple times without modifying any files
    Then the output is identical on every run
