package dsl

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// gitCmd runs a git command in ctx.repoDir using ctx.env and returns combined output.
func gitCmd(ctx *Context, args ...string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Dir = ctx.repoDir
	cmd.Env = ctx.env

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	return strings.TrimSpace(buf.String()), err
}

// InAGitRepo creates a temp directory, initialises a git repo with a baseline
// commit, and registers cleanup via t.Cleanup.
func InAGitRepo() Step {
	return Step{
		Description: "a git repository in a temp dir",
		Run: func(ctx *Context) error {
			dir, err := os.MkdirTemp("", "kanban-test-*")
			if err != nil {
				return fmt.Errorf("create temp dir: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(dir) })
			ctx.repoDir = dir

			if _, err := gitCmd(ctx, "init"); err != nil {
				return fmt.Errorf("git init: %w", err)
			}
			if _, err := gitCmd(ctx, "config", "user.email", "test@example.com"); err != nil {
				return fmt.Errorf("git config user.email: %w", err)
			}
			if _, err := gitCmd(ctx, "config", "user.name", "Test User"); err != nil {
				return fmt.Errorf("git config user.name: %w", err)
			}

			readmePath := filepath.Join(dir, "README.md")
			if err := os.WriteFile(readmePath, []byte("# test\n"), 0644); err != nil {
				return fmt.Errorf("write README.md: %w", err)
			}
			if _, err := gitCmd(ctx, "add", "."); err != nil {
				return fmt.Errorf("git add: %w", err)
			}
			if _, err := gitCmd(ctx, "commit", "-m", "Initial commit"); err != nil {
				return fmt.Errorf("git commit: %w", err)
			}
			return nil
		},
	}
}

// KanbanInitialised runs "kanban init" in ctx.repoDir and returns an error if
// the command exits non-zero.
func KanbanInitialised() Step {
	return Step{
		Description: "kanban already initialised",
		Run: func(ctx *Context) error {
			run(ctx, "init")
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban init failed (exit %d): %s", ctx.lastExit, ctx.lastOutput)
			}
			return nil
		},
	}
}

// NoKanbanSetup asserts that .kanban/ does not exist in ctx.repoDir.
func NoKanbanSetup() Step {
	return Step{
		Description: "repository has no kanban setup",
		Run: func(ctx *Context) error {
			if _, err := os.Stat(filepath.Join(ctx.repoDir, ".kanban")); !os.IsNotExist(err) {
				return fmt.Errorf("expected no .kanban/ directory but found one")
			}
			return nil
		},
	}
}

// NotAGitRepo creates a temp directory without running git init and registers
// cleanup via t.Cleanup.
func NotAGitRepo() Step {
	return Step{
		Description: "current directory is not a git repository",
		Run: func(ctx *Context) error {
			dir, err := os.MkdirTemp("", "kanban-nogit-*")
			if err != nil {
				return fmt.Errorf("create temp dir: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(dir) })
			ctx.repoDir = dir
			return nil
		},
	}
}

// ATaskWithStatus creates a task via "kanban new" and, if status is not "todo",
// directly rewrites the status field in the task file.
func ATaskWithStatus(title, status string) Step {
	return Step{
		Description: fmt.Sprintf("task %q with status %q", title, status),
		Run: func(ctx *Context) error {
			run(ctx, "new", title)
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban new failed: %s", ctx.lastOutput)
			}
			taskID := ctx.lastTaskID
			if taskID == "" {
				return fmt.Errorf("could not determine task ID from output: %s", ctx.lastOutput)
			}
			if status != "todo" {
				taskPath := taskFilePath(ctx, taskID)
				content, err := os.ReadFile(taskPath)
				if err != nil {
					return fmt.Errorf("read task file: %w", err)
				}
				updated := strings.ReplaceAll(string(content), "status: todo", "status: "+status)
				if err := os.WriteFile(taskPath, []byte(updated), 0644); err != nil {
					return fmt.Errorf("write task file: %w", err)
				}
			}
			return nil
		},
	}
}

// ATaskWithStatusAs creates a task with ATaskWithStatus logic, then renames the
// file and updates the id field if the generated ID differs from taskID.
func ATaskWithStatusAs(title, status, taskID string) Step {
	return Step{
		Description: fmt.Sprintf("task %q with status %q as %q", title, status, taskID),
		Run: func(ctx *Context) error {
			if err := ATaskWithStatus(title, status).Run(ctx); err != nil {
				return err
			}
			if taskID != "" && ctx.lastTaskID != "" && ctx.lastTaskID != taskID {
				oldPath := taskFilePath(ctx, ctx.lastTaskID)
				newPath := taskFilePath(ctx, taskID)
				content, err := os.ReadFile(oldPath)
				if err != nil {
					return fmt.Errorf("read task file: %w", err)
				}
				updated := strings.ReplaceAll(string(content), "id: "+ctx.lastTaskID, "id: "+taskID)
				if err := os.WriteFile(newPath, []byte(updated), 0644); err != nil {
					return fmt.Errorf("write renamed task file: %w", err)
				}
				if err := os.Remove(oldPath); err != nil {
					return fmt.Errorf("remove old task file: %w", err)
				}
				ctx.lastTaskID = taskID
			}
			return nil
		},
	}
}

// ATaskExists creates a task with "todo" status (shorthand for ATaskWithStatus(title, "todo")).
func ATaskExists(title string) Step {
	return Step{
		Description: fmt.Sprintf("task %q exists", title),
		Run:         ATaskWithStatus(title, "todo").Run,
	}
}

// NoTasksExist removes all files from the .kanban/tasks/ directory.
func NoTasksExist() Step {
	return Step{
		Description: "no tasks in workspace",
		Run: func(ctx *Context) error {
			dir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil // directory may not exist yet; that is fine
			}
			for _, e := range entries {
				if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
					return fmt.Errorf("remove task file %s: %w", e.Name(), err)
				}
			}
			return nil
		},
	}
}

// TaskFileExistsAs asserts that .kanban/tasks/<taskID>.md exists.
func TaskFileExistsAs(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("task file %s.md exists", taskID),
		Run: func(ctx *Context) error {
			if _, err := os.Stat(taskFilePath(ctx, taskID)); os.IsNotExist(err) {
				return fmt.Errorf("expected task file %s.md to exist", taskID)
			}
			return nil
		},
	}
}

// CommitHookInstalled runs "kanban init" (idempotent) to ensure the commit-msg
// hook is present.
func CommitHookInstalled() Step {
	return Step{
		Description: "commit hook installed",
		Run: func(ctx *Context) error {
			run(ctx, "init")
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban init (to install hook) failed: %s", ctx.lastOutput)
			}
			return nil
		},
	}
}

// EnvVarSet appends name=value to ctx.env.
func EnvVarSet(name, value string) Step {
	return Step{
		Description: fmt.Sprintf("environment variable %s=%s", name, value),
		Run: func(ctx *Context) error {
			ctx.env = append(ctx.env, name+"="+value)
			return nil
		},
	}
}

// PipelineCommitWith sets ctx.lastTaskID to taskID, stages all changes, and
// creates a git commit referencing the task ID.
func PipelineCommitWith(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("pipeline commit referencing %s", taskID),
		Run: func(ctx *Context) error {
			ctx.lastTaskID = taskID
			if _, err := gitCmd(ctx, "add", "-A"); err != nil {
				return fmt.Errorf("git add -A: %w", err)
			}
			if _, err := gitCmd(ctx, "commit", "--allow-empty", "-m", taskID+": work in progress"); err != nil {
				return fmt.Errorf("create pipeline commit referencing %s: %w", taskID, err)
			}
			return nil
		},
	}
}
