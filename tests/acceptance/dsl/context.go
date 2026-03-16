package dsl

import (
	"os"
	"path/filepath"
	"testing"
)

// Context holds the mutable state for a single acceptance test scenario.
// It invokes the kanban binary as a subprocess only — no direct internal/ calls.
//
// Fields repoDir, lastStdout, lastStderr, lastOutput, and lastExit are populated
// by run() and read by step factories added in subsequent steps (01-02+).
//
//nolint:unused
type Context struct {
	t          *testing.T
	repoDir    string // set by InAGitRepo setup step
	binPath    string // resolved at construction
	lastStdout string
	lastStderr string
	lastOutput string // lastStdout + lastStderr
	lastExit   int
	lastTaskID string // most recent TASK-NNN from output
	env        []string
}

// NewContext constructs a Context with a resolved binary path and the current
// process environment. It does NOT create a temp repo — use the InAGitRepo
// setup step for that.
func NewContext(t *testing.T) *Context {
	t.Helper()
	bin := os.Getenv("KANBAN_BIN")
	if bin == "" {
		abs, err := filepath.Abs("../../bin/kanban")
		if err != nil {
			abs = "../../bin/kanban"
		}
		bin = abs
	}
	return &Context{
		t:       t,
		binPath: bin,
		env:     os.Environ(),
	}
}

// LastTaskID returns the most recently captured TASK-NNN identifier.
func (ctx *Context) LastTaskID() string {
	return ctx.lastTaskID
}

// RepoDir returns the absolute path of the temporary git repository for this scenario.
// Primarily used by test helpers that need to inspect filesystem state directly.
func (ctx *Context) RepoDir() string {
	return ctx.repoDir
}
