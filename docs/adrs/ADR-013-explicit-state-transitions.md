# ADR-013: Explicit State Transitions — Remove Auto-Commit Infrastructure

**Status**: Accepted
**Date**: 2026-03-22
**Feature**: explicit-state-transitions
**Supersedes**: ADR-004 (Amendment, 2026-03-18), ADR-005 (Amendment, 2026-03-18), ADR-011

---

## Context

The original kanban design (ADR-004, ADR-005) included two paths where the kanban binary performed git commits autonomously:

1. **commit-msg hook** (`kanban _hook commit-msg`): fired on each developer commit, appended a transition entry to `.kanban/transitions.log`, auto-advancing tasks from `todo` to `in-progress`. Installed via `kanban install-hook`.
2. **CI step** (`kanban ci-done --commit`): after pipeline tests passed, scanned commit messages, appended done transitions to `transitions.log`, then called `git commit` to persist the log.
3. **Init step** (`kanban init`): committed the initial `.kanban/config` and hook installation.

The `board-state-in-git` feature (ADR-011) formalised `transitions.log` as the authoritative board state source, replacing the `status:` YAML field in task files.

After initial use, teams reported **negative feedback** about the tool performing git commits on their behalf. On investigation, the `transitions.log` architecture was found to be unnecessary: the `.kanban/tasks/` folder already carries state via YAML front matter, and that state changes across commits like any other project file.

---

## Decision

1. **Remove `transitions.log` and all infrastructure that writes or reads it.** The `TransitionLogRepository` port, its adapter, and all callers are deleted.

2. **`domain.Task.Status` (YAML front matter `status:` field) is the sole, authoritative board state source.** `GetBoard` reads `t.Status` directly from task structs loaded by `TaskRepository.ListAll()`.

3. **The kanban binary never calls `git commit` or `git add`** (extended constraint C-03). Removed from: `TransitionToDone`, `InitRepo`. Methods `CommitFiles` and `InstallHook` are removed from `GitPort`.

4. **The commit-msg hook is removed.** `kanban _hook commit-msg`, `kanban install-hook`, and the `.kanban/hooks/` directory are deleted. Leftover hooks on developer machines become a safe no-op (the delegated subcommand exits 0 immediately).

5. **State transitions are explicit via CLI commands:**
   - `kanban start <TASK-ID>` → sets `status: in-progress` in YAML, does not commit.
   - `kanban done <TASK-ID>` *(new)* → sets `status: done` in YAML, does not commit.
   - `kanban ci-done --from=SHA` → identifies tasks from commit messages, sets `status: done` in YAML for each, does not commit. CI config owns the subsequent `git add && git commit` step.
   - `kanban init` → creates dirs, writes config. Does not commit.

6. **`GitPort.GetIdentity()` is retained.** Creator attribution (`created_by:`) and assignee auto-population (`kanban start`) continue to read from `git config user.name/email`.

---

## Alternatives Considered

### Alternative 1: Keep transitions.log but remove only the auto-commit (CI step)

Leave the commit-msg hook. Remove only the `CommitFiles` call from `ci-done`. The CI pipeline would own the explicit `git add .kanban/transitions.log && git commit` step.

**Rejection rationale**: The transitions.log itself was questioned by the team: "I do not see a reason for the transition log. The folder .kanban contains the state." Retaining the log retains complexity (flock-based concurrent writes, platform-specific build constraints, a second state source) without benefit. The YAML `status:` field already carries the same information more simply.

### Alternative 2: Keep transitions.log as read-only audit trail, remove write paths

Keep `TransitionLogRepository.History()` for `kanban log`, but remove Append. State source for board moves back to YAML.

**Rejection rationale**: `GitPort.LogFile()` already provides the full commit history of a task file, which serves as the audit trail. A separate log file duplicating what git already tracks is redundant. `kanban log` can be powered by `LogFile` alone.

### Alternative 3: Keep commit-msg hook but make it opt-in

Add a `kanban init --with-hook` flag. Hook not installed by default.

**Rejection rationale**: The hook itself is not the problem — the auto-transition side effect is. Teams prefer to know that `git commit` only does what git commit does, not trigger kanban state changes. Opt-in complexity for a feature with negative feedback is not justified.

---

## Consequences

**Positive**:
- Kanban binary never touches git history — developer and CI config have full ownership
- Single state source eliminates the `transitions.log` vs. YAML inconsistency
- Removed flock-based platform-specific code → Windows cross-compile build failure resolved
- Simpler architecture: 1 port removed, ~8 files deleted, wiring simplified
- `kanban init` no longer requires git push credentials to succeed
- `kanban ci-done` no longer requires CI runner to have git push credentials

**Negative**:
- In-progress and done transitions are no longer automatic. Teams must add `kanban done <ID>` to their workflow or add a `kanban ci-done` step + explicit commit step to CI config.
- `kanban log` output changes: previously merged transitions.log and git log; now shows only git commit history for the task file. Transition details (from→to) are only visible in commit messages, not as structured log entries.
- `kanban init` no longer auto-commits the initial config. Developer must run `git add .kanban/ && git commit` after `kanban init`.

**Migration** (not applicable): Project is not yet publicly released. No backward compatibility required. Existing `.kanban/transitions.log` files in development repos can be deleted.

---

## Impact on Prior ADRs

| ADR | Impact |
|-----|--------|
| ADR-004 (2026-03-15 + 2026-03-18 amendment) | **Superseded**: hook strategy removed. Hook type, delivery mechanism, exit-0 guarantee, and amendment all no longer apply. |
| ADR-005 (2026-03-15 + 2026-03-18 amendment) | **Superseded**: `kanban ci-done` no longer commits. `--commit` flag removed. `[skip ci]` annotation no longer needed. CI config owns the commit step. |
| ADR-011 (transitions.log ADR) | **Superseded**: transitions.log and TransitionLogRepository removed entirely. |
| ADR-001 (hexagonal architecture) | **Unaffected** — port removed cleanly, dependency rule preserved. |
| ADR-002 (task file format) | **Clarified** — `status:` field is now confirmed as authoritative state, not secondary. |
| ADR-007 (GetIdentity) | **Unaffected** — `GitPort.GetIdentity()` retained. |
