package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmsargent/kanban/pkg/simpledsl"
)

// Assertion step factories for web acceptance tests.
// These verify observable outcomes: board state, card content, auth state.

var boardLoadsWithinDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("timeout"),
)

// BoardLoadsWithin asserts that the last board request completed within the
// given duration threshold. Required param: "timeout: <duration>" e.g. "timeout: 500ms".
func BoardLoadsWithin(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("board loads within (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := boardLoadsWithinDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("BoardLoadsWithin: %w", err)
			}
			threshold, err := time.ParseDuration(vals.Value("timeout"))
			if err != nil {
				return fmt.Errorf("BoardLoadsWithin: invalid timeout %q: %w", vals.Value("timeout"), err)
			}
			if ctx.LastDuration == 0 {
				return fmt.Errorf("BoardLoadsWithin: no request duration recorded; call IVisitTheBoard first")
			}
			if ctx.LastDuration > threshold {
				return fmt.Errorf("BoardLoadsWithin: board response took %v, exceeded threshold of %v", ctx.LastDuration, threshold)
			}
			return nil
		},
	}
}
