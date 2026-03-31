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

var iRunKanbanDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("subcommand"),
)

// IRunKanban runs "kanban <subcommand>" splitting the subcommand on whitespace.
// Required param: "subcommand: <cmd>".
func IRunKanban(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanban: %w", err)
			}
			args := strings.Fields(vals.Value("subcommand"))
			run(ctx, args...)
			return nil
		},
	}
}

var iRunKanbanNewDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
)

// IRunKanbanNew runs "kanban new <title>".
// Required param: "title: <title>".
func IRunKanbanNew(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban new (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanNewDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanbanNew: %w", err)
			}
			run(ctx, "new", vals.Value("title"))
			return nil
		},
	}
}

var iRunKanbanNewWithOptionsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
	simpledsl.NewRequiredArg("priority"),
	simpledsl.NewRequiredArg("due"),
	simpledsl.NewRequiredArg("assignee"),
)

// IRunKanbanNewWithOptions runs "kanban new <title> --priority P --due D --assignee A".
// Required params: "title: <title>", "priority: <P>", "due: <D>", "assignee: <A>".
func IRunKanbanNewWithOptions(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban new with options (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanNewWithOptionsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanbanNewWithOptions: %w", err)
			}
			run(ctx, "new", vals.Value("title"),
				"--priority", vals.Value("priority"),
				"--due", vals.Value("due"),
				"--assignee", vals.Value("assignee"),
			)
			return nil
		},
	}
}

// IRunKanbanBoard runs "kanban board".
func IRunKanbanBoard(params ...string) Step {
	return Step{
		Description: "I run kanban board",
		Run: func(ctx *Context) error {
			run(ctx, "board")
			return nil
		},
	}
}

// IRunKanbanBoardJSON runs "kanban board --json".
func IRunKanbanBoardJSON(params ...string) Step {
	return Step{
		Description: "I run kanban board --json",
		Run: func(ctx *Context) error {
			run(ctx, "board", "--json")
			return nil
		},
	}
}

var iRunKanbanStartDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
)

// IRunKanbanStart runs "kanban start <task>".
// Required param: "task: <TASK-NNN>".
func IRunKanbanStart(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban start (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanStartDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanbanStart: %w", err)
			}
			run(ctx, "start", vals.Value("task"))
			return nil
		},
	}
}

// IRunKanbanStartOnThatTask runs "kanban start <ctx.lastTaskID>" resolved at run time.
func IRunKanbanStartOnThatTask(params ...string) Step {
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

var iRunKanbanEditAddDescriptionDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewRequiredArg("description"),
)

// IRunKanbanEditAddDescription runs "kanban edit <task>" with a mock EDITOR script
// that appends the given description text after the YAML front matter.
// Required params: "task: <TASK-NNN>", "description: <text>".
func IRunKanbanEditAddDescription(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban edit and add description (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanEditAddDescriptionDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanbanEditAddDescription: %w", err)
			}
			taskID := vals.Value("task")
			description := vals.Value("description")

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

var iRunKanbanEditTitleDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewRequiredArg("title"),
)

// IRunKanbanEditTitle runs "kanban edit <task>" with a mock EDITOR that sets the title.
// Required params: "task: <TASK-NNN>", "title: <new title>".
func IRunKanbanEditTitle(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban edit title (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanEditTitleDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanbanEditTitle: %w", err)
			}
			taskID := vals.Value("task")
			newTitle := vals.Value("title")
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

var iRunKanbanDeleteDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewRequiredArg("confirm"),
)

// IRunKanbanDelete runs "kanban delete <task>" piping the confirm input to stdin.
// Required params: "task: <TASK-NNN>", "confirm: <y/n>".
func IRunKanbanDelete(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban delete (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanDeleteDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanbanDelete: %w", err)
			}
			taskID := vals.Value("task")
			confirmInput := vals.Value("confirm")
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

var iRunKanbanDeleteForceDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
)

// IRunKanbanDeleteForce runs "kanban delete <task> --force".
// Required param: "task: <TASK-NNN>".
func IRunKanbanDeleteForce(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban delete force (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iRunKanbanDeleteForceDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IRunKanbanDeleteForce: %w", err)
			}
			run(ctx, "delete", vals.Value("task"), "--force")
			return nil
		},
	}
}

var iCommitWithMessageDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("message"),
)

// ICommitWithMessage runs "git commit --allow-empty -m <message>" in ctx.repoDir.
// Required param: "message: <text>".
func ICommitWithMessage(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I commit with message (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := iCommitWithMessageDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("ICommitWithMessage: %w", err)
			}
			message := vals.Value("message")

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
func ICommitWithTaskID(params ...string) Step {
	return Step{
		Description: "I commit with message containing the task ID",
		Run: func(ctx *Context) error {
			if ctx.lastTaskID == "" {
				return fmt.Errorf("no task ID in context")
			}
			return ICommitWithMessage("message: " + ctx.lastTaskID + ": start working on this").Run(ctx)
		},
	}
}

// CIStepRunsPass runs "kanban ci-done".
func CIStepRunsPass(params ...string) Step {
	return Step{
		Description: "the CI step runs after all tests pass",
		Run: func(ctx *Context) error {
			run(ctx, "ci-done")
			return nil
		},
	}
}

// CIStepRunsFail runs "kanban ci-done" with KANBAN_TEST_EXIT=1 appended to env.
func CIStepRunsFail(params ...string) Step {
	return Step{
		Description: "the CI step runs after one or more tests fail",
		Run: func(ctx *Context) error {
			ctx.env = append(ctx.env, "KANBAN_TEST_EXIT=1")
			run(ctx, "ci-done")
			return nil
		},
	}
}