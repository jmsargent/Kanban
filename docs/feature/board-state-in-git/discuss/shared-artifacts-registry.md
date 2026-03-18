# Shared Artifacts Registry — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## Purpose

Every `${variable}` appearing in TUI mockups, journey steps, and Gherkin scenarios must have a single documented source of truth. This registry prevents integration failures by ensuring all consumers reference the same canonical source.

---

## Registry

### task_id

```yaml
artifact: task_id
example_value: "TASK-007"
source_of_truth: "tasks.NextID() — internal/usecases/add_task.go"
canonical_file: ".kanban/tasks/TASK-007.md (filename)"
owner: "AddTask use case (internal/usecases/add_task.go)"
integration_risk: "HIGH — task ID is the foreign key linking task files, transitions log, and commit messages"
validation: "All transitions.log entries referencing TASK-007 must have a corresponding .kanban/tasks/TASK-007.md file"
consumers:
  - "transitions.log entries (TASK-ID field)"
  - "kanban log output (header and filter)"
  - "kanban board display (task ID label)"
  - "commit message references (TASK-007 pattern)"
  - "commit-msg hook (regex match pattern)"
  - "kanban start command (argument)"
  - "kanban ci-done (range parsing)"
```

### author_email

```yaml
artifact: author_email
example_value: "jon@kanbandev.io"
source_of_truth: "git config user.email — GitPort.GetIdentity() — internal/ports/git.go"
owner: "GitPort adapter (internal/adapters/git/)"
integration_risk: "HIGH — author identity links transitions to developers; wrong identity breaks per-user filter"
validation: "GetIdentity() returns non-empty Email; any transition recorded with empty email is a bug"
consumers:
  - "transitions.log entries (author field)"
  - "kanban log output (author column)"
  - "kanban board --me filter (filter by email)"
  - "kanban start (records caller identity)"
  - "commit-msg hook (records committer identity)"
  - "kanban ci-done (records CI identity)"
```

### transition_entry

```yaml
artifact: transition_entry
example_value: "2026-03-15T09:14:23Z TASK-007 todo->in-progress jon@kanbandev.io manual"
format: "<ISO8601_UTC> <TASK-ID> <from>-><to> <author_email> <trigger>"
source_of_truth: ".kanban/transitions.log (append-only file)"
owner: "TransitionLogRepository (new port, Phase 2)"
integration_risk: "CRITICAL — all board state, kanban log output, and audit trail depend on this file being correct and append-only"
validation: |
  Each entry must be parseable by TransitionLogRepository.LatestStatus() and History()
  Lines must be valid UTF-8
  ISO8601 timestamps must be in UTC (Z suffix)
  from->to must be valid transitions per domain.CanTransitionTo()
consumers:
  - "kanban board (derives current status per task via LatestStatus)"
  - "kanban log (shows history via History)"
  - "kanban board --me (filters by author_email field)"
  - "kanban ci-done (reads to validate current status before appending done)"
```

### task_title

```yaml
artifact: task_title
example_value: "Add retry logic to CI step"
source_of_truth: ".kanban/tasks/TASK-007.md (YAML front matter: title field)"
owner: "TaskRepository (internal/ports/repositories.go)"
integration_risk: "LOW — display only; no logic depends on title value"
validation: "kanban log header must read title from task file, not from transitions.log"
consumers:
  - "kanban log output (header line: 'TASK-007: <title>')"
  - "kanban board display (task title in column)"
  - "kanban add confirmation output"
```

### task_current_status

```yaml
artifact: task_current_status
example_value: "in-progress"
source_of_truth: ".kanban/transitions.log — TransitionLogRepository.LatestStatus() (Phase 2) OR task YAML status field (Phase 1)"
owner: "Phase 1: TaskRepository.FindByID() | Phase 2: TransitionLogRepository.LatestStatus()"
integration_risk: "HIGH — board correctness depends entirely on this derivation being accurate"
validation: |
  Phase 2: LatestStatus must scan log for the task's MOST RECENT entry
  Tasks with no log entries are implicitly StatusTodo
  Must not return stale state after a new transition is appended
consumers:
  - "kanban board (column placement)"
  - "kanban log (most recent status indicator)"
  - "kanban start (validates todo->in-progress is allowed)"
  - "kanban ci-done (validates in-progress->done is allowed before recording)"
```

### commit_sha

```yaml
artifact: commit_sha
example_value: "a3f12bc"
source_of_truth: "git commit — emitted by commit-msg hook context"
owner: "commit-msg hook (internal/adapters/git/)"
integration_risk: "LOW — appears in transitions.log trigger field and kanban log display; used for audit correlation only"
validation: "SHA format must be 7+ hex characters (short SHA)"
consumers:
  - "transitions.log entries (trigger field: 'commit:a3f12bc')"
  - "kanban log output (commit reference line)"
```

---

## Integration Consistency Check

| Check | Status | Evidence |
|-------|--------|---------|
| All `${variable}` in mockups have documented source | PASS | 6 artifacts registered above |
| No two steps display the same data from different sources | PASS | task_current_status has one source per phase; transitions are consistent |
| Single source of truth for task identity | PASS | task_id sourced from NextID(), consumed everywhere |
| Author identity single source | PASS | author_email sourced from GitPort.GetIdentity() exclusively |
| Transition log is the sole state authority (Phase 2) | PASS | task_current_status source is TransitionLogRepository.LatestStatus() |
| Transition log does NOT appear in task YAML (Phase 2) | PASS | task files have no `status:` field in Phase 2 |
| CLI vocabulary consistent | PASS | "todo", "in-progress", "done" used consistently across log format, board display, kanban log output |

---

## Migration Artifacts (Phase 2 specific)

When migrating from Phase 1 (YAML status) to Phase 2 (transitions log), two additional artifacts are relevant:

### existing_status_field

```yaml
artifact: existing_status_field
example_value: "status: in-progress"
source_of_truth: ".kanban/tasks/*.md YAML front matter (CURRENT — Phase 1)"
migration_action: "Remove from all task files during migration; backfill transitions.log from existing status values"
risk: "Tasks with status: in-progress must produce a transitions.log entry on migration"
```

### migration_log_entry

```yaml
artifact: migration_log_entry
example_value: "2026-03-18T00:00:00Z TASK-005 todo->in-progress unknown@migration migration:backfill"
purpose: "Synthetic entry created during migration to preserve current state in the new log format"
trigger: "kanban migrate command (proposed) or kanban init --migrate"
```
