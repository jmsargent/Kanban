package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

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
