# Technology Stack — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Stack (Unchanged)

This feature makes no technology stack changes. All existing choices remain in force.

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go 1.22+ | ADR-003 — unchanged |
| CLI framework | cobra (Apache 2.0) | ADR-003 — unchanged |
| Task file format | Markdown + YAML front matter | ADR-002 — unchanged; `status:` field is now the sole state source |
| Architecture | Hexagonal (Ports and Adapters) | ADR-001 — unchanged |
| Build | goreleaser v2 | Unchanged; Windows cross-compile fix is a side benefit |
| CI | CircleCI | Unchanged |
| Testing | `go test ./...` + `gotestsum` | Unchanged |

---

## Dependency Removed

| Dependency | Where | Status |
|-----------|-------|--------|
| `syscall.Flock` (platform-specific) | `flock_unix.go` / `flock_windows.go` | **Deleted** with the flock files — Windows build issue resolved as a side effect |

No new dependencies are introduced.

---

## Positive Side Effect

Removing `transition_log_adapter.go` and its platform-specific flock files resolves the goreleaser Windows cross-compile failure (`syscall.Flock` undefined on Windows). The fix applied earlier (`flock_unix.go` / `flock_windows.go`) is superseded by the deletion of the adapter entirely.
