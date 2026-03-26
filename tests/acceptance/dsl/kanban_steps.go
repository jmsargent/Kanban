package dsl

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ---- Extended step functions for board-state-in-git feature ----
//
// These step factories extend the existing DSL (Context, Step, Given/When/Then/And)
// to cover the three user stories: US-BSG-01 (kanban log), US-BSG-02
// (transitions log), and US-BSG-03 (kanban board --me).
//
// All steps invoke the kanban binary as a subprocess (driving port). The only
// direct filesystem access permitted is reading .kanban/transitions.log for
// structural assertions (line counts, field format) — an observable file output.

// ---- Setup steps ----

// TransitionsLogExists asserts that .kanban/transitions.log is present in the
// repository. Use this as a pre-condition for scenarios that require existing log
// entries written by prior steps.
func TransitionsLogExists() Step {
	return Step{
		Description: "transitions log file exists",
		Run: func(ctx *Context) error {
			logPath := transitionsLogPath(ctx)
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				return fmt.Errorf("expected .kanban/transitions.log to exist but it does not")
			}
			return nil
		},
	}
}

// HookInstalled runs "kanban install-hook" (or "kanban init" which is idempotent
// and installs the commit-msg hook) and fails if the command exits non-zero.
// This step is a semantic alias for CommitHookInstalled with a name that
// reflects the US-BSG-02 vocabulary.
func HookInstalled() Step {
	return Step{
		Description: "commit-msg hook installed",
		Run: func(ctx *Context) error {
			run(ctx, "init")
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban init (install hook) failed (exit %d): %s", ctx.lastExit, ctx.lastOutput)
			}
			return nil
		},
	}
}

// TaskCreatedViaAdd runs "kanban new <title>", captures the resulting TASK-NNN
// into ctx.lastTaskID, and fails if the command exits non-zero.
func TaskCreatedViaAdd(title string) Step {
	return Step{
		Description: fmt.Sprintf("task created with title %q", title),
		Run: func(ctx *Context) error {
			run(ctx, "new", title)
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban new failed (exit %d): %s", ctx.lastExit, ctx.lastOutput)
			}
			if ctx.lastTaskID == "" {
				return fmt.Errorf("kanban new did not produce a TASK-NNN in output: %s", ctx.lastOutput)
			}
			return nil
		},
	}
}

// TaskStarted runs "kanban start <taskID>" and fails if the command exits non-zero.
func TaskStarted(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("task %s started", taskID),
		Run: func(ctx *Context) error {
			run(ctx, "start", taskID)
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban start %s failed (exit %d): %s", taskID, ctx.lastExit, ctx.lastOutput)
			}
			return nil
		},
	}
}

// GitCommitReferencingTask stages all changes and creates a git commit whose
// message contains taskID. This simulates a developer committing work and
// exercises the commit-msg hook path (if installed).
func GitCommitReferencingTask(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("git commit referencing %s", taskID),
		Run: func(ctx *Context) error {
			if _, err := gitCmd(ctx, "add", "-A"); err != nil {
				return fmt.Errorf("git add -A: %w", err)
			}
			if _, err := gitCmd(ctx, "commit", "--allow-empty", "-m", taskID+": work in progress"); err != nil {
				return fmt.Errorf("git commit referencing %s: %w", taskID, err)
			}
			return nil
		},
	}
}

// MultipleTasksExist creates n tasks via "kanban new <title>", accumulating
// them under distinct titles. It stores only the last task ID in ctx.lastTaskID.
func MultipleTasksExist(titles ...string) Step {
	return Step{
		Description: fmt.Sprintf("tasks exist: %s", strings.Join(titles, ", ")),
		Run: func(ctx *Context) error {
			for _, title := range titles {
				run(ctx, "new", title)
				if ctx.lastExit != 0 {
					return fmt.Errorf("kanban new %q failed (exit %d): %s", title, ctx.lastExit, ctx.lastOutput)
				}
			}
			return nil
		},
	}
}

// TaskAssignedTo runs "kanban new <title> --assignee <email>" and stores
// the new task ID in ctx.lastTaskID. Used in US-BSG-03 board --me scenarios.
func TaskAssignedTo(title, assignee string) Step {
	return Step{
		Description: fmt.Sprintf("task %q assigned to %s", title, assignee),
		Run: func(ctx *Context) error {
			run(ctx, "new", title, "--assignee", assignee)
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban new --assignee failed (exit %d): %s", ctx.lastExit, ctx.lastOutput)
			}
			return nil
		},
	}
}

