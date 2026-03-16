package dsl

// Step pairs a human-readable description with an executable function.
type Step struct {
	Description string
	Run         func(*Context) error
}

// Given executes a setup step. On non-nil error it calls t.Fatalf.
func Given(ctx *Context, step Step) {
	ctx.t.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.t.Fatalf("Given: %s: %v", step.Description, err)
	}
}

// When executes an action step. On non-nil error it calls t.Fatalf.
func When(ctx *Context, step Step) {
	ctx.t.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.t.Fatalf("When: %s: %v", step.Description, err)
	}
}

// Then executes an assertion step. On non-nil error it calls t.Fatalf.
func Then(ctx *Context, step Step) {
	ctx.t.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.t.Fatalf("Then: %s: %v", step.Description, err)
	}
}

// And is an alias for Then, used for readability in multi-step assertions.
func And(ctx *Context, step Step) {
	ctx.t.Helper()
	if err := step.Run(ctx); err != nil {
		ctx.t.Fatalf("And: %s: %v", step.Description, err)
	}
}
