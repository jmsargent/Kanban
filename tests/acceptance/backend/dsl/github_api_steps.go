package dsl

import (
	"fmt"
	"strings"

	"github.com/jmsargent/kanban/pkg/simpledsl"
	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

// ---------------------------------------------------------------------------
// DSL param schemas
// ---------------------------------------------------------------------------

var aRemoteRepoWithTasksDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("owner"),
	simpledsl.NewRequiredArg("repo"),
	simpledsl.NewRepeatingGroup("task",
		simpledsl.NewRequiredArg("title"),
		simpledsl.NewOptionalArg("status").SetDefault("todo"),
		simpledsl.NewOptionalArg("assignee"),
	),
)

var aRemoteRepoDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("owner"),
	simpledsl.NewRequiredArg("repo"),
)

var aViewerRequestsRemoteBoardDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("owner"),
	simpledsl.NewRequiredArg("repo"),
)

var boardAreaShowsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("message"),
)

var remoteColumnContainsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("column"),
	simpledsl.NewOptionalArg("title").SetAllowMultipleValues(true),
)

var urlContainsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("owner"),
	simpledsl.NewRequiredArg("repo"),
)

// ---------------------------------------------------------------------------
// Given steps — repo configuration (stub startup is internal)
// ---------------------------------------------------------------------------

// APublicRepoWithTasks configures a stub public repository that contains the
// given task files. The stub is started automatically if not already running.
//
// Required params: "owner: <owner>", "repo: <repo>"
// Repeating group: "task: <title>", optional "status: <status>", "assignee: <assignee>"
func APublicRepoWithTasks(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("a public repo with tasks (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			if err := ensureStub(ctx); err != nil {
				return err
			}
			vals, err := aRemoteRepoWithTasksDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("APublicRepoWithTasks: %w", err)
			}
			owner := vals.Value("owner")
			repo := vals.Value("repo")

			var tasks []driver.StubTaskFile
			for i, taskVals := range vals.Group("task") {
				id := fmt.Sprintf("TASK-%03d", i+1)
				filename := id + ".md"
				status := taskVals.Value("status")
				if status == "" {
					status = "todo"
				}
				assignee := taskVals.Value("assignee")
				content := buildTaskFileContent(id, taskVals.Value("title"), status, assignee)
				tasks = append(tasks, driver.StubTaskFile{Filename: filename, Content: content})
			}
			ctx.GitHubAPIStub.AddRepoWithTasks(owner, repo, tasks)
			return nil
		},
	}
}

// APublicRepoWithNoTasks configures a stub public repository that has a
// .kanban/tasks/ directory but contains no task files. The stub is started
// automatically if not already running.
//
// Required params: "owner: <owner>", "repo: <repo>"
func APublicRepoWithNoTasks(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("a public repo with no tasks (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			if err := ensureStub(ctx); err != nil {
				return err
			}
			vals, err := aRemoteRepoDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("APublicRepoWithNoTasks: %w", err)
			}
			ctx.GitHubAPIStub.AddRepoWithNoTasks(vals.Value("owner"), vals.Value("repo"))
			return nil
		},
	}
}

// APublicRepoWithNoKanbanBoard configures a stub public repository that exists
// but has no .kanban/tasks/ directory. The stub is started automatically if
// not already running.
//
// Required params: "owner: <owner>", "repo: <repo>"
func APublicRepoWithNoKanbanBoard(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("a public repo with no kanban board (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			if err := ensureStub(ctx); err != nil {
				return err
			}
			vals, err := aRemoteRepoDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("APublicRepoWithNoKanbanBoard: %w", err)
			}
			ctx.GitHubAPIStub.AddRepoWithNoBoard(vals.Value("owner"), vals.Value("repo"))
			return nil
		},
	}
}

// APrivateRepository configures a stub repository that the server cannot
// access — GitHub returns 404 for private repos and non-existent repos
// identically, so no information about existence is disclosed. The stub is
// started automatically if not already running.
//
// Required params: "owner: <owner>", "repo: <repo>"
func APrivateRepository(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("a private repository (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			if err := ensureStub(ctx); err != nil {
				return err
			}
			vals, err := aRemoteRepoDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("APrivateRepository: %w", err)
			}
			ctx.GitHubAPIStub.AddNotFoundRepo(vals.Value("owner"), vals.Value("repo"))
			return nil
		},
	}
}

