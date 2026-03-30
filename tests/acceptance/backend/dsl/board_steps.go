package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmsargent/kanban/pkg/simpledsl"
	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

var columnContainsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("column"),
	simpledsl.NewOptionalArg("title").SetAllowMultipleValues(true),
)

var cardAppearsBeforeDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("first"),
	simpledsl.NewRequiredArg("second"),
	simpledsl.NewRequiredArg("column"),
)

// IVisitTheBoard starts the kanban-web server (if not already running) and
// performs a GET /board request, storing the response in the context.
// When ctx.RepoDir is set, the server is started with --repo pointing to it.
func IVisitTheBoard() Step {
	return Step{
		Description: "I visit the board",
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

// ColumnContainsCards asserts that the named column in the last board response
// contains all the specified card titles. Required param: "column: <name>".
// Additional params are "title: <card title>" entries.
//
// The DSL parses ctx.LastBody for:
//
//	<div class="column" data-column="Todo">
//	  <div class="card" data-title="Write docs">Write docs</div>
//	</div>
func ColumnContainsCards(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("column contains cards (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := columnContainsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("ColumnContainsCards: %w", err)
			}
			column := vals.Value("column")
			titles := vals.Values("title")
			if ctx.LastBody == "" {
				return fmt.Errorf("ColumnContainsCards: no board response recorded; call IVisitTheBoard first")
			}

			// Find the column section in the HTML.
			// Expected marker: data-column="<column>"
			colMarker := fmt.Sprintf(`data-column="%s"`, column)
			colIdx := strings.Index(ctx.LastBody, colMarker)
			if colIdx < 0 {
				return fmt.Errorf("column %q not found in board HTML", column)
			}

			// Extract the column div content: from colMarker to the next </div> that
			// closes the column-level div. We look for the next top-level closing div
			// after the column opener.
			colSection := extractColumnSection(ctx.LastBody, colIdx)

			for _, title := range titles {
				cardMarker := fmt.Sprintf(`data-title="%s"`, title)
				if !strings.Contains(colSection, cardMarker) {
					return fmt.Errorf("card %q not found in column %q\nColumn HTML:\n%s", title, column, colSection)
				}
			}
			return nil
		},
	}
}

// CardAppearsBeforeInColumn asserts that the card with title "first" appears
// before the card with title "second" within the named column section of the
// last board response. Required params: "first: <title>", "second: <title>",
// "column: <column label>".
func CardAppearsBeforeInColumn(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("card appears before in column (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := cardAppearsBeforeDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("CardAppearsBeforeInColumn: %w", err)
			}
			first := vals.Value("first")
			second := vals.Value("second")
			column := vals.Value("column")

			if ctx.LastBody == "" {
				return fmt.Errorf("CardAppearsBeforeInColumn: no board response; call IVisitTheBoard first")
			}

			colMarker := fmt.Sprintf(`data-column="%s"`, column)
			colIdx := strings.Index(ctx.LastBody, colMarker)
			if colIdx < 0 {
				return fmt.Errorf("column %q not found in board HTML", column)
			}

			colSection := extractColumnSection(ctx.LastBody, colIdx)

			firstMarker := fmt.Sprintf(`data-title="%s"`, first)
			secondMarker := fmt.Sprintf(`data-title="%s"`, second)

			firstIdx := strings.Index(colSection, firstMarker)
			if firstIdx < 0 {
				return fmt.Errorf("card %q not found in column %q", first, column)
			}
			secondIdx := strings.Index(colSection, secondMarker)
			if secondIdx < 0 {
				return fmt.Errorf("card %q not found in column %q", second, column)
			}

			if firstIdx >= secondIdx {
				return fmt.Errorf("expected card %q (pos %d) to appear before %q (pos %d) in column %q",
					first, firstIdx, second, secondIdx, column)
			}
			return nil
		},
	}
}

// extractColumnSection returns the HTML content of the column div starting at
// startIdx. It finds the opening <div ... data-column="..."> and returns
// everything up to and including the matching closing </div>.
func extractColumnSection(html string, startIdx int) string {
	// Find the start of the opening tag that contains the data-column marker.
	// Walk backwards from startIdx to find the '<'.
	openTag := startIdx
	for openTag > 0 && html[openTag] != '<' {
		openTag--
	}

	// Now walk forward counting open/close div tags to find the matching close.
	depth := 0
	i := openTag
	n := len(html)
	for i < n {
		if strings.HasPrefix(html[i:], "<div") {
			depth++
			i += 4
			continue
		}
		if strings.HasPrefix(html[i:], "</div>") {
			depth--
			i += 6
			if depth == 0 {
				return html[openTag:i]
			}
			continue
		}
		i++
	}
	// Fallback: return everything from the column marker to end.
	return html[openTag:]
}
