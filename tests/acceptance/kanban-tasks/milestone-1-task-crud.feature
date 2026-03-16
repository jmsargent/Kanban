Feature: Task creation, viewing, editing, and deletion
  As a developer managing work in the terminal
  I want to create, view, edit, and remove tasks
  So that the kanban board reflects accurate, up-to-date work items

  # US-01, US-02, US-03, US-06, US-07

  Background:
    Given I am working in a git repository

  # ---------------------------------------------------------------------------
  # US-01: Repository Initialisation
  # Driving port: CLI (kanban init)
  # ---------------------------------------------------------------------------

  Scenario: Developer sets up kanban in a new git repository
    Given the repository has no kanban setup
    When I run "kanban init"
    Then the kanban workspace directory is created at ".kanban/tasks/"
    And the configuration file is created with default task pattern and column list
    And the hook log file path is added to ".gitignore"
    And output confirms "Initialised kanban at .kanban/"
    And the exit code is 0

  @skip
  Scenario: Running init a second time makes no changes
    Given the repository is initialised with kanban
    When I run "kanban init" again
    Then no existing configuration files are modified
    And output shows "Already initialised at .kanban/ -- no changes made."
    And the exit code is 0

  @skip
  Scenario: Developer cannot initialise kanban outside a git repository
    Given the current directory is not a git repository
    When I run "kanban init"
    Then the exit code is 1
    And output contains "Not a git repository"

  # ---------------------------------------------------------------------------
  # US-02: Create Task
  # Driving port: CLI (kanban new)
  # ---------------------------------------------------------------------------

  @skip
  Scenario: Developer creates a task with title only
    Given the repository is initialised with kanban
    And no tasks exist yet
    When I run "kanban new" with title "Fix OAuth login bug"
    Then a task file "TASK-001.md" is created in the tasks directory
    And the task file records status "todo" and title "Fix OAuth login bug"
    And output shows "Created" with the task ID "TASK-001" and title
    And output contains a commit tip referencing "TASK-001"
    And the exit code is 0

  @skip
  Scenario: Developer creates a task with all optional fields
    Given the repository is initialised with kanban
    When I run "kanban new" with title "API rate limiting" and priority "P1" and due date "2026-03-20" and assignee "Alex Kim"
    Then the task file records priority "P1", due date "2026-03-20", and assignee "Alex Kim"
    And "kanban board" shows all four fields on that task line
    And the exit code is 0

  @skip
  Scenario: Task IDs increment sequentially
    Given the repository is initialised with kanban
    And a task "First task" exists as "TASK-001"
    When I run "kanban new" with title "Second task"
    Then a task file "TASK-002.md" is created
    And output shows "Created" with task ID "TASK-002"

  @skip
  Scenario: Creating a task fails when no title is provided
    Given the repository is initialised with kanban
    When I run "kanban new" with an empty title
    Then the exit code is 2
    And output contains "Task title is required"
    And no new task file is created in the tasks directory

  @skip
  Scenario: Creating a task fails when the due date is in the past
    Given the repository is initialised with kanban
    When I run "kanban new" with title "Old work" and due date "2025-01-01"
    Then the exit code is 2
    And output contains "Due date must be today or in the future"
    And no new task file is created in the tasks directory

  @skip
  Scenario: Creating a task fails outside a git repository
    Given the current directory is not a git repository
    When I run "kanban new" with title "Should not work"
    Then the exit code is 1
    And output contains "Not a git repository"

  # ---------------------------------------------------------------------------
  # US-03: View Board
  # Driving port: CLI (kanban board)
  # ---------------------------------------------------------------------------

  @skip
  Scenario: Developer views the board with tasks in all three statuses
    Given the repository is initialised with kanban
    And a task "Write tests" exists with status "todo"
    And a task "Fix login bug" exists with status "in-progress"
    And a task "Deploy to staging" exists with status "done"
    When I run "kanban board"
    Then output groups tasks under headings TODO, IN PROGRESS, and DONE
    And each heading shows the count of tasks in that group
    And each task line shows ID, title, priority, due date, and assignee
    And the exit code is 0

  @skip
  Scenario: Board shows "--" for missing priority and due, "unassigned" for missing assignee
    Given the repository is initialised with kanban
    And a task "No metadata task" exists with status "todo" and no priority, due date, or assignee
    When I run "kanban board"
    Then the task line shows "--" for priority and due date
    And the task line shows "unassigned" for assignee

  @skip
  Scenario: Overdue task shows a distinct indicator on the board
    Given the repository is initialised with kanban
    And a task "Expired work" exists with due date "2026-03-10" and status "todo"
    When I run "kanban board"
    Then the task line for "Expired work" shows a distinct overdue indicator

  @skip
  Scenario: Empty board shows onboarding message
    Given the repository is initialised with kanban
    And no tasks exist yet
    When I run "kanban board"
    Then output contains "No tasks found in .kanban/tasks/"
    And output contains a suggestion to run "kanban new"

  @skip
  Scenario: Board outputs valid machine-readable format when requested
    Given the repository is initialised with kanban
    And a task "API work" exists with status "in-progress"
    When I run "kanban board" with the machine output flag
    Then output is valid JSON
    And the JSON array contains an object with fields id, title, status, priority, due, and assignee
    And the exit code is 0

  @skip
  Scenario: Board suppresses colour codes when NO_COLOR is set in the environment
    Given the repository is initialised with kanban
    And a task "Color test" exists with status "todo"
    And the environment variable "NO_COLOR" is set
    When I run "kanban board"
    Then output contains no ANSI colour escape sequences

  @skip
  Scenario: Board produces plain output when piped to another command
    Given the repository is initialised with kanban
    And a task "Piped output" exists with status "todo"
    When I run "kanban board" with output piped to another process
    Then output contains no ANSI colour escape sequences
    And output contains no spinner characters

  # ---------------------------------------------------------------------------
  # US-06: Edit Task
  # Driving port: CLI (kanban edit)
  # ---------------------------------------------------------------------------

  @skip
  Scenario: Developer adds a description to an existing task
    Given the repository is initialised with kanban
    And a task "Fix OAuth login bug" exists with no description
    When I run "kanban edit" on that task and add a description in the editor
    Then output confirms "Updated: description"
    And the task file contains the new description

  @skip
  Scenario: Edit displays all current field values before opening the editor
    Given the repository is initialised with kanban
    And a task exists with priority "P2" and due date "2026-03-18"
    When I run "kanban edit" on that task
    Then output displays all current field values before the editor opens

  @skip
  Scenario: Editing the title updates the board display
    Given the repository is initialised with kanban
    And a task "Fix OAuth bug" exists with status "todo"
    When I run "kanban edit" on that task and change the title to "Fix OAuth login -- Chrome and Firefox"
    Then "kanban board" shows "Fix OAuth login -- Chrome and Firefox" for that task

  @skip
  Scenario: Edit with no changes made reports no update
    Given the repository is initialised with kanban
    And a task "Unchanged task" exists
    When I run "kanban edit" on that task and save without making any changes
    Then output shows "No changes made."
    And the exit code is 0

  @skip
  Scenario: Editing a non-existent task reports a clear error
    Given the repository is initialised with kanban
    When I run "kanban edit" on task "TASK-099"
    Then the exit code is 1
    And output contains "Task TASK-099 not found"
    And output contains a suggestion to run "kanban board"

  # ---------------------------------------------------------------------------
  # US-07: Delete Task
  # Driving port: CLI (kanban delete)
  # ---------------------------------------------------------------------------

  @skip
  Scenario: Developer deletes a task after confirming
    Given the repository is initialised with kanban
    And a task "Write unit tests auth module" exists with status "todo" as "TASK-002"
    When I run "kanban delete" on "TASK-002" and enter "y" at the confirmation prompt
    Then the task file for "TASK-002" is removed
    And output contains "Deleted" with the task ID and title
    And output contains a suggested git commit command
    And "kanban board" no longer lists "TASK-002"
    And the exit code is 0

  @skip
  Scenario: Developer aborts a delete by entering "n"
    Given the repository is initialised with kanban
    And a task "Keep this task" exists as "TASK-001"
    When I run "kanban delete" on "TASK-001" and enter "n" at the confirmation prompt
    Then the task file for "TASK-001" still exists
    And output contains "Aborted. TASK-001 unchanged."
    And the exit code is 0

  @skip
  Scenario: Developer aborts a delete by pressing Enter without input
    Given the repository is initialised with kanban
    And a task "Safe task" exists as "TASK-001"
    When I run "kanban delete" on "TASK-001" and press Enter without typing
    Then the task file for "TASK-001" still exists
    And output contains "Aborted. TASK-001 unchanged."

  @skip
  Scenario: Force delete removes a task without prompting
    Given the repository is initialised with kanban
    And a task "Script cleanup task" exists as "TASK-005"
    When I run "kanban delete" on "TASK-005" with the force flag
    Then the task file for "TASK-005" is removed immediately
    And the exit code is 0

  @skip
  Scenario: Deleting a non-existent task reports a clear error
    Given the repository is initialised with kanban
    When I run "kanban delete" on task "TASK-099"
    Then the exit code is 1
    And output contains "Task TASK-099 not found"

  @skip
  Scenario: Kanban does not auto-commit after task deletion
    Given the repository is initialised with kanban
    And a task "To be removed" exists as "TASK-003"
    When I run "kanban delete" on "TASK-003" and enter "y"
    Then the git repository has no new commits from the delete operation
    And output contains a git commit command suggestion for the developer to run
