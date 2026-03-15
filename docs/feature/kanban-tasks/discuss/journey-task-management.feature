Feature: Kanban Task Management (Git-Native CLI)
  # Platform: CLI tool with CI/CD integration hooks
  # Persona: Rafael Rodrigues -- backend developer, 4-person team, shared git repo
  # Mental model: tasks are files in the repo; state transitions are git events and CI outcomes
  # Key CLI principles: exit codes 0/1/2, --help on every command, NO_COLOR respected

  Background:
    Given Rafael is in a directory that is a git repository
    And .kanban/tasks/ exists (created by "kanban init")
    And .kanban/config exists with default ci_task_pattern "TASK-[0-9]+"

  # ---------------------------------------------------------------------------
  # Step 1: View Board
  # ---------------------------------------------------------------------------

  Scenario: View board with tasks in all statuses
    Given task files exist in .kanban/tasks/ with the following data:
      | ID       | Title                          | Status      | Priority | Due        | Assignee          |
      | TASK-001 | Fix OAuth login bug            | todo        | P2       | 2026-03-18 | Rafael Rodrigues  |
      | TASK-003 | API rate limiting               | in-progress | P1       | 2026-03-20 | Alex Kim          |
      | TASK-006 | Deploy v1.2 to staging         | done        | P1       | 2026-03-14 | Sam Torres        |
    When Rafael runs "kanban board"
    Then the output groups tasks under headings "TODO", "IN PROGRESS", "DONE"
    And TASK-001 appears under TODO with its title, priority, due date, and assignee
    And TASK-003 appears under IN PROGRESS
    And TASK-006 appears under DONE
    And the TODO heading shows count "1"
    And the IN PROGRESS heading shows count "1"
    And the DONE heading shows count "1"
    And exit code is 0

  Scenario: View empty board shows onboarding tip
    Given .kanban/tasks/ contains no files
    When Rafael runs "kanban board"
    Then output contains "No tasks found in .kanban/tasks/"
    And output contains "kanban add"
    And exit code is 0

  Scenario: View board outside git repository
    Given the current directory is not a git repository
    When Rafael runs "kanban board"
    Then exit code is 1
    And output contains "Error: Not a git repository."
    And output contains "git init"

  Scenario: Board output in JSON format for scripting
    Given tasks exist in .kanban/tasks/
    When Rafael runs "kanban board --json"
    Then output is valid JSON
    And each task object contains fields: id, title, status, priority, due, assignee
    And exit code is 0

  # ---------------------------------------------------------------------------
  # Step 2: Create Task
  # ---------------------------------------------------------------------------

  Scenario: Create task with title only
    Given .kanban/tasks/ contains no files
    When Rafael runs: kanban add "Fix OAuth login bug"
    Then .kanban/tasks/TASK-001.md is created
    And the file contains status "todo"
    And the file contains the title "Fix OAuth login bug"
    And output shows "Created  TASK-001  Fix OAuth login bug"
    And output shows the file path ".kanban/tasks/TASK-001.md"
    And output contains a tip referencing "TASK-001" for the commit message format
    And exit code is 0

  Scenario: Create task with all optional fields
    When Rafael runs: kanban add "Fix OAuth login bug" --priority=P2 --due=2026-03-18 --assignee="Rafael Rodrigues"
    Then .kanban/tasks/TASK-001.md contains priority "P2", due "2026-03-18", assignee "Rafael Rodrigues"
    And "kanban board" shows TASK-001 with all four fields on the task line

  Scenario: Task IDs increment sequentially
    Given TASK-001.md already exists in .kanban/tasks/
    When Rafael runs: kanban add "Second task"
    Then .kanban/tasks/TASK-002.md is created
    And output shows "Created  TASK-002"

  Scenario: Create task fails with empty title
    When Rafael runs: kanban add ""
    Then exit code is 2
    And output contains "Error: Task title is required."
    And no new file is created in .kanban/tasks/

  Scenario: Create task fails with due date in the past
    When Rafael runs: kanban add "Backfill data" --due=2025-01-01
    Then exit code is 2
    And output contains "Error: Due date must be today or in the future."
    And no new file is created

  Scenario: Create task outside git repository
    Given the current directory is not a git repository
    When Rafael runs: kanban add "anything"
    Then exit code is 1
    And output contains "Error: Not a git repository."

  # ---------------------------------------------------------------------------
  # Step 3: Auto-move to In Progress (commit hook)
  # ---------------------------------------------------------------------------

  Scenario: First commit referencing task ID auto-moves task to in-progress
    Given .kanban/tasks/TASK-001.md has status "todo"
    When Rafael runs: git commit -m "TASK-001: reproduce OAuth bug on Chrome and Firefox"
    Then .kanban/tasks/TASK-001.md status field is "in-progress"
    And commit output contains "kanban: TASK-001 moved  todo -> in-progress"
    And the git commit itself succeeds with exit code 0

  Scenario: Commit with no task ID reference produces no kanban output
    When Rafael runs: git commit -m "fix typo in README"
    Then no "kanban:" lines appear in commit output
    And all task statuses are unchanged

  Scenario: Commit referencing unknown task ID warns and does not block
    Given no file exists for TASK-099
    When Rafael runs: git commit -m "TASK-099: working on this"
    Then the commit succeeds with exit code 0
    And output contains "Warning: TASK-099 not found in .kanban/tasks/ -- skipping"

  Scenario: Commit referencing already in-progress task is a no-op
    Given TASK-001 has status "in-progress"
    When Rafael commits with message "TASK-001: add more fixes"
    Then TASK-001 status remains "in-progress"
    And no transition output appears

  # ---------------------------------------------------------------------------
  # Step 4: Auto-move to Done (CI pipeline)
  # ---------------------------------------------------------------------------

  Scenario: CI pipeline passes and auto-moves referenced task to done
    Given .kanban/tasks/TASK-001.md has status "in-progress"
    And the pipeline run includes commit 9ab34e1 with message "TASK-001: final fix"
    When all CI tests pass
    Then .kanban/tasks/TASK-001.md status field is "done"
    And CI log contains "TASK-001 moved  in-progress -> done"

  Scenario: CI pipeline fails -- no status transition occurs
    Given .kanban/tasks/TASK-001.md has status "in-progress"
    When the CI pipeline runs and one or more tests fail
    Then TASK-001 status remains "in-progress"
    And CI log contains no transition output for TASK-001

  Scenario: CI run references multiple tasks and all are moved to done
    Given TASK-001 and TASK-002 both have status "in-progress"
    And commits in this pipeline run reference both TASK-001 and TASK-002
    When the pipeline passes
    Then TASK-001 status is "done"
    And TASK-002 status is "done"
    And CI log shows individual transition lines for each task

  Scenario: CI moves task from done is a no-op
    Given TASK-001 has status "done"
    When the pipeline passes and a commit references TASK-001
    Then TASK-001 status remains "done"
    And no transition output appears for TASK-001

  # ---------------------------------------------------------------------------
  # Step 5: Edit Task
  # ---------------------------------------------------------------------------

  Scenario: Edit task opens file in $EDITOR and confirms changed fields
    Given TASK-001 exists with empty description and priority P2
    When Rafael runs "kanban edit TASK-001" and adds a description in the editor and saves
    Then output contains "Updated: description"
    And .kanban/tasks/TASK-001.md contains the new description
    And "kanban board" shows the updated task

  Scenario: Edit task title
    Given TASK-001 has title "Fix OAuth login bug"
    When Rafael edits TASK-001 and changes the title to "Fix OAuth login -- Chrome and Firefox"
    Then .kanban/tasks/TASK-001.md title field is "Fix OAuth login -- Chrome and Firefox"
    And "kanban board" shows the new title for TASK-001

  Scenario: Edit shows current field values before opening editor
    Given TASK-001 has priority P2 and due date 2026-03-18
    When Rafael runs "kanban edit TASK-001"
    Then output displays the current field values before the editor prompt

  Scenario: Edit non-existent task
    When Rafael runs "kanban edit TASK-099"
    Then exit code is 1
    And output contains "Error: Task TASK-099 not found."

  # ---------------------------------------------------------------------------
  # Step 6: Delete Task
  # ---------------------------------------------------------------------------

  Scenario: Delete task with explicit confirmation
    Given .kanban/tasks/TASK-001.md exists with title "Fix OAuth login bug"
    When Rafael runs "kanban delete TASK-001" and enters "y" at the prompt
    Then .kanban/tasks/TASK-001.md is removed from the filesystem
    And output contains "Deleted  TASK-001  Fix OAuth login bug"
    And output contains a git commit suggestion
    And "kanban board" no longer lists TASK-001
    And exit code is 0

  Scenario: Delete task -- user aborts at confirmation
    Given .kanban/tasks/TASK-001.md exists
    When Rafael runs "kanban delete TASK-001" and enters "n" at the prompt
    Then .kanban/tasks/TASK-001.md still exists
    And output contains "Aborted. TASK-001 unchanged."
    And exit code is 0

  Scenario: Delete task with --force skips confirmation (for scripting)
    When Rafael runs "kanban delete TASK-001 --force"
    Then .kanban/tasks/TASK-001.md is removed without prompting
    And exit code is 0

  Scenario: Delete non-existent task
    When Rafael runs "kanban delete TASK-099"
    Then exit code is 1
    And output contains "Error: Task TASK-099 not found."

  # ---------------------------------------------------------------------------
  # Error and edge cases
  # ---------------------------------------------------------------------------

  Scenario: All commands support --help
    When Rafael runs any kanban subcommand with --help
    Then output contains usage, flags, and at least one example
    And exit code is 0

  Scenario: Output respects NO_COLOR environment variable
    Given environment variable NO_COLOR is set
    When Rafael runs "kanban board"
    Then output contains no ANSI color codes

  Scenario: Output in non-TTY (piped) disables color and animations
    When Rafael pipes "kanban board" output to another command
    Then output contains no ANSI color codes and no spinner characters

  Scenario: Overdue task shows visual indicator on board
    Given TASK-001 has due date "2026-03-10" and status "todo"
    And today is 2026-03-15
    When Rafael runs "kanban board"
    Then TASK-001 line includes an overdue indicator distinct from the due date label