// TransitionsLogMadeUnwritable changes .kanban/transitions.log permissions to
// 0o000. A t.Cleanup is registered to restore 0o644. Used to test error-path
// behaviour when the log cannot be written.
func TransitionsLogMadeUnwritable() Step {
	return Step{
		Description: "transitions log is not writable",
		Run: func(ctx *Context) error {
			logPath := transitionsLogPath(ctx)
			// Ensure the file exists so we can chmod it.
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				if err2 := os.MkdirAll(filepath.Dir(logPath), 0o755); err2 != nil {
					return fmt.Errorf("mkdir .kanban: %w", err2)
				}
				if err2 := os.WriteFile(logPath, nil, 0o644); err2 != nil {
					return fmt.Errorf("create transitions.log: %w", err2)
				}
			}
			ctx.t.Cleanup(func() {
				_ = os.Chmod(logPath, 0o644)
			})
			return os.Chmod(logPath, 0o000)
		},
	}
}

// ---- Action steps ----

// DeveloperRunsKanbanLog runs "kanban log <taskID>" and captures output/exit.
func DeveloperRunsKanbanLog(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("developer runs kanban log %s", taskID),
		Run: func(ctx *Context) error {
			run(ctx, "log", taskID)
			return nil
		},
	}
}

// DeveloperRunsKanbanStart runs "kanban start <taskID>" and captures output/exit.
func DeveloperRunsKanbanStart(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("developer runs kanban start %s", taskID),
		Run: func(ctx *Context) error {
			run(ctx, "start", taskID)
			return nil
		},
	}
}

// DeveloperRunsKanbanBoard runs "kanban board" and captures output/exit.
func DeveloperRunsKanbanBoard() Step {
	return Step{
		Description: "developer runs kanban board",
		Run: func(ctx *Context) error {
			run(ctx, "board")
			return nil
		},
	}
}

// DeveloperRunsKanbanBoardWithMe runs "kanban board --me" and captures output/exit.
func DeveloperRunsKanbanBoardWithMe() Step {
	return Step{
		Description: "developer runs kanban board --me",
		Run: func(ctx *Context) error {
			run(ctx, "board", "--me")
			return nil
		},
	}
}

// DeveloperRunsKanbanCiDone runs "kanban ci-done --since <sha>" and captures
// output/exit. Pass an empty string to omit the --since flag.
func DeveloperRunsKanbanCiDone(since string) Step {
	return Step{
		Description: fmt.Sprintf("developer runs kanban ci-done since %q", since),
		Run: func(ctx *Context) error {
			if since == "" {
				run(ctx, "ci-done")
			} else {
				run(ctx, "ci-done", "--since", since)
			}
			return nil
		},
	}
}

// ConcurrentStartsOnSameTask launches n goroutines each running "kanban start
// <taskID>" simultaneously, waits for all to complete, and captures the last
// observed exit code. Used for concurrency safety tests in US-BSG-02.
func ConcurrentStartsOnSameTask(taskID string, n int) Step {
	return Step{
		Description: fmt.Sprintf("%d concurrent starts on task %s", n, taskID),
		Run: func(ctx *Context) error {
			var wg sync.WaitGroup
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					cmdCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					cmd := exec.CommandContext(cmdCtx, ctx.binPath, "start", taskID)
					cmd.Dir = ctx.repoDir
					cmd.Env = ctx.env
					var stdout, stderr bytes.Buffer
					cmd.Stdout = &stdout
					cmd.Stderr = &stderr
					_ = cmd.Run()
				}()
			}
			wg.Wait()
			return nil
		},
	}
}

// ---- Assertion steps ----

// OutputDoesNotContain asserts ctx.lastOutput does not contain text.
func OutputDoesNotContain(text string) Step {
	return Step{
		Description: fmt.Sprintf("output does not contain %q", text),
		Run: func(ctx *Context) error {
			if strings.Contains(ctx.lastOutput, text) {
				return fmt.Errorf("expected output NOT to contain %q\nOutput:\n%s", text, ctx.lastOutput)
			}
			return nil
		},
	}
}

// ExitsSuccessfully asserts ctx.lastExit == 0.
func ExitsSuccessfully() Step {
	return Step{
		Description: "command exits successfully",
		Run: func(ctx *Context) error {
			if ctx.lastExit != 0 {
				return fmt.Errorf("expected exit 0, got %d\nOutput:\n%s", ctx.lastExit, ctx.lastOutput)
			}
			return nil
		},
	}
}

// ExitsWithCode asserts ctx.lastExit == code.
func ExitsWithCode(code int) Step {
	return Step{
		Description: fmt.Sprintf("command exits with code %d", code),
		Run:         ExitCodeIs(code).Run,
	}
}

