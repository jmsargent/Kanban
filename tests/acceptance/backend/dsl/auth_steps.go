package dsl

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jmsargent/kanban/pkg/simpledsl"
	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

var iAttemptToAuthenticateDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("token"),
	simpledsl.NewRequiredArg("display_name"),
)

var withGitHubStubDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("token"),
	simpledsl.NewRequiredArg("login"),
	simpledsl.NewRequiredArg("display_name"),
)

var iAuthenticateDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("token"),
	simpledsl.NewRequiredArg("display_name"),
)

// WithGitHubStub starts a GitHub API stub server and registers a valid token on
// it. The stub's base URL is stored in ctx.GitHubStubURL so that subsequent
// steps (IAuthenticate) can start the kanban-web server configured to use the
// stub instead of the real api.github.com.
// Required params: "token: <token>", "login: <github_login>", "display_name: <name>".
func WithGitHubStub(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("with GitHub stub (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := withGitHubStubDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("WithGitHubStub: %w", err)
			}
			stub := driver.NewGitHubStubDriver()
			ctx.T.Cleanup(stub.Close)
			stub.AddValidToken(vals.Value("token"), vals.Value("login"), vals.Value("display_name"))
			ctx.GitHubStubURL = stub.URL()
			return nil
		},
	}
}

// ensureServerWithGitHubStub starts the kanban-web server if not already running,
// injecting ctx.GitHubStubURL so the server validates tokens against the stub.
func ensureServerWithGitHubStub(ctx *WebContext) error {
	if ctx.ServerURL != "" {
		return nil
	}
	sd := driver.NewServerDriver(ctx.T)
	if err := sd.Build(); err != nil {
		return fmt.Errorf("build server: %w", err)
	}
	if ctx.RepoDir != "" {
		sd.SetRepoDir(ctx.RepoDir)
	}
	if ctx.GitHubStubURL != "" {
		sd.SetGitHubAPIURL(ctx.GitHubStubURL)
	}
	port, err := driver.FreePort()
	if err != nil {
		return fmt.Errorf("find free port: %w", err)
	}
	if err := sd.Start(port); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	ctx.ServerURL = sd.URL()
	return nil
}

// IAuthenticate POSTs to /auth/token with the given token and display_name.
// It starts the server (with the GitHub stub if configured), stores the
// stateful HTTP driver on ctx.HTTPDriver (so subsequent steps share the same
// cookie jar), and records the final response body in ctx.LastBody.
// Required params: "token: <token>", "display_name: <name>".
func IAuthenticate(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I authenticate (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := iAuthenticateDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IAuthenticate: %w", err)
			}
			if err := ensureServerWithGitHubStub(ctx); err != nil {
				return err
			}
			httpDriver := driver.NewHTTPDriver(ctx.ServerURL)
			ctx.HTTPDriver = httpDriver // share cookie jar across subsequent steps
			start := time.Now()
			formData := url.Values{
				"token":        {vals.Value("token")},
				"display_name": {vals.Value("display_name")},
			}
			resp, err := httpDriver.POST("/auth/token", formData)
			ctx.LastDuration = time.Since(start)
			if err != nil {
				return fmt.Errorf("POST /auth/token: %w", err)
			}
			ctx.LastBody = resp.Body
			return nil
		},
	}
}

// IAmOnTheBoard asserts that the current page (ctx.LastBody) is the board page.
func IAmOnTheBoard(params ...string) Step {
	return Step{
		Description: "I am on the board",
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("IAmOnTheBoard: no response recorded; call IAuthenticate first")
			}
			if !strings.Contains(ctx.LastBody, "Kanban Board") {
				return fmt.Errorf("IAmOnTheBoard: expected board page but got:\n%s", ctx.LastBody)
			}
			return nil
		},
	}
}

