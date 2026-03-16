package dsl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExitCodeIs asserts ctx.lastExit == code.
func ExitCodeIs(code int) Step {
	return Step{
		Description: fmt.Sprintf("exit code is %d", code),
		Run: func(ctx *Context) error {
			if ctx.lastExit != code {
				return fmt.Errorf("expected exit code %d, got %d\nOutput:\n%s", code, ctx.lastExit, ctx.lastOutput)
			}
			return nil
		},
	}
}

// StdoutContains asserts ctx.lastStdout contains text.
func StdoutContains(text string) Step {
	return Step{
		Description: fmt.Sprintf("stdout contains %q", text),
		Run: func(ctx *Context) error {
			if !strings.Contains(ctx.lastStdout, text) {
				return fmt.Errorf("expected stdout to contain %q\nStdout:\n%s", text, ctx.lastStdout)
			}
			return nil
		},
	}
}

// StderrContains asserts ctx.lastStderr contains text.
func StderrContains(text string) Step {
	return Step{
		Description: fmt.Sprintf("stderr contains %q", text),
		Run: func(ctx *Context) error {
			if !strings.Contains(ctx.lastStderr, text) {
				return fmt.Errorf("expected stderr to contain %q\nStderr:\n%s", text, ctx.lastStderr)
			}
			return nil
		},
	}
}

// OutputContains asserts ctx.lastOutput (stdout+stderr) contains text.
func OutputContains(text string) Step {
	return Step{
		Description: fmt.Sprintf("output contains %q", text),
		Run: func(ctx *Context) error {
			if !strings.Contains(ctx.lastOutput, text) {
				return fmt.Errorf("expected output to contain %q\nOutput:\n%s", text, ctx.lastOutput)
			}
			return nil
		},
	}
}

// OutputIsValidJSON asserts ctx.lastOutput is parseable as JSON.
func OutputIsValidJSON() Step {
	return Step{
		Description: "output is valid JSON",
		Run: func(ctx *Context) error {
			var v interface{}
			if err := json.Unmarshal([]byte(strings.TrimSpace(ctx.lastOutput)), &v); err != nil {
				return fmt.Errorf("output is not valid JSON: %w\nOutput:\n%s", err, ctx.lastOutput)
			}
			return nil
		},
	}
}

// JSONHasFields asserts the JSON array in ctx.lastOutput contains an object
// with all named fields present. fields is a comma-separated list.
func JSONHasFields(fields string) Step {
	return Step{
		Description: fmt.Sprintf("JSON has fields %s", fields),
		Run: func(ctx *Context) error {
			var tasks []map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimSpace(ctx.lastOutput)), &tasks); err != nil {
				return fmt.Errorf("output is not a JSON array: %w", err)
			}
			if len(tasks) == 0 {
				return fmt.Errorf("JSON array is empty")
			}
			required := strings.Split(fields, ",")
			for _, field := range required {
				field = strings.TrimSpace(field)
				if _, ok := tasks[0][field]; !ok {
					return fmt.Errorf("JSON object missing field %q — present: %v", field, tasks[0])
				}
			}
			return nil
		},
	}
}

// taskFilePath returns the path to the task file for the given taskID.
func taskFilePath(ctx *Context, taskID string) string {
	return filepath.Join(ctx.repoDir, ".kanban", "tasks", taskID+".md")
}

// TaskHasStatus reads the task file for taskID and asserts status == expected.
func TaskHasStatus(taskID, expected string) Step {
	return Step{
		Description: fmt.Sprintf("task %s has status %q", taskID, expected),
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(taskFilePath(ctx, taskID))
			if err != nil {
				return fmt.Errorf("task file %s not found: %w", taskID, err)
			}
			if !strings.Contains(string(content), "status: "+expected) {
				return fmt.Errorf("expected task %s to have status %q\nFile content:\n%s", taskID, expected, string(content))
			}
			return nil
		},
	}
}

// TaskStatusRemains is an alias for TaskHasStatus with a distinct description
// for use in "status unchanged" assertions.
func TaskStatusRemains(taskID, expected string) Step {
	return Step{
		Description: fmt.Sprintf("task %s status remains %q", taskID, expected),
		Run:         TaskHasStatus(taskID, expected).Run,
	}
}