// AGitHubRateLimitIsExceeded configures a stub that returns a rate-limit
// response for the given owner/repo. The stub is started automatically if
// not already running.
//
// Required params: "owner: <owner>", "repo: <repo>"
func AGitHubRateLimitIsExceeded(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("the GitHub rate limit is exceeded for (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			if err := ensureStub(ctx); err != nil {
				return err
			}
			vals, err := aRemoteRepoDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("AGitHubRateLimitIsExceeded: %w", err)
			}
			ctx.GitHubAPIStub.AddRateLimitedRepo(vals.Value("owner"), vals.Value("repo"))
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// When steps — user actions
// ---------------------------------------------------------------------------

// AViewerRequestsRemoteBoard starts the kanban-web server in github-api mode
// (if not already running) and requests the board for the given owner/repo.
// If a stub was configured by a Given step the server is pointed at the stub;
// otherwise it calls the real GitHub API.
//
// Required params: "owner: <owner>", "repo: <repo>"
func AViewerRequestsRemoteBoard(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("a viewer requests the remote board (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := aViewerRequestsRemoteBoardDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("AViewerRequestsRemoteBoard: %w", err)
			}
			owner := vals.Value("owner")
			repo := vals.Value("repo")

			if err := ensureGitHubAPIServer(ctx); err != nil {
				return err
			}

			httpD := driver.NewHTTPDriver(ctx.ServerURL)
			path := fmt.Sprintf("/remote/board?owner=%s&repo=%s", owner, repo)
			resp, err := httpD.GET(path)
			if err != nil {
				return fmt.Errorf("AViewerRequestsRemoteBoard: GET %s: %w", path, err)
			}
			ctx.LastBody = resp.Body
			ctx.LastStatusCode = resp.StatusCode
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Then steps — observable outcomes
// ---------------------------------------------------------------------------

// BoardAreaShowsMessage asserts that the board area response contains the
// given user-facing message string.
//
// Required param: "message: <expected message text>"
func BoardAreaShowsMessage(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("board area shows message (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := boardAreaShowsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("BoardAreaShowsMessage: %w", err)
			}
			msg := vals.Value("message")
			if ctx.LastBody == "" {
				return fmt.Errorf("BoardAreaShowsMessage: no response recorded; call AViewerRequestsRemoteBoard first")
			}
			if !strings.Contains(ctx.LastBody, msg) {
				return fmt.Errorf("BoardAreaShowsMessage: expected %q in board area response\nActual body:\n%s", msg, ctx.LastBody)
			}
			return nil
		},
	}
}

// RemoteBoardHasColumns asserts that the board area response contains all
// three expected columns by their data-column markers.
func RemoteBoardHasColumns(params ...string) Step {
	return Step{
		Description: "remote board has three columns (Todo, Doing, Done)",
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("RemoteBoardHasColumns: no response recorded; call AViewerRequestsRemoteBoard first")
			}
			for _, col := range []string{"Todo", "Doing", "Done"} {
				marker := fmt.Sprintf(`data-column="%s"`, col)
				if !strings.Contains(ctx.LastBody, marker) {
					return fmt.Errorf("RemoteBoardHasColumns: column %q not found in board response", col)
				}
			}
			return nil
		},
	}
}

// RemoteColumnContainsCards asserts that the named column in the board
// response contains the specified card titles.
//
// Required param: "column: <name>"
// Optional repeating: "title: <card title>"
func RemoteColumnContainsCards(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("remote column contains cards (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := remoteColumnContainsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("RemoteColumnContainsCards: %w", err)
			}
			column := vals.Value("column")
			titles := vals.Values("title")

			if ctx.LastBody == "" {
				return fmt.Errorf("RemoteColumnContainsCards: no response recorded")
			}

			colMarker := fmt.Sprintf(`data-column="%s"`, column)
			colIdx := strings.Index(ctx.LastBody, colMarker)
			if colIdx < 0 {
				return fmt.Errorf("RemoteColumnContainsCards: column %q not found in board response", column)
			}
			colSection := extractColumnSection(ctx.LastBody, colIdx)

			for _, title := range titles {
				cardMarker := fmt.Sprintf(`data-title="%s"`, title)
				if !strings.Contains(colSection, cardMarker) {
					return fmt.Errorf("RemoteColumnContainsCards: card %q not found in column %q\nColumn HTML:\n%s",
						title, column, colSection)
				}
			}
			return nil
		},
	}
}