// ICanAddTasks asserts that GET /task/new does not redirect to the auth page,
// confirming the authenticated session grants write access. It reuses the HTTP
// driver stored by IAuthenticate so the session cookie is sent automatically.
func ICanAddTasks(params ...string) Step {
	return Step{
		Description: "I can add tasks",
		Run: func(ctx *WebContext) error {
			if ctx.HTTPDriver == nil {
				return fmt.Errorf("ICanAddTasks: no authenticated HTTP driver; call IAuthenticate first")
			}
			resp, err := ctx.HTTPDriver.GET("/task/new")
			if err != nil {
				return fmt.Errorf("ICanAddTasks: GET /task/new: %w", err)
			}
			if strings.Contains(resp.Body, "token-entry") {
				return fmt.Errorf("ICanAddTasks: reached auth form instead of add-task form — not authenticated")
			}
			return nil
		},
	}
}

// ITryToAddTask makes an unauthenticated GET request to /task/new (the add-task
// path). The response is recorded in ctx.LastBody and ctx.LastResponse.
func ITryToAddTask(params ...string) Step {
	return Step{
		Description: "I try to add a task",
		Run: func(ctx *WebContext) error {
			if ctx.ServerURL == "" {
				sd := driver.NewServerDriver(ctx.T)
				if err := sd.Build(); err != nil {
					return fmt.Errorf("build server: %w", err)
				}
				if ctx.RepoDir != "" {
					sd.SetRepoDir(ctx.RepoDir)
				}
				port, err := driver.FreePort()
				if err != nil {
					return fmt.Errorf("find free port: %w", err)
				}
				if err := sd.Start(port); err != nil {
					return fmt.Errorf("start server: %w", err)
				}
				ctx.ServerURL = sd.URL()
			}

			// Use a plain HTTP client that does NOT follow redirects so we can
			// inspect the redirect response. We capture the final destination body.
			httpDriver := driver.NewHTTPDriver(ctx.ServerURL)
			start := time.Now()
			resp, err := httpDriver.GET("/task/new")
			ctx.LastDuration = time.Since(start)
			if err != nil {
				return fmt.Errorf("GET /task/new: %w", err)
			}

			ctx.LastBody = resp.Body
			return nil
		},
	}
}

// PromptedToAuthenticate asserts that the last response body contains the
// token entry form, confirming the user was redirected to authenticate.
func PromptedToAuthenticate(params ...string) Step {
	return Step{
		Description: "prompted to authenticate",
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("PromptedToAuthenticate: no response recorded; call ITryToAddTask first")
			}
			if !strings.Contains(ctx.LastBody, "token-entry") {
				return fmt.Errorf("expected token entry form but got:\n%s", ctx.LastBody)
			}
			return nil
		},
	}
}

// AnUnauthorizedUser is a context step that documents that the current user
// holds an invalid GitHub token. No server-side setup is needed because the
// GitHub stub rejects any token that was not registered via AddValidToken.
func AnUnauthorizedUser(params ...string) Step {
	return Step{
		Description: "an unauthorized user (invalid token)",
		Run: func(ctx *WebContext) error {
			// The GitHubStubDriver already rejects unknown tokens with 401.
			// This step is intentionally a no-op: it exists to make the test
			// readable and to document the pre-condition explicitly.
			return nil
		},
	}
}

// IAttemptToAuthenticate POSTs to /auth/token with credentials that are
// expected to be rejected. It starts the server (with the GitHub stub if
// configured), stores the HTTP driver on ctx.HTTPDriver, and records the
// response body in ctx.LastBody.
// Required params: "token: <token>", "display_name: <name>".
func IAttemptToAuthenticate(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I attempt to authenticate (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := iAttemptToAuthenticateDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IAttemptToAuthenticate: %w", err)
			}
			if err := ensureServerWithGitHubStub(ctx); err != nil {
				return err
			}
			httpDriver := driver.NewHTTPDriver(ctx.ServerURL)
			ctx.HTTPDriver = httpDriver
			formData := url.Values{
				"token":        {vals.Value("token")},
				"display_name": {vals.Value("display_name")},
			}
			resp, err := httpDriver.POST("/auth/token", formData)
			if err != nil {
				return fmt.Errorf("POST /auth/token: %w", err)
			}
			ctx.LastBody = resp.Body
			return nil
		},
	}
}

