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

	"github.com/jmsargent/kanban/pkg/simpledsl"
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
func InAGitRepo(params ...string) Step {
	return Step{
		Description: "a git repository in a temp dir",
		Run: func(ctx *Context) error {
			dir := ctx.t.TempDir()
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
func KanbanInitialised(params ...string) Step {
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
func NoKanbanSetup(params ...string) Step {
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
			ctx.repoDir = ctx.t.TempDir()
			return nil
		},
	}
}

var aTaskWithStatusDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
	simpledsl.NewOptionalArg("status").SetDefault("todo"),
)

// ATaskWithStatus creates a task via "kanban new" and, if status is not "todo",
// injects the status directly into the task YAML front matter.
// Required param: "title: <title>". Optional param: "status: <status>" (default "todo").
func ATaskWithStatus(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task with status (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := aTaskWithStatusDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("ATaskWithStatus: %w", err)
			}
			title := vals.Value("title")
			status := vals.Value("status")
			run(ctx, "new", title)
			if ctx.lastExit != 0 {
				return fmt.Errorf("kanban new failed: %s", ctx.lastOutput)
			}
			taskID := ctx.lastTaskID
			if taskID == "" {
				return fmt.Errorf("could not determine task ID from output: %s", ctx.lastOutput)
			}
			if status != "todo" {
				if err := injectStatusIntoYAML(ctx, taskID, status); err != nil {
					return fmt.Errorf("inject status into task YAML: %w", err)
				}
			}
			return nil
		},
	}
}

// injectStatusIntoYAML inserts a "status: <value>" line as the first YAML key
// in the task file's front matter, immediately after any comment lines that
// follow the opening "---" delimiter. This allows legacy use cases that read
// status from YAML (via FindByID) to see the correct precondition.
//
// If the file already contains a "status:" line it is replaced in-place to
// avoid duplicate keys that would cause YAML parse errors.
func injectStatusIntoYAML(ctx *Context, taskID, status string) error {
	taskPath := taskFilePath(ctx, taskID)
	data, err := os.ReadFile(taskPath)
	if err != nil {
		return fmt.Errorf("read task file: %w", err)
	}
	content := string(data)

	// If there is already a status: line, replace it rather than add another.
	if strings.Contains(content, "\nstatus:") {
		updated := replaceStatusLine(content, status)
		return os.WriteFile(taskPath, []byte(updated), 0o644)
	}

	// No existing status field: insert after the opening "---\n" and any
	// comment lines, just before the first real YAML key.
	lines := strings.Split(content, "\n")
	var out []string
	inserted := false
	inFrontMatter := false
	for i, line := range lines {
		if !inFrontMatter && line == "---" {
			inFrontMatter = true
			out = append(out, line)
			continue
		}
		if inFrontMatter && !inserted {
			trimmed := strings.TrimSpace(line)
			// Skip blank lines and comment lines before inserting.
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				out = append(out, line)
				continue
			}
			// Insert status before this line (the first real YAML key).
			out = append(out, "status: "+status)
			inserted = true
		}
		_ = i
		out = append(out, line)
	}
	return os.WriteFile(taskPath, []byte(strings.Join(out, "\n")), 0o644)
}

// replaceStatusLine replaces the first "status: <anything>" line in content
// with "status: <newStatus>".
func replaceStatusLine(content, newStatus string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "status:") {
			lines[i] = "status: " + newStatus
			break
		}
	}
	return strings.Join(lines, "\n")
}

var aTaskWithStatusAsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
	simpledsl.NewRequiredArg("status"),
	simpledsl.NewRequiredArg("id"),
)

// ATaskWithStatusAs creates a task then renames it to the given id.
// Required params: "title: <title>", "status: <status>", "id: <TASK-NNN>".
func ATaskWithStatusAs(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task with status as (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := aTaskWithStatusAsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("ATaskWithStatusAs: %w", err)
			}
			title := vals.Value("title")
			status := vals.Value("status")
			taskID := vals.Value("id")
			if err := ATaskWithStatus("title: "+title, "status: "+status).Run(ctx); err != nil {
				return err
			}
			if taskID != "" && ctx.lastTaskID != "" && ctx.lastTaskID != taskID {
				oldID := ctx.lastTaskID
				oldPath := taskFilePath(ctx, oldID)
				newPath := taskFilePath(ctx, taskID)
				content, err := os.ReadFile(oldPath)
				if err != nil {
					return fmt.Errorf("read task file: %w", err)
				}
				updated := strings.ReplaceAll(string(content), "id: "+oldID, "id: "+taskID)
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

var aTaskExistsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
)

// ATaskExists creates a task with "todo" status. Required param: "title: <title>".
func ATaskExists(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task exists (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := aTaskExistsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("ATaskExists: %w", err)
			}
			return ATaskWithStatus("title: "+vals.Value("title"), "status: todo").Run(ctx)
		},
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

// WithLastOutput sets ctx.lastOutput to the given string, allowing assertion
// step factories to be tested in isolation without invoking a subprocess.
func WithLastOutput(output string) Step {
	return Step{
		Description: fmt.Sprintf("with last output %q", output),
		Run: func(ctx *Context) error {
			ctx.lastOutput = output
			return nil
		},
	}
}
