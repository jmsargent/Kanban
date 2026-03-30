package dsl

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jmsargent/kanban/pkg/simpledsl"
	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
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
