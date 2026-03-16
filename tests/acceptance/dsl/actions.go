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

// IRunKanban runs "kanban <subcommand>" splitting the subcommand on whitespace.
func IRunKanban(subcommand string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban %s", subcommand),
		Run: func(ctx *Context) error {
			args := strings.Fields(subcommand)
			run(ctx, args...)
			return nil
		},
	}
}

// IRunKanbanNew runs "kanban new <title>".
func IRunKanbanNew(title string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban new %q", title),
		Run: func(ctx *Context) error {
			run(ctx, "new", title)
			return nil
		},
	}
}

// IRunKanbanNewWithOptions runs "kanban new <title> --priority P --due D --assignee A".
func IRunKanbanNewWithOptions(title, priority, due, assignee string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban new %q with priority=%s due=%s assignee=%s", title, priority, due, assignee),
		Run: func(ctx *Context) error {
			run(ctx, "new", title, "--priority", priority, "--due", due, "--assignee", assignee)
			return nil
		},
	}
}

// IRunKanbanBoard runs "kanban board".
func IRunKanbanBoard() Step {
	return Step{
		Description: "I run kanban board",
		Run: func(ctx *Context) error {
			run(ctx, "board")
			return nil
		},
	}
}

// IRunKanbanBoardJSON runs "kanban board --json".
func IRunKanbanBoardJSON() Step {
	return Step{
		Description: "I run kanban board --json",
		Run: func(ctx *Context) error {
			run(ctx, "board", "--json")
			return nil
		},
	}
}

// IRunKanbanStart runs "kanban start <taskID>".
func IRunKanbanStart(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban start %s", taskID),
		Run: func(ctx *Context) error {
			run(ctx, "start", taskID)
			return nil
		},
	}
}

// IRunKanbanStartOnThatTask runs "kanban start <ctx.lastTaskID>" resolved at run time.
func IRunKanbanStartOnThatTask() Step {
	return Step{
		Description: "I run kanban start on that task",
		Run: func(ctx *Context) error {
			if ctx.lastTaskID == "" {
				return fmt.Errorf("no task ID in context")
			}
			run(ctx, "start", ctx.lastTaskID)
			return nil
		},
	}
}

// IRunKanbanEditAddDescription runs "kanban edit <taskID>" with a mock EDITOR script
// that appends the given description text after the YAML front matter.
func IRunKanbanEditAddDescription(taskID, description string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban edit %s and add description", taskID),
		Run: func(ctx *Context) error {
			scriptDir, err := os.MkdirTemp("", "kanban-editor-desc-*")
			if err != nil {
				return fmt.Errorf("create editor script dir: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(scriptDir) })

			scriptPath := filepath.Join(scriptDir, "editor.sh")
			// Append description after the closing --- of the front matter.
			script := fmt.Sprintf(
				"#!/bin/sh\nprintf '\\n%s\\n' >> \"$1\"\n",
				strings.ReplaceAll(description, "'", "'\\''"),
			)
			if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
				return fmt.Errorf("write editor script: %w", err)
			}

			return runWithEditor(ctx, scriptPath, taskID)
		},
	}
}

// IRunKanbanEditTitle runs "kanban edit <taskID>" with a mock EDITOR script
// that replaces the title field in the front matter with newTitle.
func IRunKanbanEditTitle(taskID, newTitle string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban edit %s and set title to %q", taskID, newTitle),
		Run: func(ctx *Context) error {
			scriptDir, err := os.MkdirTemp("", "kanban-editor-title-*")
			if err != nil {
				return fmt.Errorf("create editor script dir: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(scriptDir) })

			scriptPath := filepath.Join(scriptDir, "editor.sh")
			script := fmt.Sprintf("#!/bin/sh\nsed -i.bak 's/^title: .*/title: %s/' \"$1\"\n", newTitle)
			if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
				return fmt.Errorf("write editor script: %w", err)
			}

			return runWithEditor(ctx, scriptPath, taskID)
		},
	}
}

// runWithEditor executes "kanban edit <taskID>" with EDITOR set to scriptPath
// and captures output into ctx fields.
func runWithEditor(ctx *Context, scriptPath, taskID string) error {
	cmdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := append(ctx.env, "EDITOR="+scriptPath)
	cmd := exec.CommandContext(cmdCtx, ctx.binPath, "edit", taskID)
	cmd.Dir = ctx.repoDir
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	captureResult(ctx, cmd.Run(), &stdout, &stderr)
	if taskID != "" {
		ctx.lastTaskID = taskID
	}
	return nil
}

// IRunKanbanDelete runs "kanban delete <taskID>" piping confirmInput+"\n" to stdin.
func IRunKanbanDelete(taskID, confirmInput string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban delete %s confirming with %q", taskID, confirmInput),
		Run: func(ctx *Context) error {
			cmdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, ctx.binPath, "delete", taskID)
			cmd.Dir = ctx.repoDir
			cmd.Env = ctx.env
			cmd.Stdin = strings.NewReader(confirmInput + "\n")

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			captureResult(ctx, cmd.Run(), &stdout, &stderr)
			return nil
		},
	}
}

// IRunKanbanDeleteForce runs "kanban delete <taskID> --force".
func IRunKanbanDeleteForce(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban delete %s --force", taskID),
		Run: func(ctx *Context) error {
			run(ctx, "delete", taskID, "--force")
			return nil
		},
	}
}

// ICommitWithMessage runs "git commit --allow-empty -m <message>" in ctx.repoDir.
// The exit code and output are captured into ctx fields.
func ICommitWithMessage(message string) Step {
	return Step{
		Description: fmt.Sprintf("I commit with message %q", message),
		Run: func(ctx *Context) error {
			cmdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, "git", "commit", "--allow-empty", "-m", message)
			cmd.Dir = ctx.repoDir
			cmd.Env = ctx.env

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			captureResult(ctx, cmd.Run(), &stdout, &stderr)
			return nil
		},
	}
}

// ICommitWithTaskID runs "git commit --allow-empty -m <ctx.lastTaskID>: start working on this".
// The task ID is resolved at run time from ctx.lastTaskID.
func ICommitWithTaskID() Step {
	return Step{
		Description: "I commit with message containing the task ID",
		Run: func(ctx *Context) error {
			if ctx.lastTaskID == "" {
				return fmt.Errorf("no task ID in context")
			}
			return ICommitWithMessage(ctx.lastTaskID + ": start working on this").Run(ctx)
		},
	}
}

// CIStepRunsPass runs "kanban ci-done".
func CIStepRunsPass() Step {
	return Step{
		Description: "the CI step runs after all tests pass",
		Run: func(ctx *Context) error {
			run(ctx, "ci-done")
			return nil
		},
	}
}

// CIStepRunsFail runs "kanban ci-done" with KANBAN_TEST_EXIT=1 appended to env.
func CIStepRunsFail() Step {
	return Step{
		Description: "the CI step runs after one or more tests fail",
		Run: func(ctx *Context) error {
			ctx.env = append(ctx.env, "KANBAN_TEST_EXIT=1")
			run(ctx, "ci-done")
			return nil
		},
	}
}
