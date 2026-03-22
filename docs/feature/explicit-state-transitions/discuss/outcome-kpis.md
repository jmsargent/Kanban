# Outcome KPIs — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Primary Outcome

**Eliminate negative feedback about kanban performing git commits.**

---

## KPIs

| KPI | Target | Measurement |
|-----|--------|-------------|
| Zero auto-commits from kanban binary | 100% — no code path calls `git commit` | Verified by `grep -r "git.*commit" internal/` returning 0 results in production code |
| `kanban ci-done` exits 0 without staging or committing | 100% of test runs | Acceptance test AC-02-2: no git subprocess invoked |
| `kanban board` reads state without transitions.log | 100% of board calls | AC-03-2: board works when transitions.log absent |
| Windows cross-compile succeeds | goreleaser builds all 6 targets | AC-05-3: `windows_amd64` build passes (no syscall.Flock) |
| All existing tests pass after removal | 0 test failures | AC-05-4: `go test ./...` green |

---

## North Star

> Developers feel that kanban is a read/query tool that also updates task files — not a tool that owns their git history.

This is qualitative and assessed by the team on first use of the updated binary.
