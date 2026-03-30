package dsl

import (
	"fmt"
	"time"
)

// Assertion step factories for web acceptance tests.
// These verify observable outcomes: board state, card content, auth state.

// BoardLoadsWithin asserts that the last board request completed within the
// given duration threshold.
func BoardLoadsWithin(threshold time.Duration) Step {
	return Step{
		Description: fmt.Sprintf("board loads within %v", threshold),
		Run: func(ctx *WebContext) error {
			if ctx.LastDuration == 0 {
				return fmt.Errorf("no request duration recorded; call IVisitTheBoard first")
			}
			if ctx.LastDuration > threshold {
				return fmt.Errorf("board response took %v, exceeded threshold of %v", ctx.LastDuration, threshold)
			}
			return nil
		},
	}
}
