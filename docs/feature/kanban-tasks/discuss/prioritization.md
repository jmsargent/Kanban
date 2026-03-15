# Prioritization: kanban-tasks

Date: 2026-03-15
Method: Value x Urgency / Effort (1-5 scale), tie-break: Walking Skeleton > Riskiest Assumption > Highest Value

---

## Riskiest Assumptions

Before optimising value delivery, these assumptions must be validated by the walking skeleton:

1. Developers will actually reference task IDs in commit messages consistently (if not, auto-transitions never fire)
2. File-based task storage is fast enough and conflict-free for a 4-person team committing concurrently
3. CI/CD integration is low-friction enough that teams install it alongside the CLI

These are kill-level risks. The walking skeleton validates all three within the first release.

---

## Release Priority

| Priority | Release | Target Outcome | North Star KPI | Rationale |
|---|---|---|---|---|
| 1 | Walking Skeleton | Developer completes full work loop in terminal | % of tasks whose status changes were automatic (not manual) | Validates riskiest assumptions; proves end-to-end works |
| 2 | Release 1 | Board is useful as a daily reference | Developer checks kanban board daily without switching tools | High value; fills gaps that make WS barely usable |
| 3 | Release 2 | Board scales to busy team | Time to find own tasks on a 10+ task board under 10 seconds | Urgency increases as team grows |
| 4 | Release 3 | Installation is frictionless | Time to first successful CI integration under 5 minutes | Lowers adoption barrier for new teams |

---

## Backlog Suggestions

Note: story IDs assigned in Phase 4 (Requirements). These are task-level placeholders from the story map.

| Task | Release | Priority | Value | Urgency | Effort | Score | Outcome Link | Dependencies |
|---|---|---|---|---|---|---|---|---|
| kanban init | WS | P1 | 5 | 5 | 1 | 25.0 | End-to-end loop | None |
| kanban add (title only) | WS | P1 | 5 | 5 | 2 | 12.5 | Task capture | kanban init |
| kanban board (grouped) | WS | P1 | 5 | 5 | 2 | 12.5 | Board visibility | kanban add |
| commit hook: todo->in-progress | WS | P1 | 5 | 5 | 3 | 8.3 | Auto-transition | kanban add |
| CI step: in-progress->done | WS | P1 | 5 | 5 | 3 | 8.3 | Auto-transition | commit hook |
| kanban edit ($EDITOR) | WS | P1 | 4 | 4 | 2 | 8.0 | Correction loop | kanban add |
| kanban delete (with confirm) | WS | P1 | 4 | 4 | 2 | 8.0 | Board hygiene | kanban add |
| kanban add with optional flags | R1 | P2 | 4 | 4 | 2 | 8.0 | Daily reference | kanban add |
| Overdue indicator on board | R1 | P2 | 4 | 3 | 1 | 12.0 | Daily reference | kanban board |
| kanban board --json | R1 | P2 | 3 | 3 | 1 | 9.0 | Scripting/CI | kanban board |
| Actionable error messages | R1 | P2 | 4 | 4 | 2 | 8.0 | Daily reference | All commands |
| kanban board --filter status | R2 | P3 | 4 | 2 | 2 | 4.0 | Team scale | kanban board |
| kanban board --filter assignee | R2 | P3 | 4 | 2 | 2 | 4.0 | Team scale | kanban board |
| kanban list (flat output) | R2 | P3 | 3 | 2 | 1 | 6.0 | Scripting | kanban board |
| kanban show TASK-NNN | R2 | P3 | 3 | 2 | 2 | 3.0 | Full detail view | kanban add |
| kanban install-hook | R3 | P3 | 4 | 2 | 2 | 4.0 | Frictionless install | commit hook |
| GitHub Actions CI step | R3 | P3 | 4 | 3 | 3 | 4.0 | Frictionless install | CI step |
| GitLab CI step | R3 | P3 | 3 | 2 | 2 | 3.0 | Frictionless install | CI step |
| kanban add --dry-run | R3 | P4 | 2 | 1 | 1 | 2.0 | Safety net | kanban add |

---

## Walking Skeleton Dependency Order

```
kanban init
    -> kanban add (title only)
        -> kanban board (grouped output)
        -> commit hook (todo -> in-progress)
            -> CI step (in-progress -> done)
        -> kanban edit
        -> kanban delete
```

All walking skeleton tasks must ship together -- a partial skeleton does not validate the
core assumption. Delivering kanban add without the commit hook means developers have to
manually update status, defeating the core value proposition.

---

## Deferred (Won't Have in first releases)

| Item | Reason |
|---|---|
| Web UI / GUI | Out of scope -- CLI and terminal is the target medium |
| User authentication / access control | Single-user and small team model; git repo ACL is sufficient |
| Task comments / threads | Complexity not warranted for MVP; use git commit messages |
| Integrations with GitHub Issues / Jira | Out of scope -- this tool replaces those, not integrates with them |
| Custom column definitions (beyond 3) | Configurable in .kanban/config but advanced; defer UI for editing columns |
