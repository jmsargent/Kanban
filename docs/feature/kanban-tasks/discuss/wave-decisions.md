# Wave Decisions: kanban-tasks

Feature: Git-Native Kanban Task Management CLI
Date: 2026-03-15
Wave: DISCUSS (product-owner)

This document records decisions made during the DISCUSS wave, open questions
deferred to DESIGN, and risks surfaced for downstream waves.

---

## Decisions Made in DISCUSS Wave

### D-01: Developer-centric, not PO-centric

The kanban board is a developer tool. Task creation, viewing, and lifecycle management
are developer responsibilities executed in the terminal. The product owner does not
own task creation. Tasks are created by developers as they identify work.

Rationale: The core value proposition is that the board stays accurate because
developers interact with it as a natural extension of their git workflow, not as a
separate management tool.

---

### D-02: Board state is git-native (files in repo)

Tasks are stored as files in .kanban/tasks/ within the git repository itself.
There is no external database, no API, no synchronisation service.

Rationale: This is the mechanism by which the board cannot go stale -- git history
IS the audit log, and the task files are subject to the same peer review and merge
workflow as code.

---

### D-03: Task file format is deferred to DESIGN wave

The internal format of .kanban/tasks/TASK-NNN files (markdown with front matter,
YAML, TOML, or a custom format) is a DESIGN wave decision. Requirements are
format-agnostic: they specify fields (title, status, priority, due, assignee,
description) and behaviours, not serialisation format.

Impact: The commit hook, CI step, kanban board, kanban edit, and kanban delete all
depend on being able to parse the status field reliably. DESIGN must choose a format
that is both human-readable (developers may edit files directly in their editor) and
machine-parseable (hook and CI step must read/write status without external parsers).

---

### D-04: Status transitions are directional and automatic

Transitions move in one direction: todo -> in-progress -> done.
They are triggered by git events, not manual commands.
Manual override via kanban edit is the escape hatch, not the primary path.

Rejected alternative: manual kanban move TASK-001 in-progress. Rejected because
it requires a separate command after the commit, which is exactly the kind of
extra step that causes boards to go stale.

---

### D-05: Commit hook never blocks git

The commit-msg hook must always exit 0. A kanban error must never prevent a
developer from committing code. Errors are logged to .kanban/hook.log.

Rationale: Developer workflow is the primary concern. kanban is a supporting tool.
If the hook breaks, code must still flow.

---

### D-06: CI pipeline failure means no task transition

Tasks only move to "done" when all tests pass (CI exit code 0). A failing pipeline
is not done -- the task stays in-progress. No partial transitions.

Rationale: "Done" means the tests said so. This is the semantic value of the CI
integration: the definition of done is objective and automatic.

---

### D-07: Assignee is free-text for MVP

The assignee field accepts any string. No validation against a user list.

Rationale: Validated user lists require a user management feature. Free-text unblocks
the walking skeleton. Teams will typically use consistent names naturally.
Upgrade to validated user picker is deferred to a later release.

---

### D-08: kanban does not auto-commit on delete

When a task is deleted, the developer must commit the deletion manually.
Output suggests the git commit command.

Rationale: Developers own their commit history. Auto-commits from CLI tools
create noise in git log and can interfere with in-progress rebases or stashes.

---

## Open Decisions (Deferred to DESIGN Wave)

| ID | Decision | Options | Impact |
|---|---|---|---|
| OD-01 | Task file format | Markdown with YAML front matter, pure YAML, TOML, custom | Affects how hook and CI step parse status field; affects "kanban edit" ergonomics |
| OD-02 | Task ID generation strategy | Sequential TASK-NNN vs content-addressed hash | Sequential is human-friendly; hash avoids concurrent-add ID collisions |
| OD-03 | Column configuration | Fixed (todo/in-progress/done) vs configurable per repo in .kanban/config | Configurable is more powerful; adds validation complexity to status writes |
| OD-04 | Implementation language | Go, Rust, Node.js, Python, shell | Affects performance, distribution, and CI step portability |
| OD-05 | Hook delivery mechanism | Committed to repo as .kanban/hooks/ vs installed to .git/hooks/ via kanban install-hook | .git/hooks is local only; .kanban/hooks/ can be shared and version-controlled |
| OD-06 | CI step distribution | Shell script in repo, npm package, GitHub Action, GitLab CI template | Affects discoverability and ease of adoption |