// TaskFilePresent asserts the file .kanban/tasks/<taskID>.md is present.
func TaskFilePresent(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("task file %s.md is present", taskID),
		Run: func(ctx *Context) error {
			if _, err := os.Stat(taskFilePath(ctx, taskID)); os.IsNotExist(err) {
				return fmt.Errorf("expected task file %s.md to exist but it does not", taskID)
			}
			return nil
		},
	}
}

// TaskFileRemoved asserts the file .kanban/tasks/<taskID>.md is absent.
func TaskFileRemoved(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("task file %s.md is removed", taskID),
		Run: func(ctx *Context) error {
			if _, err := os.Stat(taskFilePath(ctx, taskID)); err == nil {
				return fmt.Errorf("expected task file %s.md to be removed but it still exists", taskID)
			}
			return nil
		},
	}
}

// BoardShowsTaskUnder runs "kanban board" and asserts that title appears
// after the given heading line in the output.
func BoardShowsTaskUnder(title, heading string) Step {
	return Step{
		Description: fmt.Sprintf("board shows %q under %s", title, heading),
		Run: func(ctx *Context) error {
			run(ctx, "board")
			output := ctx.lastOutput
			lines := strings.Split(output, "\n")
			headingLine := -1
			for i, line := range lines {
				if strings.Contains(strings.ToUpper(line), strings.ToUpper(heading)) {
					headingLine = i
					break
				}
			}
			if headingLine == -1 {
				return fmt.Errorf("heading %q not found in board output:\n%s", heading, output)
			}
			for i := headingLine + 1; i < len(lines); i++ {
				if strings.Contains(lines[i], title) {
					return nil
				}
				// Stop searching when the next heading is reached.
				if i > headingLine+1 {
					upper := strings.ToUpper(lines[i])
					if strings.Contains(upper, "TODO") ||
						strings.Contains(upper, "IN PROGRESS") ||
						strings.Contains(upper, "DONE") {
						break
					}
				}
			}
			return fmt.Errorf("task %q not found under heading %q in board output:\n%s", title, heading, output)
		},
	}
}

// BoardNotListsTask runs "kanban board" and asserts taskID does not appear.
func BoardNotListsTask(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("board does not list %s", taskID),
		Run: func(ctx *Context) error {
			run(ctx, "board")
			if strings.Contains(ctx.lastOutput, taskID) {
				return fmt.Errorf("expected board to not list %s but found it:\n%s", taskID, ctx.lastOutput)
			}
			return nil
		},
	}
}

// GitCommitExitCodeIs asserts ctx.lastExit == code after a git commit action.
// It is a semantic alias for ExitCodeIs with a more descriptive name.
func GitCommitExitCodeIs(code int) Step {
	return Step{
		Description: fmt.Sprintf("git commit exit code is %d", code),
		Run:         ExitCodeIs(code).Run,
	}
}

// WorkspaceReady asserts .kanban/tasks/ exists in the repo directory.
func WorkspaceReady() Step {
	return Step{
		Description: "kanban workspace is ready for use",
		Run: func(ctx *Context) error {
			tasksDir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
				return fmt.Errorf("expected .kanban/tasks/ to exist but it does not")
			}
			return nil
		},
	}
}

// ConfigFileHasDefaults asserts .kanban/config exists and contains TASK- and todo.
func ConfigFileHasDefaults() Step {
	return Step{
		Description: "config file has default task pattern and column list",
		Run: func(ctx *Context) error {
			configPath := filepath.Join(ctx.repoDir, ".kanban", "config")
			content, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("config file not found: %w", err)
			}
			s := string(content)
			if !strings.Contains(s, "TASK-") {
				return fmt.Errorf("config missing task pattern (TASK-)\nContent:\n%s", s)
			}
			if !strings.Contains(s, "todo") {
				return fmt.Errorf("config missing column definitions (todo)\nContent:\n%s", s)
			}
			return nil
		},
	}
}

