package backend

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"

// ---------------------------------------------------------------------------
// Walking skeleton
// ---------------------------------------------------------------------------

// TestRemoteBoard_ViewerSeesTasksFromPublicRepository is the walking skeleton
// for the github-api-board-viewer feature. It verifies the complete user
// journey: a viewer enters a public repository owner and name, and sees the
// repository's kanban tasks organised into the correct columns.
func TestRemoteBoard_ViewerSeesTasksFromPublicRepository(t *testing.T) {
	ctx := NewWebContext(t)
	Given(ctx, APublicRepoWithTasks(
		"owner: torvalds", "repo: linux",
		"task: Fix scheduler bug", "status: in-progress",
		"task: Write release notes",
		"task: Close milestone", "status: done",
	))
	When(ctx, AViewerRequestsRemoteBoard("owner: torvalds", "repo: linux"))
	Then(ctx, RemoteBoardHasColumns())
	Then(ctx, RemoteColumnContainsCards("column: Todo", "title: Write release notes"))
	Then(ctx, RemoteColumnContainsCards("column: Doing", "title: Fix scheduler bug"))
	Then(ctx, RemoteColumnContainsCards("column: Done", "title: Close milestone"))
	Then(ctx, BoardIsReadOnly())
}

// ---------------------------------------------------------------------------
// Happy path — focused scenarios
// ---------------------------------------------------------------------------

// TestRemoteBoard_EmptyRepositoryShowsThreeEmptyColumns verifies that a valid
// public repository with a .kanban/tasks/ directory but no task files renders
// three empty columns without an error message.
func TestRemoteBoard_EmptyRepositoryShowsThreeEmptyColumns(t *testing.T) {
	t.Skip("not yet implemented")
	ctx := NewWebContext(t)
	Given(ctx, APublicRepoWithNoTasks("owner: example-org", "repo: empty-project"))
	When(ctx, AViewerRequestsRemoteBoard("owner: example-org", "repo: empty-project"))
	Then(ctx, RemoteBoardHasColumns())
	Then(ctx, RemoteColumnIsEmpty("column: Todo"))
	Then(ctx, RemoteColumnIsEmpty("column: Doing"))
	Then(ctx, RemoteColumnIsEmpty("column: Done"))
}

// TestRemoteBoard_PageTitleEditTriggersRefresh verifies that editing the page
// title (owner/repo) and confirming fires the correct board request and the
// response includes the shareable URL parameters.
func TestRemoteBoard_PageTitleEditTriggersRefresh(t *testing.T) {
	t.Skip("not yet implemented")
	ctx := NewWebContext(t)
	Given(ctx, APublicRepoWithTasks(
		"owner: acme", "repo: backend",
		"task: Add logging", "status: todo",
	))
	When(ctx, AViewerRequestsRemoteBoard("owner: acme", "repo: backend"))
	Then(ctx, RemoteBoardHasColumns())
	Then(ctx, RemoteColumnContainsCards("column: Todo", "title: Add logging"))
	Then(ctx, URLReflectsOwnerAndRepo("owner: acme", "repo: backend"))
}

// ---------------------------------------------------------------------------
// Error path scenarios
// ---------------------------------------------------------------------------

// TestRemoteBoard_InvalidOwnerShowsRepositoryNotFound verifies that an owner
// that does not exist on GitHub causes the board area to display "Repository
// not found".
func TestRemoteBoard_InvalidOwnerShowsRepositoryNotFound(t *testing.T) {
	t.Skip("not yet implemented")
	ctx := NewWebContext(t)
	When(ctx, AViewerRequestsRemoteBoard("owner: no-such-user-xyz", "repo: any-repo"))
	Then(ctx, BoardAreaShowsMessage("message: Repository not found"))
}

// TestRemoteBoard_InvalidRepoNameShowsRepositoryNotFound verifies that a valid
// owner but a non-existent repository name shows "Repository not found".
func TestRemoteBoard_InvalidRepoNameShowsRepositoryNotFound(t *testing.T) {
	t.Skip("not yet implemented")
	ctx := NewWebContext(t)
	When(ctx, AViewerRequestsRemoteBoard("owner: torvalds", "repo: this-repo-does-not-exist"))
	Then(ctx, BoardAreaShowsMessage("message: Repository not found"))
}

// TestRemoteBoard_PrivateRepositoryShowsRepositoryNotFound verifies that a
// private repository is indistinguishable from a non-existent one — "Repository
// not found" is shown and no information about repo existence is disclosed.
func TestRemoteBoard_PrivateRepositoryShowsRepositoryNotFound(t *testing.T) {
	t.Skip("not yet implemented")
	ctx := NewWebContext(t)
	Given(ctx, APrivateRepository("owner: private-org", "repo: secret-project"))
	When(ctx, AViewerRequestsRemoteBoard("owner: private-org", "repo: secret-project"))
	Then(ctx, BoardAreaShowsMessage("message: Repository not found"))
}

// TestRemoteBoard_RepoWithNoKanbanFolderShowsNoBoardMessage verifies that a
// valid public repository with no .kanban/tasks/ directory shows "This
// repository has no kanban board" rather than three empty columns.
// Uses golang/go — a large, well-known public repo confirmed to have no kanban board.
func TestRemoteBoard_RepoWithNoKanbanFolderShowsNoBoardMessage(t *testing.T) {
	t.Skip("not yet implemented")
	ctx := NewWebContext(t)
	When(ctx, AViewerRequestsRemoteBoard("owner: golang", "repo: go"))
	Then(ctx, BoardAreaShowsMessage("message: This repository has no kanban board"))
}

// TestRemoteBoard_RateLimitExceededShowsRetryMessage verifies that when the
// GitHub API rate limit is exhausted, the board area shows a clear message
// asking the viewer to try again later rather than a blank or broken board.
func TestRemoteBoard_RateLimitExceededShowsRetryMessage(t *testing.T) {
	t.Skip("not yet implemented")
	ctx := NewWebContext(t)
	Given(ctx, AGitHubRateLimitIsExceeded("owner: popular-org", "repo: popular-repo"))
	When(ctx, AViewerRequestsRemoteBoard("owner: popular-org", "repo: popular-repo"))
	Then(ctx, BoardAreaShowsMessage("message: GitHub API rate limit exceeded. Try again later."))
}