// TransitionsLogLineCountIs reads .kanban/transitions.log and asserts it has
// exactly n non-empty lines. Acceptable for structural assertions per the
// Port-to-Port principle note in the configuration.
func TransitionsLogLineCountIs(n int) Step {
	return Step{
		Description: fmt.Sprintf("transitions log has %d entries", n),
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(transitionsLogPath(ctx))
			if err != nil {
				return fmt.Errorf("read transitions.log: %w", err)
			}
			count := countNonEmptyLines(string(content))
			if count != n {
				return fmt.Errorf("expected %d lines in transitions.log, got %d\nContent:\n%s", n, count, string(content))
			}
			return nil
		},
	}
}

// TransitionsLogLineCountAtLeast asserts .kanban/transitions.log has at least n
// non-empty lines. Used in concurrency tests where the exact count from concurrent
// writes may vary (some "already in progress" calls are expected to be no-ops).
func TransitionsLogLineCountAtLeast(n int) Step {
	return Step{
		Description: fmt.Sprintf("transitions log has at least %d entries", n),
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(transitionsLogPath(ctx))
			if err != nil {
				return fmt.Errorf("read transitions.log: %w", err)
			}
			count := countNonEmptyLines(string(content))
			if count < n {
				return fmt.Errorf("expected at least %d lines in transitions.log, got %d\nContent:\n%s", n, count, string(content))
			}
			return nil
		},
	}
}

// TransitionsLogLastLineContains reads .kanban/transitions.log and asserts the
// last non-empty line contains all of the provided fields.
func TransitionsLogLastLineContains(fields ...string) Step {
	return Step{
		Description: fmt.Sprintf("transitions log last line contains %v", fields),
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(transitionsLogPath(ctx))
			if err != nil {
				return fmt.Errorf("read transitions.log: %w", err)
			}
			last := lastNonEmptyLine(string(content))
			if last == "" {
				return fmt.Errorf("transitions.log is empty")
			}
			for _, field := range fields {
				if !strings.Contains(last, field) {
					return fmt.Errorf("expected transitions.log last line to contain %q\nLine: %s\nFull content:\n%s", field, last, string(content))
				}
			}
			return nil
		},
	}
}

// TransitionsLogHasNoTruncatedLines reads .kanban/transitions.log and asserts
// every non-empty line has exactly 5 space-separated fields (the required format).
func TransitionsLogHasNoTruncatedLines() Step {
	return Step{
		Description: "transitions log has no truncated lines",
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(transitionsLogPath(ctx))
			if err != nil {
				return fmt.Errorf("read transitions.log: %w", err)
			}
			scanner := bufio.NewScanner(strings.NewReader(string(content)))
			lineNum := 0
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				lineNum++
				fields := strings.Fields(line)
				if len(fields) != 5 {
					return fmt.Errorf("truncated or malformed line %d in transitions.log (expected 5 fields, got %d): %q", lineNum, len(fields), line)
				}
			}
			return nil
		},
	}
}

// TaskFileDoesNotContainStatusField reads the task file for taskID and asserts
// it does not contain a "status:" YAML field. This validates the DESIGN wave
// decision that status is derived solely from transitions.log.
func TaskFileDoesNotContainStatusField(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("task file %s has no status field", taskID),
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(taskFilePath(ctx, taskID))
			if err != nil {
				return fmt.Errorf("read task file %s: %w", taskID, err)
			}
			if strings.Contains(string(content), "status:") {
				return fmt.Errorf("expected task file %s to have no status: field\nContent:\n%s", taskID, string(content))
			}
			return nil
		},
	}
}

// TaskFileContainsComment reads the task file for taskID and asserts it contains
// the given comment text. Used to verify the onboarding comment written on creation.
func TaskFileContainsComment(taskID, text string) Step {
	return Step{
		Description: fmt.Sprintf("task file %s contains comment %q", taskID, text),
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(taskFilePath(ctx, taskID))
			if err != nil {
				return fmt.Errorf("read task file %s: %w", taskID, err)
			}
			if !strings.Contains(string(content), text) {
				return fmt.Errorf("expected task file %s to contain %q\nContent:\n%s", taskID, text, string(content))
			}
			return nil
		},
	}
}

