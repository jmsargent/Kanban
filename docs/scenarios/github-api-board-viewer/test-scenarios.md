# Acceptance Test Scenarios: github-api-board-viewer

**Date**: 2026-04-03
**Wave**: DISTILL
**Feature**: Remote board viewing via GitHub API proxy

---

## Scenario Inventory

| # | Test Function | Category | Enabled |
|---|---|---|---|
| 1 | `TestRemoteBoard_ViewerSeesTasksFromPublicRepository` | Walking skeleton | Yes (first to implement) |
| 2 | `TestRemoteBoard_EmptyRepositoryShowsThreeEmptyColumns` | Happy path | Skipped |
| 3 | `TestRemoteBoard_PageTitleEditTriggersRefresh` | Happy path | Skipped |
| 4 | `TestRemoteBoard_InvalidOwnerShowsRepositoryNotFound` | Error path | Skipped |
| 5 | `TestRemoteBoard_InvalidRepoNameShowsRepositoryNotFound` | Error path | Skipped |
| 6 | `TestRemoteBoard_PrivateRepositoryShowsRepositoryNotFound` | Error path | Skipped |
| 7 | `TestRemoteBoard_RepoWithNoKanbanFolderShowsNoBoardMessage` | Error path | Skipped |
| 8 | `TestRemoteBoard_RateLimitExceededShowsRetryMessage` | Error path | Skipped |

**Error path ratio**: 5 of 8 = 62.5% (target >= 40%, target met)

**Walking skeletons**: 1
**Focused scenarios**: 7

---

## Story Coverage Traceability

| Story scenario | Test function |
|---|---|
| S1: Valid owner + repo with tasks → board renders correct columns | `TestRemoteBoard_ViewerSeesTasksFromPublicRepository` |
| S2: Valid owner + repo with no tasks → three empty columns | `TestRemoteBoard_EmptyRepositoryShowsThreeEmptyColumns` |
| S3: Invalid owner → "Repository not found" | `TestRemoteBoard_InvalidOwnerShowsRepositoryNotFound` |
| S4: Valid owner, invalid repo → "Repository not found" | `TestRemoteBoard_InvalidRepoNameShowsRepositoryNotFound` |
| S5: Private repository → "Repository not found" | `TestRemoteBoard_PrivateRepositoryShowsRepositoryNotFound` |
| S6: No .kanban/tasks/ folder → "no kanban board" message | `TestRemoteBoard_RepoWithNoKanbanFolderShowsNoBoardMessage` |
| S7: Rate limit exceeded → retry message | `TestRemoteBoard_RateLimitExceededShowsRetryMessage` |
| Inline page title edit + shareable URL | `TestRemoteBoard_PageTitleEditTriggersRefresh` |

All 7 story scenarios plus the title-edit behaviour are covered.

---

## Walking Skeleton

`TestRemoteBoard_ViewerSeesTasksFromPublicRepository` is the walking skeleton.

Litmus test:
- Title: "Viewer sees tasks from public repository" — user goal, not technical flow. Pass.
- Given/When: viewer actions (configures stub, requests board). Pass.
- Then: user observations (board has columns, tasks in correct columns, board is read-only). Pass.
- Stakeholder demo-ability: a product owner can confirm "yes, that is what we need". Pass.

The skeleton touches all new layers as a consequence of the user journey:
`HTTP request -> web handler -> GetRemoteBoard use case -> GitHubAPIPort -> githubapi.Adapter (stub) -> board.html template`.

---

## Driving Port

All scenarios invoke through `GET /remote/board?owner=X&repo=Y` — the HTTP interface of `kanban-web --mode=github-api`. This is the driving port defined in the architecture.

No internal components (`GetRemoteBoard` use case, `GitHubAPIPort`, `githubapi.Adapter`) are invoked directly. The GitHub API is stubbed at the network boundary via `httptest.Server` in `driver/github_api_stub_driver.go`.

---

## Implementation Sequence (one at a time)

1. Enable: `TestRemoteBoard_ViewerSeesTasksFromPublicRepository` (walking skeleton — currently the only enabled test)
2. Enable: `TestRemoteBoard_EmptyRepositoryShowsThreeEmptyColumns`
3. Enable: `TestRemoteBoard_InvalidOwnerShowsRepositoryNotFound`
4. Enable: `TestRemoteBoard_InvalidRepoNameShowsRepositoryNotFound`
5. Enable: `TestRemoteBoard_PrivateRepositoryShowsRepositoryNotFound`
6. Enable: `TestRemoteBoard_RepoWithNoKanbanFolderShowsNoBoardMessage`
7. Enable: `TestRemoteBoard_RateLimitExceededShowsRetryMessage`
8. Enable: `TestRemoteBoard_PageTitleEditTriggersRefresh`

Remove `t.Skip(...)` one test at a time. Commit after each passes.

---

## Mandate Compliance Evidence

### CM-A: Driving Port Usage

All test files import only the DSL package (dot import). DSL functions call the HTTP driver. No test file imports `internal/usecases/`, `internal/adapters/githubapi/`, or any internal package.

```
tests/acceptance/backend/github_api_board_test.go
  import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"
  (no other imports)
```

### CM-B: Zero Technical Terms in Gherkin

Test function names and DSL step names use business language:
- "Viewer sees tasks from public repository" (not "HTTP 200 from /remote/board")
- "Repository not found" (not "404 from GitHub API")
- "GitHub API rate limit exceeded. Try again later." (exact user-facing message, not "HTTP 429")
- "Board is read-only" (not "POST /task route not registered")

### CM-C: Walking Skeleton + Focused Scenario Counts

- Walking skeletons: 1 (`TestRemoteBoard_ViewerSeesTasksFromPublicRepository`)
- Focused scenarios: 7
- Error path scenarios: 5 (62.5%)
- All 7 story scenarios covered plus inline title edit behaviour
