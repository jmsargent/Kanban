package dsl

// Step pairs a human-readable description with an executable function.
type Step struct {
	Description string
	Run         func(*WebContext) error
}

// Given executes a setup step. On non-nil error it calls t.Fatalf.
func Given(ctx *WebContext, step Step) {
	ctx.T.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.T.Fatalf("Given: %s: %v", step.Description, err)
	}
}

// When executes an action step. On non-nil error it calls t.Fatalf.
func When(ctx *WebContext, step Step) {
	ctx.T.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.T.Fatalf("When: %s: %v", step.Description, err)
	}
}

// Then executes an assertion step. On non-nil error it calls t.Fatalf.
func Then(ctx *WebContext, step Step) {
	ctx.T.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.T.Fatalf("Then: %s: %v", step.Description, err)
	}
}

// And is an alias for Then, used for readability in multi-step assertions.
func And(ctx *WebContext, step Step) {
	ctx.T.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.T.Fatalf("And: %s: %v", step.Description, err)
	}
}
