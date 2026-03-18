# ADR-012: Migration Strategy — YAML Status Field to Transitions Log

**Status**: Superseded — Not Required
**Date**: 2026-03-18
**Superseded**: 2026-03-18
**Feature**: board-state-in-git (US-BSG-02)

---

## Context

This ADR was drafted during the DESIGN wave under the assumption that existing repositories with `status:` YAML front matter fields would need a migration path when US-BSG-02 ships.

## Decision

**No migration command is required.**

kanban-tasks has not been publicly released. There are no existing user repositories in the wild with `status:` YAML fields that require migration. The concern motivating this ADR — backward compatibility with pre-release repositories — does not apply.

US-BSG-02 can make a clean break:

- New task files created after US-BSG-02 ships do **not** contain a `status:` field
- `.kanban/transitions.log` is the sole authoritative state source from day one
- `GetBoard` does **not** include a YAML fallback path — there is no dual-source logic
- The `bootstrap` sentinel value in ADR-011 is removed from scope; all `from` values are standard `TaskStatus` values

## Consequences

**Positive**:
- Simpler `GetBoard` implementation — single state source, no fallback branch
- Simpler `LatestStatus` return type — no need for `(TaskStatus, bool)` to signal "no entry"
- No `kanban migrate` subcommand to build, document, or test
- No `bootstrap` sentinel type in `domain` layer
- Fewer acceptance criteria (AC-02 migration scenarios eliminated)

**Impact on US-BSG-02 design**:
- `TransitionLogRepository.LatestStatus` returns `(domain.TaskStatus, error)` — callers treat a missing task entry as `StatusTodo` (implicit; no `bool` needed)
- `GetBoard` uses `LatestStatus` directly; if no entry exists, the task is `todo` by definition (BR-2 from requirements.md)
- Task files with no log entries are valid and represent `todo` status — this is the normal state for newly created tasks

**Note**: The internal `.kanban/tasks/TASK-001.md` file in this repository currently has `status: in-progress` in its YAML. This file was created before US-BSG-02. It is a developer scratch file, not a released artifact. It will be cleaned up as part of the US-BSG-02 implementation (remove the `status:` field from the template and update the existing task file).
