# Lean Canvas — Board State in Git
**Feature**: board-state-in-git
**Phase**: 4 — Market Viability
**Status**: COMPLETE
**Date**: 2026-03-18

---

## Lean Canvas

### 1. Problem (Phase 1 validated)

Top 3 problems confirmed with past-behavior signals:

1. **No audit trail**: There is no way to know *when* a task transitioned or *who* moved it — only the current status is visible.
2. **Hook writes to task files**: The commit-msg hook mutating `.kanban/tasks/` files as a side effect feels architecturally wrong — state bleeds into definition storage.
3. **Repository noise from state rewrites**: Status field writes create unnecessary file churn in commits and diffs, mixing task-state housekeeping with real work.

### 2. Customer Segments (by JTBD)

Primary segment: **Solo developers or small teams (2-5) using git-native workflows who want task tracking without leaving git.**

Job-to-be-done: "When I commit code, help me know that my task state is recorded automatically and permanently without me having to maintain a parallel system."

Secondary segment (future): Teams using pull-request workflows where task state transitions correspond naturally to PR merge events.

### 3. Unique Value Proposition

> Task state lives where it belongs — in git history. No separate task-state files, no hook side-effects, no "where did that go" moments. One `kanban log TASK-001` shows you every transition, when, and by whom.

### 4. Solution (top 3 features for top 3 problems)

Based on Phase 3 findings, **Concept B (Append-Only Transition Log)** is the recommended solution:

1. **`.kanban/transitions.log`** — Append-only file where each line records a state transition: `<ISO8601> <TASK-ID> <from>-><to> <author-email> <trigger>`. The hook appends here instead of rewriting task files. Git tracks history of this file.
2. **`kanban log <TASK-ID>`** — New CLI command that reads the log filtered by task ID. Provides the audit trail the user needs in one command.
3. **`status` field removed from task YAML** — Task files store definition only (title, priority, due, assignee, description). Status is derived from the transitions log, not the file. The hook no longer writes to task files.

### 5. Channels

- Primary: `go install github.com/kanban-tasks/kanban/cmd/kanban@latest`
- Secondary: Homebrew tap (`brew install kanban`)
- Discovery: GitHub repository, word-of-mouth among Go developers

### 6. Revenue Streams

Open-source, no direct revenue. Value is developer productivity. Indirect: contributor reputation, potential for future hosted/team tier.

### 7. Cost Structure

- Developer time (sole contributor)
- GitHub Actions CI minutes (minimal)
- Homebrew tap maintenance (near-zero)

### 8. Key Metrics

- Time from `git commit` to `kanban board` reflecting correct state (target: <1s)
- Number of merge conflicts on `.kanban/` per month (target: 0)
- CLI commands to see full task history (target: 1)
- Files modified by a state transition (target: 1, the log, down from 1 per task)

### 9. Unfair Advantage

Git-native by design — not bolted onto git, but using git's own history model as the state store. No external database, no server, no sync daemon. The approach is conceptually correct for a developer workflow tool in a way that file-based status fields are not.

---

## Four Big Risks Assessment

### Risk 1: Value Risk — Will developers want this?

**Question**: Will removing status from task files and replacing it with a log feel like an improvement, or will it feel like a regression (less visible state)?

**Evidence**:
- Primary motivator (audit trail) is directly solved
- The "less clutter" goal is met — status field writes are gone from task files
- `kanban board` still shows current state — the *display* doesn't change, only where state is stored
- Risk: developers who edit task files directly (via `kanban edit`) are accustomed to seeing `status:` in the YAML — removing it may confuse them initially

**Residual risk**: YELLOW — the value is clear but the transition from the current mental model requires documentation.

**Mitigation**: Add a `# status is tracked in .kanban/transitions.log — use kanban board to view` comment in the task file template.

### Risk 2: Usability Risk — Can developers use this?

**Question**: Is the new model discoverable? Does `kanban board` still work? Does `kanban start` still work?

**Evidence from Phase 3**:
- `kanban board`: reads task files for definitions + reads transitions log for state — behavior unchanged from user perspective
- `kanban start`: appends to transitions log — behavior unchanged from user perspective
- `kanban log TASK-001`: new command — needs to be documented but is intuitive
- Commit-msg hook: appends to log instead of rewriting file — behavior invisible to developer (which is the goal)
- `kanban ci-done`: commits the transitions log instead of task files — internally different, externally the same

**Residual risk**: GREEN — the user-facing API is largely unchanged. The internal mechanism changes, but the commands work the same.

### Risk 3: Feasibility Risk — Can we build this?

**Question**: Is the hexagonal architecture compatible with this change? What is the implementation scope?

**Evidence from codebase analysis**:

The `TaskRepository` port interface requires `FindByID`, `ListAll`, `Save`, `Update`, `Delete`, `NextID`. Under the new model, `Update` for status changes is no longer needed (state goes to the log). However, `Update` is still needed for other field changes (title, assignee, etc.).

The new component needed: a `TransitionLogRepository` port with `Append(repoRoot string, entry TransitionEntry)` and `LatestStatus(repoRoot, taskID string) (TaskStatus, error)` and `History(repoRoot, taskID string) ([]TransitionEntry, error)`.

Implementation scope:
- New port interface: `TransitionLogRepository` — 1 day
- New filesystem adapter: `AppendOnlyLogAdapter` — 1-2 days
- Modify `GetBoard` to derive state from log instead of task files — 1 day
- Modify `StartTask` to write to log instead of task file — 0.5 days
- Modify `TransitionToInProgress` (hook) to write to log — 0.5 days
- Modify `TransitionToDone` (ci-done) to write to log instead of task files — 0.5 days
- Remove `status` field from `taskFrontMatter`, `marshalTask`, `unmarshalTask` — 0.5 days
- New use case + CLI command: `kanban log TASK-001` — 1 day
- Update all tests — 2-3 days

**Total estimated scope**: 8-10 developer-days

**Architecture compatibility**: HIGH — hexagonal ports-and-adapters design makes this straightforward. The `TransitionLogRepository` is a new secondary port; the domain changes minimally. No architecture rule violations expected.

**Residual risk**: GREEN — scope is well-defined and bounded.

### Risk 4: Viability Risk — Does this work for the product?

**Question**: Does this change strengthen or weaken the product's core value proposition?

**Core value proposition**: "Git-native task management — tasks live in your repository, state transitions happen automatically on commit."

**Assessment**: The change *strengthens* the value proposition. Today the product is "git-native for storage but file-mutation for state." After this change it is "git-native for storage AND for state." The product more fully delivers on its stated identity.

**ADR-002 compatibility**: ADR-002 chose Markdown+YAML front matter specifically to minimize diff noise and support machine parseability. Removing the `status` field from task files *reduces* diff noise further (status writes no longer happen). The ADR's rationale is strengthened, not undermined. However, ADR-002 would need to be updated to document the new format (no `status` field) and the transitions log.

**Non-negotiable constraint compatibility** (CLAUDE.md):
- "The binary never auto-commits" — PRESERVED. The hook appends to the log but does not commit. Committing the log is the developer's next natural commit. `kanban ci-done` still makes one commit (of the log).
- "Atomic file writes" — PRESERVED. The log append can be made atomic with file locking or rename patterns.
- "Exit 0 guarantee on hook" — PRESERVED. Log append is simpler than file rewrite; less surface area for failure.

**Residual risk**: GREEN — the change deepens the product's identity and respects all constraints.

---

## Go/No-Go Recommendation

### On the original question: "Should task/board state be stored in git commits (commit trailers)?"

**Answer: NO for full commit-trailer approach (Concept A). YES for the hybrid transition-log approach (Concept B).**

The specific reasons Option A (commit trailers as the *sole* state store) should not be built:
1. `kanban start` cannot record state without making an auto-commit — violates a non-negotiable constraint
2. `kanban board` performance degrades linearly with commits, not tasks
3. Interactive rebase silently destroys state history
4. `kanban add` still requires a task file for definition — task files cannot be fully eliminated

The reasons Concept B (append-only transitions log) should be built:
1. Solves all three primary problems (audit trail, hook-writes-files, repo clutter)
2. Preserves all current commands and UX
3. Respects all architectural constraints
4. Strengthens the product's "git-native" identity
5. Implementation scope is well-bounded (~8-10 days)
6. All four risks are GREEN or YELLOW

### On the underlying need: "Does the current system already partially solve this?"

**Yes — and this is worth knowing before committing to the change.**

The current system's `git log -- .kanban/tasks/TASK-001.md` already provides an audit trail. A `kanban log TASK-001` CLI command wrapping this call could be built in 2 hours and would address the primary stated pain with zero architectural change.

**Recommendation**: Build `kanban log TASK-001` first (2 hours). Ship it. Use it for 2-4 weeks. If the audit trail need is satisfied and the other problems (hook-writes-files, clutter) remain bothersome, proceed to the full Concept B implementation. This reduces risk by validating the core assumption before the larger investment.

---

## Gate G4 Evaluation

| Criterion | Target | Current | Status |
|-----------|--------|---------|--------|
| Lean Canvas complete | Yes | Yes | PASS |
| All 4 risks acceptable | All green/yellow | 3 green, 1 yellow | PASS |
| Channel validated | 1+ | go install + Homebrew | PASS |
| Stakeholder sign-off | Required | Solo developer — self-sign-off | PASS |
| Go/no-go documented | Yes | YES — Concept B recommended | PASS |

**Gate G4: PASS**
