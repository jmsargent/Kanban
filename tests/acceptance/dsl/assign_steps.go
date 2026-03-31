package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmsargent/kanban/pkg/simpledsl"
)

// --- Setup Steps for Auto-Assign on Start ---

var gitIdentityConfiguredAsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("name"),
)

// GitIdentityConfiguredAs reconfigures the git user.name for the current repo.
// Required param: "name: <name>".
func GitIdentityConfiguredAs(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("git identity configured as (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := gitIdentityConfiguredAsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("GitIdentityConfiguredAs: %w", err)
			}
			name := vals.Value("name")
			if _, err := gitCmd(ctx, "config", "user.name", name); err != nil {
				return fmt.Errorf("git config user.name %q: %w", name, err)
			}
			return nil
		},
	}
}

// GitIdentityUnset unsets git user.name in the local repo config and isolates
// the global and system git config so that no fallback identity is available.
func GitIdentityUnset(params ...string) Step {
	return Step{
		Description: "git identity not configured",
		Run: func(ctx *Context) error {
			// Unset the local repo user.name set by InAGitRepo().
			_, _ = gitCmd(ctx, "config", "--unset", "user.name")

			// Isolate global and system config so no fallback name is found.
			noHome, err := os.MkdirTemp("", "kanban-nohome-*")
			if err != nil {
				return fmt.Errorf("create isolated home dir: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(noHome) })

			ctx.env = filterEnv(ctx.env, "HOME", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM",
				"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL")
			ctx.env = append(ctx.env,
				"HOME="+noHome,
				"GIT_CONFIG_NOSYSTEM=1",
				// Supply author/committer identity via env vars so git operations
				// (if any) still work, but `git config user.name` returns empty.
				"GIT_AUTHOR_NAME=Test User",
				"GIT_AUTHOR_EMAIL=test@example.com",
				"GIT_COMMITTER_NAME=Test User",
				"GIT_COMMITTER_EMAIL=test@example.com",
			)
			return nil
		},
	}
}

var aTaskAssigneeSetToDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewRequiredArg("assignee"),
)

// ATaskAssigneeSetTo patches the assignee field in the task file.
// Required params: "task: <TASK-NNN>", "assignee: <name>".
func ATaskAssigneeSetTo(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task assignee set to (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := aTaskAssigneeSetToDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("ATaskAssigneeSetTo: %w", err)
			}
			taskID := vals.Value("task")
			assignee := vals.Value("assignee")
			path := filepath.Join(ctx.repoDir, ".kanban", "tasks", taskID+".md")
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read task file %s: %w", taskID, err)
			}
			updated := strings.ReplaceAll(string(content), "assignee: \"\"", "assignee: "+assignee)
			if !strings.Contains(updated, "assignee: "+assignee) {
				// Fallback: insert after status line.
				updated = strings.ReplaceAll(updated,
					"\nassignee:", "\nassignee: "+assignee+" # replaced")
			}
			if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
				return fmt.Errorf("write task file %s: %w", taskID, err)
			}
			return nil
		},
	}
}

// --- Assertion Steps for Auto-Assign on Start ---

var taskHasAssigneeDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewRequiredArg("assignee"),
)

// TaskHasAssignee asserts the task file contains the given assignee.
// Required params: "task: <TASK-NNN>", "assignee: <name>".
func TaskHasAssignee(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task has assignee (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := taskHasAssigneeDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("TaskHasAssignee: %w", err)
			}
			taskID := vals.Value("task")
			expected := vals.Value("assignee")
			content, err := os.ReadFile(taskFilePath(ctx, taskID))
			if err != nil {
				return fmt.Errorf("task file %s not found: %w", taskID, err)
			}
			needle := "assignee: " + expected
			if !strings.Contains(string(content), needle) {
				return fmt.Errorf("expected task %s to contain %q\nFile content:\n%s",
					taskID, needle, string(content))
			}
			return nil
		},
	}
}

// TaskAssigneeRemains is a semantic alias for TaskHasAssignee.
// Required params: "task: <TASK-NNN>", "assignee: <name>".
func TaskAssigneeRemains(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task assignee remains (%s)", strings.Join(params, ", ")),
		Run:         TaskHasAssignee(params...).Run,
	}
}

var taskHasNoAssigneeDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
)

// TaskHasNoAssignee asserts the assignee field is absent or empty.
// Required param: "task: <TASK-NNN>".
func TaskHasNoAssignee(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task has no assignee (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := taskHasNoAssigneeDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("TaskHasNoAssignee: %w", err)
			}
			taskID := vals.Value("task")
			content, err := os.ReadFile(taskFilePath(ctx, taskID))
			if err != nil {
				return fmt.Errorf("task file %s not found: %w", taskID, err)
			}
			s := string(content)
			// Accept "assignee: " (empty value) or absence of the field.
			if strings.Contains(s, "assignee: ") {
				// Field present — check the value is empty.
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(strings.TrimSpace(line), "assignee:") {
						val := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "assignee:"))
						val = strings.Trim(val, `"`)
						if val != "" {
							return fmt.Errorf("expected task %s to have no assignee but found %q\nFile content:\n%s",
								taskID, val, s)
						}
					}
				}
			}
			return nil
		},
	}
}

var stdoutDoesNotContainDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("text"),
)

// StdoutDoesNotContain asserts ctx.lastStdout does NOT contain text.
// Required param: "text: <text>".
func StdoutDoesNotContain(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("stdout does not contain (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := stdoutDoesNotContainDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("StdoutDoesNotContain: %w", err)
			}
			text := vals.Value("text")
			if strings.Contains(ctx.lastStdout, text) {
				return fmt.Errorf("expected stdout NOT to contain %q\nStdout:\n%s", text, ctx.lastStdout)
			}
			return nil
		},
	}
}