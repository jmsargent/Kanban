package dsl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jmsargent/kanban/pkg/simpledsl"
	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

var iViewCardDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
)

var cardShowsDSL = simpledsl.NewDslParams(
	simpledsl.NewOptionalArg("title"),
	simpledsl.NewOptionalArg("assignee"),
	simpledsl.NewOptionalArg("status"),
)

// IViewCard navigates from the board to the card detail page for the card
// matching the given title. Requires IVisitTheBoard to have been called first.
func IViewCard(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I view card (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := iViewCardDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IViewCard: %w", err)
			}
			title := vals.Value("title")
			if ctx.LastBody == "" {
				return fmt.Errorf("IViewCard: no board response recorded; call IVisitTheBoard first")
			}
			if ctx.ServerURL == "" {
				return fmt.Errorf("IViewCard: server URL not set; call IVisitTheBoard first")
			}

			href, err := findCardHref(ctx.LastBody, title)
			if err != nil {
				return fmt.Errorf("IViewCard: %w", err)
			}

			httpDriver := driver.NewHTTPDriver(ctx.ServerURL)
			resp, err := httpDriver.GET(href)
			if err != nil {
				return fmt.Errorf("IViewCard: GET %s: %w", href, err)
			}
			ctx.LastBody = resp.Body
			return nil
		},
	}
}

// CardShows asserts that the card detail page shows the given field values.
// Supported params: "title", "assignee", "status".
func CardShows(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("card shows (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := cardShowsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("CardShows: %w", err)
			}
			if ctx.LastBody == "" {
				return fmt.Errorf("CardShows: no card detail response recorded; call IViewCard first")
			}

			for _, field := range []string{"title", "assignee", "status"} {
				expected := vals.Value(field)
				if expected == "" {
					continue
				}
				if err := assertFieldValue(ctx.LastBody, field, expected); err != nil {
					return fmt.Errorf("CardShows: %w", err)
				}
			}
			return nil
		},
	}
}

// findCardHref extracts the href of the anchor inside the card with the given
// data-title attribute. Expects: <div class="card" data-title="..."><a href="...">
func findCardHref(body, title string) (string, error) {
	// Match <div ... data-title="<title>">...<a href="<href>">
	pattern := fmt.Sprintf(`data-title="%s"[^>]*>[\s\S]*?<a\s+href="([^"]+)"`, regexp.QuoteMeta(title))
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("compile card href pattern: %w", err)
	}
	m := re.FindStringSubmatch(body)
	if m == nil {
		return "", fmt.Errorf("no card link found for title %q", title)
	}
	return m[1], nil
}

// assertFieldValue checks that the body contains a data-field="<field>" element
// whose text content includes expected.
func assertFieldValue(body, field, expected string) error {
	marker := fmt.Sprintf(`data-field="%s"`, field)
	idx := strings.Index(body, marker)
	if idx < 0 {
		return fmt.Errorf("field %q not found in card detail", field)
	}
	// Extract content from the marker to the next closing tag.
	rest := body[idx:]
	closeIdx := strings.Index(rest, "</")
	if closeIdx < 0 {
		closeIdx = len(rest)
	}
	section := rest[:closeIdx]
	if !strings.Contains(section, expected) {
		return fmt.Errorf("field %q: expected %q, got section: %q", field, expected, section)
	}
	return nil
}
