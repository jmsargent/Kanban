package dsl

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"time"
)

// taskIDPattern matches TASK-NNN identifiers in command output.
var taskIDPattern = regexp.MustCompile(`TASK-\d+`)

// captureResult writes the stdout, stderr, exit code, and last task ID from cmd
// into ctx. cmd must already have its Stdout and Stderr set to the provided buffers.
func captureResult(ctx *Context, err error, stdout, stderr *bytes.Buffer) {
	ctx.lastStdout = stdout.String()
	ctx.lastStderr = stderr.String()
	ctx.lastOutput = ctx.lastStdout + ctx.lastStderr
	ctx.lastExit = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		ctx.lastExit = exitErr.ExitCode()
	} else if err != nil {
		ctx.lastExit = 1
	}
	if match := taskIDPattern.FindString(ctx.lastOutput); match != "" {
		ctx.lastTaskID = match
	}
}

// run executes the kanban binary with the given args in ctx.repoDir.
// It captures stdout and stderr separately, records the exit code, and
// extracts the first TASK-NNN match into ctx.lastTaskID.
func run(ctx *Context, args ...string) {
	if ctx.t != nil {
		ctx.t.Helper()
	}
	cmdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, ctx.binPath, args...)
	cmd.Dir = ctx.repoDir
	cmd.Env = ctx.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	captureResult(ctx, cmd.Run(), &stdout, &stderr)
}
