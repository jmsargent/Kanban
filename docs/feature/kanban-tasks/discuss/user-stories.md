<!-- markdownlint-disable MD024 -->
# User Stories: kanban-tasks

Feature: Git-Native Kanban Task Management CLI
Date: 2026-03-15

---

## US-01: Repository Initialisation

### Problem

Rafael is on a new team and wants to start using kanban in their shared repo. There is no
.kanban/ directory yet. Running any kanban command fails with a confusing error. He finds it
frustrating to manually create directories and config files before he can do anything useful.

### Who

- Backend developer | Working in an existing git repository | Wants to start tracking work

### Solution

A single command that sets up the .kanban/ directory structure with sensible defaults, ready
for the team to start adding tasks immediately.

### Domain Examples

#### 1: Fresh repository
Rafael has just run "git init" on a new project. He runs "kanban init" and gets .kanban/tasks/
created with a default config file that includes ci_task_pattern = TASK-[0-9]+ and column
definitions todo, in-progress, done.

#### 2: Already initialised
Priya clones the team repo where Rafael already ran kanban init. She runs "kanban init" and
gets "Already initialised at .kanban/ -- no changes made." Nothing is overwritten.

#### 3: Not a git repository
Rafael accidentally runs "kanban init" in his home directory (not a git repo). He gets:
"Error: Not a git repository. Run 'git init' first."

### UAT Scenarios (BDD)

#### Scenario: Initialise kanban in a new git repository
Given Rafael is in a directory with a git repository and no .kanban/ directory
When he runs "kanban init"
Then .kanban/tasks/ is created
And .kanban/config is created with default ci_task_pattern and column definitions
And .kanban/hook.log is added to .gitignore
And output confirms "Initialised kanban at .kanban/"
And exit code is 0

#### Scenario: Initialise is idempotent
Given .kanban/ already exists in the repository
When Rafael runs "kanban init" again
Then no existing files are modified
And output shows "Already initialised at .kanban/ -- no changes made."
And exit code is 0

#### Scenario: Initialise fails outside git repository
Given the current directory is not a git repository
When Rafael runs "kanban init"
Then exit code is 1
And output contains "Error: Not a git repository."

### Acceptance Criteria

- [ ] "kanban init" creates .kanban/tasks/ and .kanban/config in the repo root
- [ ] .kanban/config contains default ci_task_pattern and column list (todo, in-progress, done)
- [ ] .kanban/hook.log is added to .gitignore by init
- [ ] Running "kanban init" a second time produces no changes and a status message
- [ ] Running "kanban init" outside a git repo exits with code 1 and a clear error

### Outcome KPIs

- Who: developer on a new team
- Does what: completes kanban setup without reading documentation
- By how much: 100% of teams complete init in under 2 minutes
- Measured by: time from clone to first "kanban board" output
- Baseline: none (new feature)

### Technical Notes

- Must detect repo_root via "git rev-parse --show-toplevel"
- .kanban/ should be committed to the repo (shared state)
- Must not overwrite existing .kanban/config on re-init

---

## US-02: Create Task

### Problem

Rafael has identified a bug and needs to capture it as a task. He wants to add it without
leaving the terminal, without logging into a browser, and with the confidence that the task
will be tracked in the same history as the code fix.

### Who

- Backend developer | In terminal, mid-work session | Wants to capture work without context switch

### Solution

A "kanban add" command that creates a task file in .kanban/tasks/ with a generated ID,
initialised status of "todo", and optional metadata fields.

### Domain Examples

#### 1: Minimal creation
Rafael runs: kanban add "Fix OAuth login bug"
He sees TASK-001.md created in .kanban/tasks/ with status "todo" and a tip showing how to
reference the task ID in his next commit message.

#### 2: Full metadata
Before starting a P1 issue, Rafael runs:
kanban add "API rate limiting" --priority=P1 --due=2026-03-20 --assignee="Alex Kim"
The task file contains all four fields. "kanban board" shows the complete task line.

#### 3: Validation prevents past due date
Rafael makes a typo: kanban add "Old thing" --due=2025-01-01
He gets: "Error: Due date must be today or in the future." No file is created.

### UAT Scenarios (BDD)

#### Scenario: Create task with title only
Given .kanban/tasks/ exists and is empty
When Rafael runs: kanban add "Fix OAuth login bug"
Then .kanban/tasks/TASK-001.md is created
And the file contains status "todo" and title "Fix OAuth login bug"
And output shows "Created  TASK-001  Fix OAuth login bug"
And output contains a commit tip referencing TASK-001
And exit code is 0

