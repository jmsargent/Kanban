package dsl_test

import (
	"testing"
	"time"

	"github.com/jmsargent/kanban/tests/acceptance/backend/dsl"
)

// Test budget: 1 behavior (duration threshold check) × 2 = 2 unit tests max.
// Using parametrize: 1 test covering both within-threshold and over-threshold cases.

func TestBoardLoadsWithin_EnforcesThreshold(t *testing.T) {
	cases := []struct {
		name         string
		lastDuration time.Duration
		timeout      string
		expectError  bool
	}{
		{
			name:         "passes when duration is under threshold",
			lastDuration: 100 * time.Millisecond,
			timeout:      "timeout: 500ms",
			expectError:  false,
		},
		{
			name:         "fails when duration exceeds threshold",
			lastDuration: 600 * time.Millisecond,
			timeout:      "timeout: 500ms",
			expectError:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := dsl.NewWebContext(t)
			ctx.LastDuration = tc.lastDuration

			step := dsl.BoardLoadsWithin(tc.timeout)
			err := step.Run(ctx)

			if tc.expectError && err == nil {
				t.Errorf("expected error for duration %v > %s, got nil", tc.lastDuration, tc.timeout)
			}
			if !tc.expectError && err != nil {
				t.Errorf("expected no error for duration %v <= %s, got: %v", tc.lastDuration, tc.timeout, err)
			}
		})
	}
}