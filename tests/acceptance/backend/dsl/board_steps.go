package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

// IVisitTheBoard starts the kanban-web server (if not already running) and
// performs a GET /board request, storing the response in the context.
func IVisitTheBoard() Step {
	return Step{
		Description: "I visit the board",
		Run: func(ctx *WebContext) error {
			if ctx.ServerURL == "" {
				sd := driver.NewServerDriver(ctx.T)
				if err := sd.Build(); err != nil {
					return fmt.Errorf("build server: %w", err)
				}
				port := 18080
				if err := sd.Start(port); err != nil {
					return fmt.Errorf("start server: %w", err)
				}
				ctx.ServerURL = sd.URL()
			}

			httpDriver := driver.NewHTTPDriver(ctx.ServerURL)
			start := time.Now()
			resp, err := httpDriver.GET("/board")
			ctx.LastDuration = time.Since(start)
			if err != nil {
				return fmt.Errorf("GET /board: %w", err)
			}

			ctx.LastBody = resp.Body
			return nil
		},
	}
}

// BoardIsVisible asserts that the last response contains a non-empty board page.
func BoardIsVisible() Step {
	return Step{
		Description: "board is visible",
		Run: func(ctx *WebContext) error {
			if ctx.LastBody == "" {
				return fmt.Errorf("board response body is empty")
			}
			if strings.Contains(ctx.LastBody, "404") || strings.Contains(ctx.LastBody, "page not found") {
				return fmt.Errorf("board returned error page: %q", ctx.LastBody)
			}
			return nil
		},
	}
}