// AuthenticationIsRejected asserts that the last response is the token entry
// form (id="token-entry") and contains an error indicator (id="auth-error"),
// confirming the submitted token was rejected.
func AuthenticationIsRejected(params ...string) Step {
	return Step{
		Description: "authentication is rejected",
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("AuthenticationIsRejected: no response recorded; call IAttemptToAuthenticate first")
			}
			if !strings.Contains(ctx.LastBody, "token-entry") {
				return fmt.Errorf("AuthenticationIsRejected: expected token entry form but got:\n%s", ctx.LastBody)
			}
			if !strings.Contains(ctx.LastBody, "auth-error") {
				return fmt.Errorf("AuthenticationIsRejected: expected error message in form but got:\n%s", ctx.LastBody)
			}
			return nil
		},
	}
}

// ICannotAddTasks asserts that GET /task/new redirects back to the auth form,
// confirming no session cookie was set after a failed authentication attempt.
// It reuses ctx.HTTPDriver so any cookies from the failed POST are sent.
func ICannotAddTasks(params ...string) Step {
	return Step{
		Description: "I cannot add tasks",
		Run: func(ctx *WebContext) error {
			if ctx.HTTPDriver == nil {
				return fmt.Errorf("ICannotAddTasks: no HTTP driver; call IAttemptToAuthenticate first")
			}
			resp, err := ctx.HTTPDriver.GET("/task/new")
			if err != nil {
				return fmt.Errorf("ICannotAddTasks: GET /task/new: %w", err)
			}
			if !strings.Contains(resp.Body, "token-entry") {
				return fmt.Errorf("ICannotAddTasks: expected auth form redirect but got:\n%s", resp.Body)
			}
			return nil
		},
	}
}

var anAuthenticatedUserDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("token"),
	simpledsl.NewRequiredArg("display_name"),
)

// AnAuthenticatedUser performs the full authentication flow and stores the
// stateful HTTPDriver (with session cookie) on ctx.HTTPDriver so subsequent
// steps automatically send the session cookie. This step is the setup
// counterpart to IAuthenticate, named from the user's perspective for Given
// clauses. Required params: "token: <token>", "display_name: <name>".
func AnAuthenticatedUser(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("an authenticated user (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := anAuthenticatedUserDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("AnAuthenticatedUser: %w", err)
			}
			if err := ensureServerWithGitHubStub(ctx); err != nil {
				return err
			}
			httpDriver := driver.NewHTTPDriver(ctx.ServerURL)
			ctx.HTTPDriver = httpDriver
			formData := url.Values{
				"token":        {vals.Value("token")},
				"display_name": {vals.Value("display_name")},
			}
			resp, err := httpDriver.POST("/auth/token", formData)
			if err != nil {
				return fmt.Errorf("AnAuthenticatedUser: POST /auth/token: %w", err)
			}
			ctx.LastBody = resp.Body
			return nil
		},
	}
}

// AddTaskFormIsShown asserts that GET /task/new returns the add-task form
// (not a redirect to the authentication page), confirming the authenticated
// session is still active on a subsequent request. It reuses ctx.HTTPDriver
// so the session cookie is sent automatically.
func AddTaskFormIsShown(params ...string) Step {
	return Step{
		Description: "add task form is shown",
		Run: func(ctx *WebContext) error {
			if ctx.HTTPDriver == nil {
				return fmt.Errorf("AddTaskFormIsShown: no authenticated HTTP driver; call AnAuthenticatedUser first")
			}
			resp, err := ctx.HTTPDriver.GET("/task/new")
			if err != nil {
				return fmt.Errorf("AddTaskFormIsShown: GET /task/new: %w", err)
			}
			if strings.Contains(resp.Body, "token-entry") {
				return fmt.Errorf("AddTaskFormIsShown: reached auth form instead of add-task form — session not persisted")
			}
			return nil
		},
	}
}

// AddTaskOptionIsVisible asserts that the board HTML contains an "Add Task"
// link or button, confirming the option is visible even to unauthenticated users.
func AddTaskOptionIsVisible(params ...string) Step {
	return Step{
		Description: "add task option is visible",
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("AddTaskOptionIsVisible: no board response recorded; call IVisitTheBoard first")
			}
			if !strings.Contains(ctx.LastBody, "add-task") {
				return fmt.Errorf("expected 'add-task' element in board HTML but not found")
			}
			return nil
		},
	}
}
