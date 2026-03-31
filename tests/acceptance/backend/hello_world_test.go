package backend

import (
	"testing"
)

import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"


// TestHelloWorld_ServerResponds verifies the walking skeleton:
// the kanban-web server starts and responds to board requests.
func TestHelloWorld_ServerResponds(t *testing.T) {
	ctx := NewWebContext(t)
	When(ctx, IVisitTheBoard())
	Then(ctx, BoardIsVisible())
}

// TestHelloWorld_ResponseTime verifies the board page loads within the
// 500ms performance threshold.
func TestHelloWorld_ResponseTime(t *testing.T) {
	ctx := NewWebContext(t)
	When(ctx, IVisitTheBoard())
	Then(ctx, BoardLoadsWithin("timeout: 500ms"))
}
