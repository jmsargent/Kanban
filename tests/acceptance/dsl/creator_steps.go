package dsl

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jmsargent/kanban/pkg/simpledsl"
)

// --- Setup Steps for Creator Attribution ---

// InAGitRepoWithoutGitIdentity creates a temp git repo with no git user.name configured.
func InAGitRepoWithoutGitIdentity(params ...string) Step {
	return Step{
		Description: "a git repository with no git identity configured",
		Run: func(ctx *Context) error {
			dir, err := os.MkdirTemp("", "kanban-test-*")
			if err != nil {
				return fmt.Errorf("create temp dir: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(dir) })
			ctx.repoDir = dir

			// Isolated HOME: no .gitconfig → git config user.name returns empty.
			noHome, err := os.MkdirTemp("", "kanban-nohome-*")
			if err != nil {
				return fmt.Errorf("create isolated home dir: %w", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(noHome) })

			// Build env: override HOME and skip system config; supply author identity
			// via env vars only (not via git config) so git commit works but
			// `git config user.name` still returns empty.
			env := filterEnv(ctx.env, "HOME", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM",
				"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL")
			env = append(env,
				"HOME="+noHome,
				"GIT_CONFIG_NOSYSTEM=1",
				"GIT_AUTHOR_NAME=Test User",
				"GIT_AUTHOR_EMAIL=test@example.com",
				"GIT_COMMITTER_NAME=Test User",
				"GIT_COMMITTER_EMAIL=test@example.com",
			)
			ctx.env = env

			if _, err := gitCmd(ctx, "init"); err != nil {
				return fmt.Errorf("git init: %w", err)
			}
			// Set local git config so kanban init's git commit can succeed.
			// GitIdentityUnconfigured() removes these before the kanban binary
			// is invoked by the test, so GetIdentity() still detects no identity.
			if _, err := gitCmd(ctx, "config", "user.name", "Test User"); err != nil {
				return fmt.Errorf("git config user.name: %w", err)
			}
			if _, err := gitCmd(ctx, "config", "user.email", "test@example.com"); err != nil {
				return fmt.Errorf("git config user.email: %w", err)
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

// GitIdentityUnconfigured removes user.name and user.email from the repo's local git config.
func GitIdentityUnconfigured(params ...string) Step {
	return Step{
		Description: "git identity removed from local repo config",
		Run: func(ctx *Context) error {
			for _, key := range []string{"user.name", "user.email"} {
				if _, err := gitCmd(ctx, "config", "--unset", key); err != nil {
					var exitErr *exec.ExitError
					if errors.As(err, &exitErr) && exitErr.ExitCode() == 5 {
						continue
					}
					return fmt.Errorf("git config --unset %s: %w", key, err)
				}
			}
			return nil
		},
	}
}

var aPreExistingTaskWithoutCreatorDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
)

// APreExistingTaskWithoutCreator writes a minimal task file without a created_by field.
// Required param: "title: <title>".
func APreExistingTaskWithoutCreator(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("pre-existing task without creator field (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := aPreExistingTaskWithoutCreatorDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("APreExistingTaskWithoutCreator: %w", err)
			}
			title := vals.Value("title")
			tasksDir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			if err := os.MkdirAll(tasksDir, 0o755); err != nil {
				return fmt.Errorf("ensure tasks dir: %w", err)
			}
			entries, _ := os.ReadDir(tasksDir)
			maxNum := 0
			for _, e := range entries {
				var n int
				if _, err := fmt.Sscanf(e.Name(), "TASK-%d.md", &n); err == nil && n > maxNum {
					maxNum = n
				}
			}
			taskID := fmt.Sprintf("TASK-%03d", maxNum+1)
			// Legacy format: no created_by field.
			content := fmt.Sprintf("---\nid: %s\ntitle: %s\nstatus: todo\npriority: \"\"\ndue: \"\"\nassignee: \"\"\n---\n", taskID, title)
			if err := os.WriteFile(filepath.Join(tasksDir, taskID+".md"), []byte(content), 0o644); err != nil {
				return fmt.Errorf("write pre-existing task file: %w", err)
			}
			ctx.lastTaskID = taskID
			return nil
		},
	}
}

// --- Assertion Steps for Creator Attribution ---

var taskHasCreatorDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewRequiredArg("creator"),
)

// TaskHasCreator asserts the task file contains `created_by: <creator>`.
// Required params: "task: <TASK-NNN>", "creator: <name>".
func TaskHasCreator(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task has creator (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := taskHasCreatorDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("TaskHasCreator: %w", err)
			}
			taskID := vals.Value("task")
			creator := vals.Value("creator")
			content, err := os.ReadFile(taskFilePath(ctx, taskID))
			if err != nil {
				return fmt.Errorf("task file %s not found: %w", taskID, err)
			}
			expected := "created_by: " + creator
			if !strings.Contains(string(content), expected) {
				return fmt.Errorf("expected task %s to contain %q\nFile content:\n%s",
					taskID, expected, string(content))
			}
			return nil
		},
	}
}

var boardRowForTaskContainsDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewRequiredArg("text"),
)

// BoardRowForTaskContains asserts the board row for a task contains text.
// Required params: "task: <TASK-NNN>", "text: <text>".
func BoardRowForTaskContains(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("board row for task contains (%s)", strings.Join(params, ", ")),
		Run: func(ctx *Context) error {
			vals, err := boardRowForTaskContainsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("BoardRowForTaskContains: %w", err)
			}
			taskID := vals.Value("task")
			text := vals.Value("text")
			for _, line := range strings.Split(ctx.lastOutput, "\n") {
				if strings.Contains(line, taskID) {
					if strings.Contains(line, text) {
						return nil
					}
					return fmt.Errorf("board row for %s does not contain %q\nRow: %q\nFull output:\n%s",
						taskID, text, strings.TrimSpace(line), ctx.lastOutput)
				}
			}
			return fmt.Errorf("board row for %s not found\nFull output:\n%s", taskID, ctx.lastOutput)
		},
	}
}

// --- Helpers ---

// filterEnv returns a copy of env with all entries whose key matches any of keys removed.
func filterEnv(env []string, keys ...string) []string {
	result := make([]string, 0, len(env))
	for _, entry := range env {
		keep := true
		for _, key := range keys {
			if strings.HasPrefix(entry, key+"=") {
				keep = false
				break
			}
		}
		if keep {
			result = append(result, entry)
		}
	}
	return result
}

// TasksDirIsEmpty asserts that .kanban/tasks/ contains no .md task files.
func TasksDirIsEmpty(params ...string) Step {
	return Step{
		Description: "tasks directory contains no task files",
		Run: func(ctx *Context) error {
			dir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(dir)
			if os.IsNotExist(err) {
				return nil // directory absent → definitively empty
			}
			if err != nil {
				return fmt.Errorf("read tasks dir: %w", err)
			}
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".md") {
					return fmt.Errorf("expected tasks directory to be empty but found: %s", e.Name())
				}
			}
			return nil
		},
	}
}