// HookLogInGitignore asserts .gitignore contains "hook.log".
func HookLogInGitignore() Step {
	return Step{
		Description: "hook log path is in .gitignore",
		Run: func(ctx *Context) error {
			gitignorePath := filepath.Join(ctx.repoDir, ".gitignore")
			content, err := os.ReadFile(gitignorePath)
			if err != nil {
				return fmt.Errorf(".gitignore not found: %w", err)
			}
			if !strings.Contains(string(content), "hook.log") {
				return fmt.Errorf(".gitignore does not contain hook.log\nContent:\n%s", string(content))
			}
			return nil
		},
	}
}

// NoTempFilesRemain asserts no *.tmp files exist in .kanban/tasks/.
func NoTempFilesRemain() Step {
	return Step{
		Description: "no partial or temporary files remain in the tasks directory",
		Run: func(ctx *Context) error {
			dir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil // directory may not exist; that is fine
			}
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".tmp") {
					return fmt.Errorf("temporary file found in tasks directory: %s", e.Name())
				}
			}
			return nil
		},
	}
}

// UpdatedTaskCommitted asserts the recent git log contains a commit referencing "kanban".
func UpdatedTaskCommitted() Step {
	return Step{
		Description: "updated task file is committed back to the repository",
		Run: func(ctx *Context) error {
			out, err := gitCmd(ctx, "log", "--oneline", "-5")
			if err != nil {
				return fmt.Errorf("git log failed: %w", err)
			}
			if !strings.Contains(strings.ToLower(out), "kanban") {
				return fmt.Errorf("expected a kanban commit in git log but found:\n%s", out)
			}
			return nil
		},
	}
}

// NoAutoCommitFromDelete asserts the most recent git log entry does not reference
// "delete" or "remove task".
func NoAutoCommitFromDelete() Step {
	return Step{
		Description: "git repository has no new commits from the delete operation",
		Run: func(ctx *Context) error {
			out, err := gitCmd(ctx, "log", "--oneline", "-1")
			if err != nil {
				return fmt.Errorf("git log failed: %w", err)
			}
			lower := strings.ToLower(out)
			if strings.Contains(lower, "delete") || strings.Contains(lower, "remove task") {
				return fmt.Errorf("expected no auto-commit from delete but found:\n%s", out)
			}
			return nil
		},
	}
}

// NoKanbanOutputLines asserts no line in ctx.lastOutput starts with "kanban:".
func NoKanbanOutputLines() Step {
	return Step{
		Description: "output contains no kanban: prefix lines",
		Run: func(ctx *Context) error {
			for _, line := range strings.Split(ctx.lastOutput, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "kanban:") {
					return fmt.Errorf("expected no kanban output lines but found: %q", line)
				}
			}
			return nil
		},
	}
}

// NoTransitionLines asserts no line in ctx.lastOutput contains both "moved" and "->".
func NoTransitionLines() Step {
	return Step{
		Description: "output contains no kanban transition lines",
		Run: func(ctx *Context) error {
			for _, line := range strings.Split(ctx.lastOutput, "\n") {
				if strings.Contains(line, "moved") && strings.Contains(line, "->") {
					return fmt.Errorf("expected no transition output but found: %q", line)
				}
			}
			return nil
		},
	}
}

// ansiPattern matches ANSI CSI colour escape sequences.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// NoANSIEscapeCodes asserts ctx.lastOutput contains no ANSI escape sequences.
func NoANSIEscapeCodes() Step {
	return Step{
		Description: "output contains no ANSI colour escape sequences",
		Run: func(ctx *Context) error {
			if ansiPattern.MatchString(ctx.lastOutput) {
				return fmt.Errorf("output contains ANSI escape codes:\n%s", ctx.lastOutput)
			}
			return nil
		},
	}
}

// spinnerChars lists the Unicode Braille spinner characters to detect.
var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NoSpinnerChars asserts ctx.lastOutput contains no Unicode spinner characters.
func NoSpinnerChars() Step {
	return Step{
		Description: "output contains no spinner characters",
		Run: func(ctx *Context) error {
			for _, s := range spinnerChars {
				if strings.Contains(ctx.lastOutput, s) {
					return fmt.Errorf("output contains spinner character %q", s)
				}
			}
			return nil
		},
	}
}
