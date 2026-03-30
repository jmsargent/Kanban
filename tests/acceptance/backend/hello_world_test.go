package backend

import (
	"testing"
	"time"

	"github.com/jmsargent/kanban/tests/acceptance/backend/dsl"
)

// TestHelloWorld_ServerResponds verifies the walking skeleton:
// the kanban-web server starts and responds to board requests.
func TestHelloWorld_ServerResponds(t *testing.T) {
	ctx := dsl.NewWebContext(t)
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.Then(ctx, dsl.BoardIsVisible())
}

// TestHelloWorld_ResponseTime verifies the board page loads within the
// 500ms performance threshold.
func TestHelloWorld_ResponseTime(t *testing.T) {
	ctx := dsl.NewWebContext(t)
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.Then(ctx, dsl.BoardLoadsWithin(500*time.Millisecond))
}
