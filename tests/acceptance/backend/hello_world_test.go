package backend

import (
	"testing"

	"github.com/jmsargent/kanban/tests/acceptance/backend/dsl"
)

// TestHelloWorld_ServerResponds verifies the walking skeleton:
// the kanban-web server starts and responds to board requests.
func TestHelloWorld_ServerResponds(t *testing.T) {
	ctx := dsl.NewWebContext(t)
	dsl.When(ctx, dsl.IVisitTheBoard())
	dsl.Then(ctx, dsl.BoardIsVisible())
}