// TaskFileMtimeUnchangedAfter records the mtime of the task file for taskID,
// executes action, then asserts the mtime has not changed. This validates that
// a command did not touch the task file on disk.
func TaskFileMtimeUnchangedAfter(taskID string, action func(*Context) error) Step {
	return Step{
		Description: fmt.Sprintf("task file %s mtime unchanged after action", taskID),
		Run: func(ctx *Context) error {
			path := taskFilePath(ctx, taskID)
			before, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("stat task file %s before action: %w", taskID, err)
			}
			if err := action(ctx); err != nil {
				return fmt.Errorf("action during mtime check: %w", err)
			}
			after, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("stat task file %s after action: %w", taskID, err)
			}
			if !after.ModTime().Equal(before.ModTime()) {
				return fmt.Errorf("task file %s was modified during action (before=%v after=%v)", taskID, before.ModTime(), after.ModTime())
			}
			return nil
		},
	}
}

// BoardShowsTaskInColumn runs "kanban board" and asserts that taskID appears
// under the named column heading. heading should be one of: "To Do", "In Progress", "Done".
func BoardShowsTaskInColumn(taskID, heading string) Step {
	return Step{
		Description: fmt.Sprintf("board shows %s under %s", taskID, heading),
		Run:         BoardShowsTaskUnder(taskID, heading).Run,
	}
}

// BoardOutputContains asserts the current ctx.lastOutput (set by a prior
// DeveloperRunsKanbanBoard or DeveloperRunsKanbanBoardWithMe call) contains text.
func BoardOutputContains(text string) Step {
	return Step{
		Description: fmt.Sprintf("board output contains %q", text),
		Run:         OutputContains(text).Run,
	}
}

// BoardOutputDoesNotContain asserts ctx.lastOutput does not contain text.
func BoardOutputDoesNotContain(text string) Step {
	return Step{
		Description: fmt.Sprintf("board output does not contain %q", text),
		Run:         OutputDoesNotContain(text).Run,
	}
}

// GitHeadSHA returns the current HEAD commit SHA. Used to capture a "since"
// point for ci-done tests. Stores the SHA in ctx as a side-effect by setting
// ctx.lastOutput — callers should capture via a closure or helper.
func CurrentGitHeadSHA(ctx *Context) (string, error) {
	out, err := gitCmd(ctx, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// CiDoneCommitContainsOnlyTransitionsLog asserts the most recent commit touched
// only .kanban/transitions.log and no task files.
func CiDoneCommitContainsOnlyTransitionsLog() Step {
	return Step{
		Description: "ci-done commit touches only transitions.log, not task files",
		Run: func(ctx *Context) error {
			out, err := gitCmd(ctx, "show", "--name-only", "--pretty=format:", "HEAD")
			if err != nil {
				return fmt.Errorf("git show HEAD: %w", err)
			}
			files := strings.Fields(out)
			for _, f := range files {
				if !strings.Contains(f, "transitions.log") {
					return fmt.Errorf("ci-done commit includes unexpected file %q (expected only transitions.log)\nFiles in commit:\n%s", f, out)
				}
			}
			if len(files) == 0 {
				return fmt.Errorf("ci-done commit appears to be empty (no files changed)")
			}
			return nil
		},
	}
}

// RebaseSquashingAllCommits performs a non-interactive rebase that squashes all
// commits onto the initial commit. Used in rebase safety tests for US-BSG-02.
func RebaseSquashingAllCommits() Step {
	return Step{
		Description: "rebase squashing all non-initial commits",
		Run: func(ctx *Context) error {
			// Count commits to rebase
			out, err := gitCmd(ctx, "rev-list", "--count", "HEAD")
			if err != nil {
				return fmt.Errorf("count commits: %w", err)
			}
			count := strings.TrimSpace(out)
			_ = count

			// Use GIT_SEQUENCE_EDITOR to squash all commits after first.
			// GIT_EDITOR=true prevents git from opening an interactive editor
			// for the squash commit message (the `true` command exits 0 immediately).
			env := append(ctx.env,
				`GIT_SEQUENCE_EDITOR=sed -i.bak '2,$s/^pick/squash/'`,
				`GIT_EDITOR=true`,
			)
			cmdCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, "git", "rebase", "-i", "--root")
			cmd.Dir = ctx.repoDir
			cmd.Env = env

			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = &buf

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("git rebase -i --root: %w\nOutput: %s", err, buf.String())
			}
			return nil
		},
	}
}

// ---- helpers ----

// transitionsLogPath returns the absolute path to .kanban/transitions.log in ctx.repoDir.
func transitionsLogPath(ctx *Context) string {
	return filepath.Join(ctx.repoDir, ".kanban", "transitions.log")
}

// countNonEmptyLines counts lines in s that are non-empty after trimming.
func countNonEmptyLines(s string) int {
	count := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// lastNonEmptyLine returns the last non-empty line of s.
func lastNonEmptyLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return strings.TrimSpace(lines[i])
		}
	}
	return ""
}