---

## Risks

### R-01: Commit message discipline (HIGH probability, HIGH impact)

Risk: Developers do not consistently include task IDs in commit messages.
If adoption is less than ~70%, auto-transitions rarely fire and the board
degrades to a manual-update tool -- defeating the core value proposition.

Mitigation: The commit tip shown after "kanban add" prompts the correct format.
The kanban install-hook experience should reinforce the pattern. Consider a
future "kanban lint" that warns when commits lack task references (without blocking).
Owner: solution-architect to consider in design; team culture is also a factor.

---

### R-02: Concurrent kanban add ID collisions (MEDIUM probability, HIGH impact)

Risk: Two developers run "kanban add" simultaneously in the same repo and both
generate TASK-005.md. One file is lost.

Mitigation: ID generation must be atomic. The solution depends on the chosen
implementation language and file system guarantees. Content-addressed IDs (OD-02)
eliminate this risk entirely.
Owner: solution-architect.

---

### R-03: CI step commits causing merge conflicts (LOW probability, MEDIUM impact)

Risk: The CI step commits updated task files back to the repo. If another commit
hits the same branch concurrently, the CI commit may conflict.

Mitigation: Task files are rarely modified by more than one pipeline simultaneously
(each pipeline run processes its own commits). If a conflict occurs, the CI step
should retry with a rebase. The file format should minimise diff surface area
(OD-01: a single-line status field is safest).
Owner: solution-architect.

---

### R-04: .kanban/config format mismatch between hook and CI (MEDIUM probability, HIGH impact)

Risk: The commit hook and CI step use different config-reading logic and interpret
ci_task_pattern differently, causing inconsistent transitions.

Mitigation: Both must use the same shared config-reading code (not duplicated parsing logic).
This is an implementation constraint for the DESIGN wave.
Owner: solution-architect.

---

## Handoff Package Summary for DESIGN Wave

Artifacts produced in this DISCUSS wave:

| Artifact | Path | Purpose |
|---|---|---|
| Visual journey map | docs/feature/kanban-tasks/discuss/journey-task-management-visual.md | CLI mockups, emotional arc, flow |
| Journey schema | docs/feature/kanban-tasks/discuss/journey-task-management.yaml | Structured journey with Gherkin per step |
| Gherkin scenarios | docs/feature/kanban-tasks/discuss/journey-task-management.feature | Full acceptance scenario suite |
| Shared artifact registry | docs/feature/kanban-tasks/discuss/shared-artifacts-registry.md | Integration points and risk |
| Story map | docs/feature/kanban-tasks/discuss/story-map.md | Walking skeleton and release slices |
| Prioritization | docs/feature/kanban-tasks/discuss/prioritization.md | Priority order with rationale |
| Requirements | docs/feature/kanban-tasks/discuss/requirements.md | Functional, NFR, business rules, glossary |
| User stories | docs/feature/kanban-tasks/discuss/user-stories.md | 7 stories, all DoR-passed |
| Acceptance criteria | docs/feature/kanban-tasks/discuss/acceptance-criteria.md | Testable, derived from UAT |
| DoR checklist | docs/feature/kanban-tasks/discuss/dor-checklist.md | All 7 stories PASSED |
| Outcome KPIs | docs/feature/kanban-tasks/discuss/outcome-kpis.md | North star + leading indicators |
| Wave decisions | docs/feature/kanban-tasks/discuss/wave-decisions.md | This document |

All 7 user stories pass Definition of Ready.
Feature is cleared for DESIGN wave handoff to solution-architect.
