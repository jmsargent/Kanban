package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ---- Action steps for explicit-state-transitions feature ----

// IRunKanbanDone runs "kanban done <taskID>" and captures output/exit.
func IRunKanbanDone(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban done %s", taskID),
		Run: func(ctx *Context) error {
			run(ctx, "done", taskID)
			return nil
		},
	}
}

// IRunKanbanHookCommitMsg invokes "kanban _hook commit-msg <msgFile>" to
// simulate a leftover git hook delegation. msgFile should be the path to a
// file containing a commit message. Creates a temp file containing content
// and passes its path as the argument.
func IRunKanbanHookCommitMsg(content string) Step {
	return Step{
		Description: "I run kanban _hook commit-msg with a commit message",
		Run: func(ctx *Context) error {
			f, err := os.CreateTemp("", "commit-msg-*")
			if err != nil {
				return fmt.Errorf("create temp commit-msg file: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.Remove(f.Name()) })
			if _, err := f.WriteString(content); err != nil {
				_ = f.Close()
				return fmt.Errorf("write commit-msg file: %w", err)
			}
			_ = f.Close()
			run(ctx, "_hook", "commit-msg", f.Name())
			return nil
		},
	}
}

// ---- Assertion steps for explicit-state-transitions feature ----

// CaptureGitHeadSHA captures the current HEAD commit SHA into the given
// pointer. Use before a command to record the "before" SHA.
func CaptureGitHeadSHA(sha *string) Step {
	return Step{
		Description: "capture current git HEAD SHA",
		Run: func(ctx *Context) error {
			out, err := gitCmd(ctx, "rev-parse", "HEAD")
			if err != nil {
				return fmt.Errorf("git rev-parse HEAD: %w", err)
			}
			*sha = strings.TrimSpace(out)
			return nil
		},
	}
}

// GitHeadSHAIs asserts the current HEAD SHA equals the value pointed to by sha.
// Use after a command to verify no new commit was created.
func GitHeadSHAIs(sha *string) Step {
	return Step{
		Description: "git HEAD SHA is unchanged",
		Run: func(ctx *Context) error {
			out, err := gitCmd(ctx, "rev-parse", "HEAD")
			if err != nil {
				return fmt.Errorf("git rev-parse HEAD: %w", err)
			}
			current := strings.TrimSpace(out)
			if current != *sha {
				return fmt.Errorf("expected HEAD SHA to be unchanged\nBefore: %s\nAfter:  %s\nA new commit was created", *sha, current)
			}
			return nil
		},
	}
}

// TransitionsLogAbsent asserts that .kanban/transitions.log does NOT exist.
func TransitionsLogAbsent() Step {
	return Step{
		Description: "transitions.log does not exist",
		Run: func(ctx *Context) error {
			logPath := filepath.Join(ctx.repoDir, ".kanban", "transitions.log")
			if _, err := os.Stat(logPath); err == nil {
				return fmt.Errorf("expected .kanban/transitions.log to be absent but it exists")
			}
			return nil
		},
	}
}

// TaskFileStatusIs reads the task YAML front matter for taskID and asserts its
// status: field equals expected. Unlike TaskHasStatus, this reads ONLY from the
// YAML file — it does not fall back to transitions.log. Use in tests that
// verify the new YAML-only state source.
func TaskFileStatusIs(taskID, expected string) Step {
	return Step{
		Description: fmt.Sprintf("task file %s has YAML status %q", taskID, expected),
		Run: func(ctx *Context) error {
			actual := statusFromYAML(ctx, taskID)
			if actual == "" {
				actual = "todo" // default when field absent
			}
			if actual != expected {
				content, _ := os.ReadFile(taskFilePath(ctx, taskID))
				return fmt.Errorf("expected task %s YAML status to be %q, got %q\nFile:\n%s",
					taskID, expected, actual, string(content))
			}
			return nil
		},
	}
}

// InitDidNotAutoCommit asserts that the most recent git commit is NOT a
// kanban init commit. Used to verify kanban init no longer auto-commits.
func InitDidNotAutoCommit() Step {
	return Step{
		Description: "kanban init did not create an auto-commit",
		Run: func(ctx *Context) error {
			out, err := gitCmd(ctx, "log", "--oneline", "-1")
			if err != nil {
				return fmt.Errorf("git log: %w", err)
			}
			lower := strings.ToLower(out)
			if strings.Contains(lower, "kanban: initialise") || strings.Contains(lower, "kanban: initialize") {
				return fmt.Errorf("kanban init created an auto-commit but should not have:\n%s", out)
			}
			return nil
		},
	}
}

// IRunKanbanCiDoneFrom runs "kanban ci-done --since <*sha>" where *sha is
// evaluated at step execution time. Use with CaptureGitHeadSHA to pass the
// base-ref captured before the commits under test.
func IRunKanbanCiDoneFrom(sha *string) Step {
	return Step{
		Description: "I run kanban ci-done from captured SHA",
		Run: func(ctx *Context) error {
			run(ctx, "ci-done", "--since", *sha)
			return nil
		},
	}
}

// KanbanDotKanbanDirectoryExists asserts .kanban/ was created by kanban init.
func KanbanDotKanbanDirectoryExists() Step {
	return Step{
		Description: ".kanban/ directory exists after init",
		Run: func(ctx *Context) error {
			dir := filepath.Join(ctx.repoDir, ".kanban")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("expected .kanban/ to exist after kanban init")
			}
			return nil
		},
	}
}