#### Scenario: Create task with all optional fields
When Rafael runs: kanban add "API rate limiting" --priority=P1 --due=2026-03-20 --assignee="Alex Kim"
Then the task file contains priority P1, due 2026-03-20, assignee "Alex Kim"
And "kanban board" shows all four fields on the TASK line

#### Scenario: Task IDs increment
Given TASK-001.md exists
When Rafael runs: kanban add "Second task"
Then TASK-002.md is created and output shows "Created  TASK-002"

#### Scenario: Create fails with empty title
When Rafael runs: kanban add ""
Then exit code is 2
And output contains "Error: Task title is required."
And no file is created in .kanban/tasks/

#### Scenario: Create fails with past due date
When Rafael runs: kanban add "Old task" --due=2025-01-01
Then exit code is 2
And output contains "Error: Due date must be today or in the future."

### Acceptance Criteria

- [ ] "kanban add" with non-empty title creates .kanban/tasks/TASK-NNN.md with status "todo"
- [ ] Task IDs are unique and auto-increment (TASK-001, TASK-002, ...)
- [ ] Optional flags --priority, --due, --assignee are stored in the task file when provided
- [ ] Empty title exits with code 2 and a clear error; no file created
- [ ] Past due date exits with code 2 and a clear error; no file created
- [ ] Output always shows file path, ID, status, and a commit tip with the correct task ID

### Outcome KPIs

- Who: developer who identifies a new piece of work
- Does what: captures task in the terminal without context switch
- By how much: task creation takes under 10 seconds including typing the command
- Measured by: time-on-task observation; absence of "I forgot to add it to the board" in team retros
- Baseline: none (new feature)

### Technical Notes

- ID generation must be atomic (safe for two developers running "kanban add" near-simultaneously)
- File format left to DESIGN wave (markdown front matter, YAML, or TOML)
- Must validate that .kanban/tasks/ exists before writing; suggest "kanban init" if not

---

## US-03: View Board

### Problem

Priya wants to know what her teammates are working on before the standup. She does not want to
open a browser, log into Jira, or send a Slack message. She wants to type one command and see
the current state of the team's work.

### Who

- Developer (primary or secondary) | In terminal, start of day or before switching tasks | Wants current team status

### Solution

A "kanban board" command that reads all task files and displays them grouped by status with
clean, scannable terminal output.

### Domain Examples

