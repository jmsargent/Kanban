# Evolution: github-api-board-viewer

**Date**: 2026-04-06
**Feature**: Remote board viewing via GitHub API proxy
**Status**: Complete — all 8 steps COMMITTED, all 355 tests passing

---

## Summary

Extended `kanban-web` with a second operating mode (`--mode=github-api`) that serves a read-only kanban board fetched from any public GitHub repository via the GitHub REST API. No new binary, no new external runtime dependencies — the existing hexagonal architecture was extended with one new secondary adapter package.

---

## Business Context

Users wanted to view kanban boards from public GitHub repositories without cloning or having repository access. The feature enables sharing a board URL (`/remote/board?owner=X&repo=Y`) that any visitor can open in a browser to see a live board fetched from GitHub.

---

## Architecture Decisions

| Decision | ADR | Outcome |
|---|---|---|
| `--mode` flag required, no default, fail fast | ADR-019 | Prevents misconfiguration; both modes coexist in one binary |
| In-process TTL cache (60s) in githubapi.Adapter | ADR-020 | Avoids hammering GitHub API on repeated loads |

### New Components

| Component | Location | Role |
|---|---|---|
| `GitHubAPIPort` | `internal/ports/github_api.go` | Driven port for remote task fetching |
| `GetRemoteBoard` | `internal/usecases/get_remote_board.go` | Maps `[]domain.Task` → `domain.Board` |
| `githubapi.Adapter` | `internal/adapters/githubapi/adapter.go` | Fetches from GitHub REST API; owns TTL cache |
| `RemoteBoardHandler` | `internal/adapters/web/handler.go` | Handles `GET /remote/board?owner=X&repo=Y` |

### Dependency Inversion

```
RemoteBoardHandler (primary adapter)
  → GetRemoteBoard use case
    → GitHubAPIPort (interface, internal/ports)
      ← githubapi.Adapter (secondary adapter, implements port)
```

`cmd/kanban-web/main.go` is the sole wiring point. No adapter imports another adapter.

---

## Key Implementation Details

### Error Distinction (ErrRepositoryNotFound vs ErrNoBoardFound)

The GitHub contents API returns 404 for both "repo doesn't exist" and "repo exists but path absent". The adapter resolves this ambiguity with a secondary call to `GET /repos/{owner}/{repo}`: 200 → `ErrNoBoardFound`, 404 → `ErrRepositoryNotFound`.

### HTML Template Escaping

Go's `html/template` HTML-escapes `&` to `&amp;` in attribute values. The `hx-push-url` assertion in the DSL was updated to check for `&amp;` (correct HTML) rather than the raw `&`.

### GIT_INDEX_FILE Propagation Fix

A pre-existing bug: `GIT_INDEX_FILE` (set by git during hook execution) was leaking into test subprocess environments, causing temp-repo git operations to corrupt the main repo's index. Fixed by filtering `GIT_*` env vars in 4 places: `git_adapter_test.go`, `repo_driver.go`, `server_driver.go`, `tests/acceptance/dsl/context.go`.

### Index Corruption Recovery

101 accumulated `prePush` stashes from previous failed commit attempts created dangling objects. A test-leaked blob reference (`723c3d53`) in the index caused "Error building trees". Fixed by cleaning the stash list and re-adding the corrupted index entry with a valid blob.

---

## Steps Completed

| Step | Name | Notes |
|---|---|---|
| 01-01 | Walking skeleton | All components wired end-to-end |
| 01-02 | Empty repository → three empty columns | Already handled by use case |
| 01-03 | Page title hx-push-url reflects owner/repo | Template already correct; assertion needed `&amp;` fix |
| 01-04 | Invalid owner → Repository not found | Real GitHub API call in test |
| 01-05 | Invalid repo name → Repository not found | Real GitHub API call in test |
| 01-06 | Private repo → Repository not found | Stub: `AddNotFoundRepo` |
| 01-07 | No kanban folder → "no kanban board" | Required `classifyNotFound()` secondary API call |
| 01-08 | Rate limit → retry message | Stub: `AddRateLimitedRepo` |

---

## Refactoring

- Extracted shared `columnView` struct and `boardColumns()` helper from `BoardHandler` and `RemoteBoardHandler`, eliminating duplicated type definitions and loop bodies.

---

## Test Infrastructure Added

- `GitHubAPIStubDriver`: httptest server stubbing GitHub contents API + per-file download endpoints + repo-existence check
- `github_api_steps.go`: DSL Given/When/Then steps for GitHub API board scenarios
- `WebContext` extended with `GitHubAPIStub` and `GitHubStubURL` fields

---

## Lessons Learned

1. **GIT_INDEX_FILE leakage is insidious** — any test that spawns `git` subprocesses inside a pre-commit hook context must filter `GIT_*` env vars. The fix is now in all four subprocess-spawning test helpers.

2. **Stash accumulation corrupts the index** — repeated failed commits create prePush stashes that can accumulate stale/corrupt object references. Periodic `git stash clear` when stashes are not needed prevents this.

3. **Go HTML template `&` escaping** — custom HTML attributes (like htmx's `hx-push-url`) are not recognized as URL contexts by `html/template`, so `&` is always HTML-escaped to `&amp;`. Test assertions must match the escaped form.

4. **GitHub contents 404 is ambiguous** — a secondary `/repos/{owner}/{repo}` call is needed to distinguish "repo not found" from "no kanban board". The stub must also handle this secondary endpoint.

---

## Artifacts Migrated

| Source | Destination |
|---|---|
| `docs/feature/github-api-board-viewer/design/architecture-design.md` | `docs/architecture/github-api-board-viewer/architecture-design.md` |
| `docs/feature/github-api-board-viewer/distill/test-scenarios.md` | `docs/scenarios/github-api-board-viewer/test-scenarios.md` |
| ADR-019, ADR-020 | Already at `docs/adrs/` |