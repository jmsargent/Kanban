# Story Map: kanban-tasks

User: Rafael Rodrigues -- backend developer on a small team sharing a git repository
Goal: Track team work without leaving the terminal or maintaining a separate tool
Date: 2026-03-15

---

## Backbone (User Activities -- left to right, chronological)

| Create | View | Work | Complete | Maintain |
|--------|------|------|----------|----------|
| Add a task to the board | See current board state | Start working on a task | Finish work and close task | Correct or remove tasks |

---

## Story Map

| Create | View | Work | Complete | Maintain |
|--------|------|------|----------|---------|
| kanban add (title only) | kanban board (all statuses) | Commit triggers in-progress | CI pass triggers done | kanban edit (open in $EDITOR) |
| kanban add (with priority, due, assignee) | kanban board --json (machine output) | Commit with unknown ID warns, not blocks | CI fail leaves status unchanged | kanban delete (with confirmation) |
| kanban init (set up .kanban/ in repo) | kanban board --filter (by status/assignee) | Multiple tasks in one commit | Multiple tasks in one CI run | kanban delete --force (no prompt, for scripts) |
| Task ID auto-increment | Overdue task indicator | Hook installed via kanban install-hook | CI step config (GitHub Actions / GitLab) | kanban show TASK-NNN (detailed view) |
| kanban add --dry-run (preview) | kanban list (flat list for scripts) | | | |

---

## Walking Skeleton

The thinnest end-to-end slice that proves the system works at all.
Every backbone activity is represented. All five activities covered.

```
Create              View                Work                Complete           Maintain
-----------         -----------         -----------         -----------        -----------
kanban add          kanban board        commit hook moves   CI step moves      kanban edit
(title only)    ->  (all statuses)  ->  todo->in-progress ->in-progress->done->($EDITOR)
                                                                               kanban delete
                                                                               (with confirm)
```

Walking Skeleton stories:
- WS-1: kanban init -- initialise .kanban/ in a git repo
- WS-2: kanban add "title" -- create task file with status "todo"
- WS-3: kanban board -- read all task files and display grouped by status
- WS-4: commit hook -- auto-move task to in-progress on first referencing commit
- WS-5: CI step -- auto-move task to done when pipeline passes
- WS-6: kanban edit -- open task file in $EDITOR and confirm changes
- WS-7: kanban delete -- remove task file with confirmation prompt

The walking skeleton delivers the complete developer loop: capture work, see it, do it (auto-updates), finish it (auto-updates), fix mistakes. A team can run this end-to-end on day one.

---

## Release Slices by Outcome

### Walking Skeleton: "A developer can track real work end-to-end in the terminal"

Outcome: developers on a shared repo complete the full loop (add -> view -> auto-progress -> auto-done -> edit -> delete) without leaving the terminal or using an external tool.

Stories:
- WS-1: kanban init
- WS-2: kanban add (title only)
- WS-3: kanban board (grouped output)
- WS-4: commit hook auto-moves to in-progress
- WS-5: CI step auto-moves to done
- WS-6: kanban edit ($EDITOR)
- WS-7: kanban delete (with confirmation)

---

### Release 1: "The board is useful as a daily reference"

Outcome: developers check "kanban board" at the start of their day and get accurate, scannable information without additional filtering or context-switching.

Adds to walking skeleton:
- R1-1: kanban add with optional flags (--priority, --due, --assignee)
- R1-2: Overdue indicator on board (red/distinct label when today > due date)
- R1-3: kanban board --json (machine-parseable output for scripting)
- R1-4: Error messages: clear and actionable for all invalid inputs (empty title, past due, not-a-repo)

---

### Release 2: "The board scales to a busy team without noise"

Outcome: on a team with 10+ active tasks, developers find their own tasks and priorities in under 10 seconds without scrolling through irrelevant work.

Adds:
- R2-1: kanban board --filter status=in-progress (filter by status)
- R2-2: kanban board --filter assignee="Rafael Rodrigues" (filter by assignee)
- R2-3: kanban list (flat list output -- grep/awk friendly)
- R2-4: kanban show TASK-NNN (full task detail including description)

---

### Release 3: "Installation is frictionless for any CI/CD environment"

Outcome: a new team member can set up kanban hooks and CI integration in under 5 minutes, confirmed by the integration tests passing on their first push.

Adds:
- R3-1: kanban install-hook (installs git commit-msg hook automatically)
- R3-2: Published CI step for GitHub Actions (kanban-ci action)
- R3-3: Published CI step for GitLab CI (.gitlab-ci.yml snippet)
- R3-4: kanban add --dry-run (preview what would be created without writing)

---

## Scope Assessment: PASS

- 7 walking skeleton stories, 4 release stories in R1, 4 in R2, 4 in R3
- 3 bounded contexts: task file management, commit hook integration, CI step integration
- Estimated walking skeleton: 5-7 days
- Estimated full scope (WS + R1): 8-10 days

Walking skeleton is independently shippable. R1 through R3 are incremental releases each
adding a measurable improvement to a distinct outcome.