// RemoteColumnIsEmpty asserts that the named column exists in the board
// response and contains no card elements.
//
// Required param: "column: <name>"
func RemoteColumnIsEmpty(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("remote column is empty (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("RemoteColumnIsEmpty: no response recorded")
			}
			vals, err := columnIsEmptyDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("RemoteColumnIsEmpty: %w", err)
			}
			column := vals.Value("column")
			colMarker := fmt.Sprintf(`data-column="%s"`, column)
			colIdx := strings.Index(ctx.LastBody, colMarker)
			if colIdx < 0 {
				return fmt.Errorf("RemoteColumnIsEmpty: column %q not found in board response", column)
			}
			colSection := extractColumnSection(ctx.LastBody, colIdx)
			if strings.Contains(colSection, `class="card"`) {
				return fmt.Errorf("RemoteColumnIsEmpty: expected column %q to be empty but found cards", column)
			}
			return nil
		},
	}
}

// BoardIsReadOnly asserts that the board response contains no add-task form.
// In github-api mode the board is read-only and must not expose task creation.
func BoardIsReadOnly(params ...string) Step {
	return Step{
		Description: "board is read-only (no add-task form present)",
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("BoardIsReadOnly: no response recorded")
			}
			if strings.Contains(ctx.LastBody, `action="/task"`) || strings.Contains(ctx.LastBody, `id="add-task"`) {
				return fmt.Errorf("BoardIsReadOnly: add-task form found in board response — board should be read-only in github-api mode")
			}
			return nil
		},
	}
}

// URLReflectsOwnerAndRepo asserts that the board fragment embeds the correct
// hx-push-url attribute so the browser URL updates to a shareable link.
//
// Required params: "owner: <owner>", "repo: <repo>"
func URLReflectsOwnerAndRepo(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("URL reflects owner and repo (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := urlContainsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("URLReflectsOwnerAndRepo: %w", err)
			}
			owner := vals.Value("owner")
			repo := vals.Value("repo")

			if ctx.LastBody == "" {
				return fmt.Errorf("URLReflectsOwnerAndRepo: no response recorded")
			}

			expectedPushURL := fmt.Sprintf(`hx-push-url="?owner=%s&repo=%s"`, owner, repo)
			if !strings.Contains(ctx.LastBody, expectedPushURL) {
				return fmt.Errorf("URLReflectsOwnerAndRepo: expected hx-push-url attribute %q not found in response", expectedPushURL)
			}
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// ensureStub starts the GitHub API stub server if it has not already been
// started for this test context. Stores the stub on ctx so subsequent Given
// steps can register repos on it, and records the stub URL so the server
// under test can be pointed at it.
func ensureStub(ctx *WebContext) error {
	if ctx.GitHubAPIStub != nil {
		return nil
	}
	stub := driver.NewGitHubAPIStubDriver()
	ctx.T.Cleanup(stub.Close)
	ctx.GitHubAPIStub = stub
	ctx.GitHubStubURL = stub.URL()
	return nil
}

// ensureGitHubAPIServer starts the kanban-web binary in --mode=github-api
// (if not already running). If a stub was configured by a Given step the
// server is pointed at the stub via --github-api-url; otherwise it calls
// the real GitHub API.
func ensureGitHubAPIServer(ctx *WebContext) error {
	if ctx.ServerURL != "" {
		return nil
	}

	sd := driver.NewServerDriver(ctx.T)
	sd.SetMode("github-api")

	if ctx.GitHubStubURL != "" {
		sd.SetGitHubAPIURL(ctx.GitHubStubURL)
	}

	if err := sd.Build(); err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	port, err := driver.FreePort()
	if err != nil {
		return fmt.Errorf("find free port: %w", err)
	}
	if err := sd.Start(port); err != nil {
		return fmt.Errorf("start server in github-api mode: %w", err)
	}
	ctx.ServerURL = sd.URL()
	return nil
}

// buildTaskFileContent returns the raw Markdown task file content in the
// kanban format (YAML front matter + body).
func buildTaskFileContent(id, title, status, assignee string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "id: %s\n", id)
	fmt.Fprintf(&sb, "title: %s\n", title)
	fmt.Fprintf(&sb, "status: %s\n", status)
	if assignee != "" {
		fmt.Fprintf(&sb, "assignee: %s\n", assignee)
	}
	sb.WriteString("---\n")
	return sb.String()
}

