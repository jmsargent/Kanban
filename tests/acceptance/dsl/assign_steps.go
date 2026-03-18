package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// --- Setup Steps for Auto-Assign on Start ---

// GitIdentityConfiguredAs reconfigures the git user.name for the current repo
// to name. Must be called after InAGitRepo() so that ctx.repoDir is set.
// Use this to override the default "Test User" identity for tests that need a
// specific developer name.
func GitIdentityConfiguredAs(name string) Step {
	return Step{
		Description: fmt.Sprintf("git identity configured as %q", name),
		Run: func(ctx *Context) error {
			if _, err := gitCmd(ctx, "config", "user.name", name); err != nil {
				return fmt.Errorf("git config user.name %q: %w", name, err)
			}
			return nil
		},
	}
}

// GitIdentityUnset unsets git user.name in the local repo config and isolates
// the global and system git config so that no fallback identity is available.
// Call after InAGitRepo() + KanbanInitialised() + any task setup steps — the
// task setup steps require identity to run kanban new.
// After this step, git config user.name returns empty, triggering
// ErrGitIdentityNotConfigured in kanban commands that require identity.
func GitIdentityUnset() Step {
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

// ATaskAssigneeSetTo patches the assignee field in the task file for taskID.
// The task file must already exist (created via ATaskWithStatusAs or similar).
// If the file already contains an "assignee:" line it is replaced; if not,
// the field is inserted after the "status:" line.
func ATaskAssigneeSetTo(taskID, assignee string) Step {
	return Step{
		Description: fmt.Sprintf("task %s assignee set to %q", taskID, assignee),
		Run: func(ctx *Context) error {
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

// TaskHasAssignee reads the task file for taskID and asserts it contains
// `assignee: <expected>` in the YAML front matter.
func TaskHasAssignee(taskID, expected string) Step {
	return Step{
		Description: fmt.Sprintf("task %s has assignee %q", taskID, expected),
		Run: func(ctx *Context) error {
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

// TaskAssigneeRemains is a semantic alias for TaskHasAssignee used in assertions
// that verify the assignee was not changed by an operation.
func TaskAssigneeRemains(taskID, expected string) Step {
	return Step{
		Description: fmt.Sprintf("task %s assignee remains %q", taskID, expected),
		Run:         TaskHasAssignee(taskID, expected).Run,
	}
}

// TaskHasNoAssignee reads the task file for taskID and asserts the assignee
// field is absent or empty.
func TaskHasNoAssignee(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("task %s has no assignee", taskID),
		Run: func(ctx *Context) error {
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

// StdoutDoesNotContain asserts ctx.lastStdout does NOT contain text.
func StdoutDoesNotContain(text string) Step {
	return Step{
		Description: fmt.Sprintf("stdout does not contain %q", text),
		Run: func(ctx *Context) error {
			if strings.Contains(ctx.lastStdout, text) {
				return fmt.Errorf("expected stdout NOT to contain %q\nStdout:\n%s", text, ctx.lastStdout)
			}
			return nil
		},
	}
}