#### 1: Typical board with 3 active tasks
Rafael runs "kanban board" before standup. He sees 2 tasks in TODO (including one with an
overdue indicator), 1 in IN PROGRESS (Alex Kim's API rate limiting), and 3 in DONE.
Each line shows ID, title, priority, due date, and assignee.

#### 2: Scripted JSON output
Priya has a shell script that processes task counts. She runs "kanban board --json" and gets
a JSON array of all tasks with all fields. She pipes it through jq.

#### 3: Empty board onboarding
Sam clones the repo and runs "kanban board". No tasks exist yet. He sees:
"No tasks found in .kanban/tasks/" with a suggestion to run "kanban add".

### UAT Scenarios (BDD)

#### Scenario: Board shows tasks grouped by status
Given tasks exist with statuses todo, in-progress, and done
When Rafael runs "kanban board"
Then tasks are grouped under headings TODO, IN PROGRESS, DONE
And each task line shows ID, title, priority, due date, assignee
And each heading shows the count of tasks in that group
And exit code is 0

#### Scenario: Overdue task shows visual indicator
Given TASK-001 has due date 2026-03-10 and today is 2026-03-15
When Rafael runs "kanban board"
Then TASK-001 line shows a distinct overdue indicator alongside the due date

#### Scenario: Empty board shows onboarding message
Given .kanban/tasks/ is empty
When Rafael runs "kanban board"
Then output contains "No tasks found in .kanban/tasks/"
And output contains a suggestion for "kanban add"

#### Scenario: JSON output for machine consumption
When Rafael runs "kanban board --json"
Then output is valid JSON containing an array of task objects
And each object has fields: id, title, status, priority, due, assignee
And exit code is 0

#### Scenario: NO_COLOR respected
Given environment variable NO_COLOR is set
When Rafael runs "kanban board"
Then output contains no ANSI color escape sequences

### Acceptance Criteria

- [ ] Output groups tasks under TODO, IN PROGRESS, DONE headings
- [ ] Each heading shows count of tasks in that group
- [ ] Each task line shows: ID, title, priority, due date, assignee (blank fields shown as "--" or "unassigned")
- [ ] Overdue tasks (today > due date) show a distinct indicator
- [ ] --json flag produces valid JSON array of task objects
- [ ] Empty board shows onboarding message with "kanban add" suggestion
- [ ] NO_COLOR env var disables all ANSI color codes
- [ ] First output within 100ms for repos with up to 500 task files

### Outcome KPIs

- Who: developer on a shared team
- Does what: checks team work status in the terminal without switching tools
- By how much: developer checks "kanban board" at least once per working day
- Measured by: shell history frequency (survey proxy); team reports reduced "what are you working on?" interruptions
- Baseline: none (new feature)

### Technical Notes

- Reads all .md files in .kanban/tasks/ on each invocation (no cache for MVP)
- --json schema is a versioned contract once published
- Non-TTY (piped) output must disable color and animations automatically
- Overdue calculation: UTC date comparison (no time component)

---

## US-04: Auto-move to In Progress (Commit Hook)

### Problem

Rafael makes his first commit on the "Fix OAuth login bug" task. He wants the board to show
the task as in-progress without making him type a separate command. Any extra step after
"git commit" means he might forget, and the board becomes stale immediately.

### Who

- Developer making first commit on a task | In terminal, coding session | Wants zero-effort board updates

### Solution

A git commit-msg hook that reads the commit message, extracts any task IDs, and updates the
matching task files from "todo" to "in-progress". Output appears inline in the git commit response.

### Domain Examples

#### 1: First commit auto-transitions
Rafael commits: git commit -m "TASK-001: reproduce OAuth bug on Chrome"
Below the standard git output he sees:
"kanban: TASK-001 moved  todo -> in-progress"
He runs "kanban board" and TASK-001 is under IN PROGRESS.

#### 2: Commit with unknown task ID warns but succeeds
Rafael types TASK-099 by mistake (task does not exist). His commit still goes through.
He sees: "Warning: TASK-099 not found in .kanban/tasks/ -- skipping"

#### 3: Commit with no task reference is invisible
Rafael commits a typo fix with no task reference. No kanban output appears. Clean commit log.

### UAT Scenarios (BDD)

#### Scenario: First commit referencing task moves it to in-progress
Given TASK-001 has status "todo"
When Rafael commits with message "TASK-001: reproduce OAuth bug"
Then .kanban/tasks/TASK-001.md status field is "in-progress"
And commit output contains "kanban: TASK-001 moved  todo -> in-progress"
And "kanban board" shows TASK-001 under IN PROGRESS
And git commit exit code is 0

#### Scenario: Commit with already in-progress task is a no-op
Given TASK-001 has status "in-progress"
When Rafael commits "TASK-001: more fixes"
Then TASK-001 status remains "in-progress"
And no transition output appears

#### Scenario: Commit with unknown task ID warns and does not block
Given no file exists for TASK-099
When Rafael commits "TASK-099: working on this"
Then git commit succeeds with exit code 0
And output contains "Warning: TASK-099 not found in .kanban/tasks/ -- skipping"

#### Scenario: Commit with no task reference produces no kanban output
When Rafael commits "fix typo in README"
Then no "kanban:" lines appear in commit output
And all task statuses are unchanged

### Acceptance Criteria

- [ ] Commit message matching ci_task_pattern triggers status update for matched task
- [ ] Status updated from "todo" to "in-progress" in task file
- [ ] Transition logged inline in git commit output
- [ ] Missing task ID logs warning; git commit is never blocked (hook always exits 0)
- [ ] Task already in-progress or done: no-op, no output
- [ ] Commit with no task reference: no kanban output at all

### Outcome KPIs

- Who: developer making first commit on a task
- Does what: task status updates to in-progress without a manual command
- By how much: 90% of in-progress transitions are automatic (not from "kanban edit")
- Measured by: ratio of auto-transitions to manual status edits in .kanban/tasks/ commit history
- Baseline: none (new feature)

### Technical Notes

- Delivered as a commit-msg hook: .git/hooks/commit-msg
- Must be installed via "kanban install-hook" (not manually; reduces friction)
- Hook must read ci_task_pattern from .kanban/config (not hardcoded)
- Hook failure must exit 0 and log error to .kanban/hook.log
- Hook must complete within 500ms to avoid perceptibly slowing git commit

---

## US-05: Auto-move to Done (CI Pipeline)

### Problem

Alex has finished implementing API rate limiting. The tests pass in CI. He wants the task to
move to "done" automatically -- the passing tests are the definition of done, so no human
should have to move the card. A manual step here is exactly the kind of thing that makes
boards go stale.

### Who

- Developer whose CI pipeline just passed | In CI/CD environment | Wants done to mean done

### Solution

A CI step (script) that runs after all tests pass, scans commit messages in the pipeline run
for task IDs, and updates matching task files to "done". The CI step commits the updated files
back to the repo.

### Domain Examples

#### 1: Single task completed via CI
CI runs on Alex's push. Tests pass. Commit "TASK-003: implement throttle middleware -- tests green"
is in the pipeline. CI log shows:
"[kanban] TASK-003 moved  in-progress -> done"
After the CI commit, "kanban board" shows TASK-003 under DONE.

#### 2: Multiple tasks in one pipeline run
Rafael and Alex both have tasks referenced in the same pipeline. Tests pass.
CI moves TASK-001 and TASK-003 to done in the same run.

#### 3: Pipeline fails -- no transition
CI runs but two tests fail. Exit code is non-zero.
No kanban transitions fire. TASK-001 stays "in-progress".

### UAT Scenarios (BDD)

#### Scenario: Pipeline passes and auto-moves task to done
Given TASK-003 has status "in-progress"
And the pipeline run includes a commit with "TASK-003" in the message
When all CI tests pass (exit code 0)
Then .kanban/tasks/TASK-003.md status is "done"
And CI log contains "[kanban] TASK-003 moved  in-progress -> done"
And the updated task file is committed back to the repo

#### Scenario: Pipeline fails -- no status transition
Given TASK-003 has status "in-progress"
When CI runs and one or more tests fail
Then TASK-003 status remains "in-progress"
And no transition output appears in CI log

#### Scenario: Multiple tasks moved in one pipeline run
Given TASK-001 and TASK-003 are both "in-progress"
When pipeline passes with commits referencing both
Then both are updated to "done"
And CI log shows a transition line for each

#### Scenario: Already-done task referenced in passing pipeline is a no-op
Given TASK-001 has status "done"
When pipeline passes and a commit references TASK-001
Then TASK-001 status remains "done"
And no transition output appears for TASK-001

### Acceptance Criteria

- [ ] CI step runs only after test suite exits with code 0
- [ ] Scans all commit messages in the pipeline run for ci_task_pattern matches
- [ ] Updates status from "in-progress" to "done" in matched task files
- [ ] Commits updated task files back to repo (so team sees update on next pull)
- [ ] Pipeline test failure: no transitions, no file modifications
- [ ] Already-done task: no-op, no output
- [ ] Multiple tasks: each transition logged individually

### Outcome KPIs

- Who: developer whose work just passed CI
- Does what: task moves to done without a manual step
- By how much: 95% of done transitions are automatic (not from "kanban edit")
- Measured by: ratio of auto-transitions to manual status edits for "done" in commit history
- Baseline: none (new feature)

### Technical Notes

- Deliverable as a standalone shell script runnable on any CI platform
- R3 delivers platform-specific wrappers (GitHub Actions, GitLab CI)
- CI step must run in non-TTY environment (no interactive prompts, no color)
- Must read ci_task_pattern from .kanban/config in the checked-out repo
- Must handle "git add .kanban/tasks/ && git commit" cleanly (no merge conflicts expected; only status field changes)

---

## US-06: Edit Task

### Problem

Rafael created TASK-001 without a description. A day later he realises the description would
help Alex understand the reproduction steps. He wants to add it without creating a new task
or deleting and recreating. He also sometimes makes typos in titles that need fixing.

### Who

- Developer | Has an existing task that needs updated metadata | Wants to correct without ceremony

### Solution

A "kanban edit" command that displays current field values and opens the task file in $EDITOR.
After save and exit, confirms which fields changed.

### Domain Examples

#### 1: Add description after creation
Rafael runs: kanban edit TASK-001
He sees current fields including empty description. Editor opens. He adds:
"Reproduces on Chrome and Firefox when using Google OAuth. Token exchange fails intermittently."
He saves. Output: "Updated: description"

#### 2: Fix a title typo
Rafael notices he misspelled "implementaiton" in TASK-003's title.
He runs kanban edit TASK-003, fixes the typo, saves.
Output: "Updated: title" -- "kanban board" shows the corrected title.

#### 3: Task not found
Rafael misremembers the ID: kanban edit TASK-099
He gets: "Error: Task TASK-099 not found." with a suggestion to run "kanban board" to see IDs.

### UAT Scenarios (BDD)

#### Scenario: Edit adds description to existing task
Given TASK-001 has empty description
When Rafael runs "kanban edit TASK-001" and adds a description in the editor and saves
Then output contains "Updated: description"
And .kanban/tasks/TASK-001.md contains the new description
And "kanban board" reflects no visual change (description not shown in summary)

#### Scenario: Edit shows current values before opening editor
Given TASK-001 has priority P2 and due 2026-03-18
When Rafael runs "kanban edit TASK-001"
Then output displays all current field values before the editor prompt

#### Scenario: Edit title updates board display
When Rafael edits TASK-001 and changes title to "Fix OAuth login -- Chrome and Firefox"
Then "kanban board" shows the new title for TASK-001

#### Scenario: Edit non-existent task
When Rafael runs "kanban edit TASK-099"
Then exit code is 1
And output contains "Error: Task TASK-099 not found."
And output contains suggestion to run "kanban board"

### Acceptance Criteria

- [ ] "kanban edit TASK-NNN" displays current field values before opening editor
- [ ] Opens task file in $EDITOR (falls back to vi if $EDITOR unset)
- [ ] After save and exit, output lists which fields changed
- [ ] If no fields changed, output shows "No changes made."
- [ ] Non-existent task ID exits with code 1 and actionable error message

### Outcome KPIs

- Who: developer correcting task metadata
- Does what: updates task without leaving the terminal
- By how much: 100% of edit operations complete without switching to a browser or GUI
- Measured by: no external tool usage for task edits (team survey / retro)
- Baseline: none (new feature)

### Technical Notes

- Uses $EDITOR environment variable; documents fallback to vi
- Must not corrupt the task file if editor exits without saving (compare file mtime)
- status field can be overridden via edit; this is the escape hatch for out-of-band corrections

---

## US-07: Delete Task

### Problem

A task was added for a feature that was descoped mid-sprint. It should not appear on the board
as "todo" forever. Rafael needs to remove it cleanly, with the removal tracked in git history.
He does not want accidental deletions -- a confirmation step protects against fat-finger errors.

### Who

- Developer | Has a cancelled or erroneous task | Wants to clean up without accidental data loss

### Solution

A "kanban delete" command that shows a confirmation prompt (with task title for recognition),
removes the file on confirm, and suggests the git commit command to share the deletion with the team.

### Domain Examples

#### 1: Delete cancelled task
Rafael runs: kanban delete TASK-002
He sees: "Delete  TASK-002  Write unit tests auth module  [todo]?"
He confirms. File is removed. Output suggests: git add -A && git commit -m "Remove TASK-002"

#### 2: Abort accidental delete
Priya accidentally types kanban delete TASK-001 instead of TASK-002.
She sees the task title in the prompt, recognises the mistake, types "n".
Nothing changes.

#### 3: Delete in a script using --force
A CI cleanup script runs: kanban delete TASK-005 --force
No prompt is shown. File is removed immediately. Exit code 0.

### UAT Scenarios (BDD)

#### Scenario: Delete task with confirmation
Given .kanban/tasks/TASK-002.md exists with title "Write unit tests auth module"
When Rafael runs "kanban delete TASK-002" and enters "y"
Then .kanban/tasks/TASK-002.md is removed
And output contains "Deleted  TASK-002  Write unit tests auth module"
And output contains a git commit command suggestion
And "kanban board" no longer lists TASK-002
And exit code is 0

#### Scenario: Abort delete
When Rafael runs "kanban delete TASK-002" and enters "n"
Then .kanban/tasks/TASK-002.md still exists
And output contains "Aborted. TASK-002 unchanged."
And exit code is 0

#### Scenario: Force delete skips confirmation
When Rafael runs "kanban delete TASK-002 --force"
Then .kanban/tasks/TASK-002.md is removed immediately without prompting
And exit code is 0

#### Scenario: Delete non-existent task
When Rafael runs "kanban delete TASK-099"
Then exit code is 1
And output contains "Error: Task TASK-099 not found."

### Acceptance Criteria

- [ ] Confirmation prompt shows task title and current status before deletion
- [ ] "y" (or "yes") removes task file; "n" (or Enter) aborts with no change
- [ ] --force flag removes file without prompting
- [ ] Deletion output includes a git commit command suggestion
- [ ] Non-existent task ID exits with code 1 and actionable error
- [ ] kanban does not auto-commit; developer owns the commit

### Outcome KPIs

- Who: developer removing a cancelled or erroneous task
- Does what: removes task from board cleanly with git history
- By how much: zero accidental deletions (no user reports of "I deleted the wrong task")
- Measured by: user feedback; absence of "restore deleted task" requests
- Baseline: none (new feature)

### Technical Notes

- Must read task file immediately before displaying confirmation (no stale title)
- Default prompt answer is "N" (safe default -- abort on Enter without input)
- git commit is the developer's responsibility; kanban only removes the file
