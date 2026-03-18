# Wave Decisions — board-state-in-git (DESIGN wave)

**Feature**: board-state-in-git
**Date**: 2026-03-18
**Wave**: DESIGN

---

## Decisions Made This Wave

| ID | Decision | Rationale | ADR |
|----|---------|-----------|-----|
| WD-01 | transitions.log line format: 5 space-separated fields, ISO 8601 UTC timestamp, `from->to` compound field, trigger with colon-separated SHA | No quoting needed (no field contains spaces); grep-friendly; no parser library required; deterministic field count enables strict validation | ADR-011 |
| WD-02 | File locking via `syscall.Flock(LOCK_EX)` + `O_APPEND` | Stdlib only; sufficient for single-machine advisory locking; no third-party dependency; Windows out of scope | ADR-011 |
| WD-03 | `git log --follow` for `kanban log` history source | Tracks file renames within branch history; no second data store needed; directly auditable | ADR-011 |
| WD-04 | No migration command required | kanban-tasks has not been publicly released; no existing user repos have YAML `status:` fields in the wild; clean break is safe | ADR-012 (superseded) |
| WD-05 | ~~Migration entries use `bootstrap->status` form~~ | Eliminated — no migration; `from` field is always a standard TaskStatus | N/A |
| WD-06 | `GetBoard` does NOT fall back to YAML `status:` | Single authoritative state source from day one; no dual-source logic; simpler `LatestStatus` returns `(TaskStatus, error)` with implicit todo for missing entries | ADR-012 (superseded) |
| WD-07 | `GetTaskHistory` sources history from git commit log (US-BSG-01), not transitions.log | Walking skeleton validates the git log integration path independently; transitions.log history can be surfaced later without changing the port interface | architecture-design.md §6.3 |
| WD-08 | `kanban log --json` deferred to post-MVP | No current consumer of JSON output; deferred scope keeps US-BSG-01 to its 1-day estimate | RQ-3 |
| WD-09 | `TransitionToDone` commits only `.kanban/transitions.log` | Narrows commit surface; reduces merge conflicts; task files no longer change on status transitions | ADR-005 update |
| WD-10 | `assignee` field normalised to email in task YAML | Required for `--me` matching against `git config user.email`; consistent with `author_email` field in transitions.log | component-boundaries.md |

---

## Deferred Decisions

| ID | Question | Deferred To |
|----|---------|------------|
| DD-01 | In-memory index for transitions.log at scale (>100k entries) | Post-MVP; performance acceptable without index at expected scale |
| DD-02 | `kanban log --json` output format | Post-MVP; no current consumer |
| DD-03 | Windows support for file locking | Out of scope; Windows not a supported platform |
| DD-04 | Stripping legacy `status:` fields from task files after migration | Post-migration cleanup; not part of this feature scope |

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| `git log --follow` orphans entries after rebase/force-push | Low | Medium | Documented as known limitation in ADR-011; not a blocker; no architecture change needed |
| Concurrent hook + CI write to transitions.log | Low | High | `Flock(LOCK_EX)` prevents corruption; tested with concurrent test cases |
| Developer edits transitions.log manually and corrupts format | Low | Medium | Parser skips malformed lines with warning; does not crash; integrity recoverable via git |
| Migration creates duplicate entries if run twice | None | Low | `kanban migrate` is idempotent: checks for existing `migration` trigger entries before appending |
