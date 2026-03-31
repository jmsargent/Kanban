package dsl

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jmsargent/kanban/pkg/simpledsl"
)

var taskCreationFailsDSL = simpledsl.NewDslParams(
	simpledsl.NewOptionalArg("error"),
)

// TaskCreationFails asserts that the last response re-rendered the add-task form
// (i.e. no redirect occurred) and optionally that the given error message appears.
// Optional param: "error: <message>".
func TaskCreationFails(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task creation fails (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := taskCreationFailsDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("TaskCreationFails: %w", err)
			}
			if !strings.Contains(ctx.LastBody, "add-task-form") {
				return fmt.Errorf("TaskCreationFails: expected add-task form in response, got:\n%s", ctx.LastBody)
			}
			if errMsg := vals.Value("error"); errMsg != "" {
				if !strings.Contains(ctx.LastBody, errMsg) {
					return fmt.Errorf("TaskCreationFails: expected error %q in response, got:\n%s", errMsg, ctx.LastBody)
				}
			}
			return nil
		},
	}
}

// NoNewTaskInRepo asserts that no task files were created in ctx.RepoDir during
// the test (the tasks directory remains empty).
func NoNewTaskInRepo(params ...string) Step {
	return Step{
		Description: "no new task in repo",
		Run: func(ctx *WebContext) error {
			if ctx.RepoDir == "" {
				return fmt.Errorf("NoNewTaskInRepo: RepoDir not set; call ARepoWithNoTasks first")
			}
			tasksDir := filepath.Join(ctx.RepoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return fmt.Errorf("NoNewTaskInRepo: read tasks dir: %w", err)
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
					return fmt.Errorf("NoNewTaskInRepo: unexpected task file found: %s", entry.Name())
				}
			}
			return nil
		},
	}
}

var iAddTaskDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
	simpledsl.NewOptionalArg("description"),
	simpledsl.NewOptionalArg("priority"),
	simpledsl.NewOptionalArg("assignee"),
)

var taskExistsInRepoDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
	simpledsl.NewOptionalArg("priority"),
	simpledsl.NewOptionalArg("assignee"),
	simpledsl.NewOptionalArg("created_by"),
	simpledsl.NewOptionalArg("description"),
	simpledsl.NewOptionalArg("status"),
)

// IAddTask submits the add-task form (POST /task) with the provided fields.
// Requires AnAuthenticatedUser to have been called so ctx.HTTPDriver holds a
// session cookie. After submission the board response is stored in ctx.LastBody.
// Required param: "title: <title>".
// Optional params: "description: <text>", "priority: <priority>", "assignee: <name>".
func IAddTask(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("I add a task (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := iAddTaskDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("IAddTask: %w", err)
			}
			if ctx.HTTPDriver == nil {
				return fmt.Errorf("IAddTask: no authenticated HTTP driver; call AnAuthenticatedUser first")
			}

			formData := url.Values{
				"title":       {vals.Value("title")},
				"description": {vals.Value("description")},
				"priority":    {vals.Value("priority")},
				"assignee":    {vals.Value("assignee")},
			}

			resp, err := ctx.HTTPDriver.POST("/task", formData)
			if err != nil {
				return fmt.Errorf("IAddTask: POST /task: %w", err)
			}
			ctx.LastBody = resp.Body
			return nil
		},
	}
}

// TaskFileIsValidFormat asserts that the task file created in the repo follows
// Markdown + YAML front matter format (ADR-002): starts with "---", contains
// a YAML block with id:, title: fields, and is closed with "---".
func TaskFileIsValidFormat(params ...string) Step {
	return Step{
		Description: "task file is valid format",
		Run: func(ctx *WebContext) error {
			if ctx.RepoDir == "" {
				return fmt.Errorf("TaskFileIsValidFormat: RepoDir not set; call ARepoWithNoTasks first")
			}
			tasksDir := filepath.Join(ctx.RepoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return fmt.Errorf("TaskFileIsValidFormat: read tasks dir: %w", err)
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				content, err := os.ReadFile(filepath.Join(tasksDir, entry.Name()))
				if err != nil {
					return fmt.Errorf("TaskFileIsValidFormat: read file %s: %w", entry.Name(), err)
				}
				text := string(content)
				if !strings.HasPrefix(text, "---\n") {
					return fmt.Errorf("TaskFileIsValidFormat: file %s does not start with '---'", entry.Name())
				}
				rest := text[4:]
				closeIdx := strings.Index(rest, "\n---")
				if closeIdx < 0 {
					return fmt.Errorf("TaskFileIsValidFormat: file %s has no closing '---' delimiter", entry.Name())
				}
				yamlBlock := rest[:closeIdx]
				for _, required := range []string{"id:", "title:"} {
					if !strings.Contains(yamlBlock, required) {
						return fmt.Errorf("TaskFileIsValidFormat: file %s missing field %q in front matter", entry.Name(), required)
					}
				}
				return nil
			}
			return fmt.Errorf("TaskFileIsValidFormat: no task .md file found in %s", tasksDir)
		},
	}
}

var sequentialIDPattern = regexp.MustCompile(`^TASK-\d{3}$`)

// TaskHasSequentialID asserts that the task file in the repo has an ID matching
// the TASK-NNN sequential pattern (zero-padded 3-digit number).
func TaskHasSequentialID(params ...string) Step {
	return Step{
		Description: "task has sequential ID",
		Run: func(ctx *WebContext) error {
			if ctx.RepoDir == "" {
				return fmt.Errorf("TaskHasSequentialID: RepoDir not set; call ARepoWithNoTasks first")
			}
			tasksDir := filepath.Join(ctx.RepoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return fmt.Errorf("TaskHasSequentialID: read tasks dir: %w", err)
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				content, err := os.ReadFile(filepath.Join(tasksDir, entry.Name()))
				if err != nil {
					return fmt.Errorf("TaskHasSequentialID: read file %s: %w", entry.Name(), err)
				}
				text := string(content)
				idPattern := regexp.MustCompile(`(?m)^id:\s*(\S+)`)
				matches := idPattern.FindStringSubmatch(text)
				if matches == nil {
					return fmt.Errorf("TaskHasSequentialID: no id: field found in %s", entry.Name())
				}
				id := strings.TrimSpace(matches[1])
				if !sequentialIDPattern.MatchString(id) {
					return fmt.Errorf("TaskHasSequentialID: id %q does not match TASK-NNN pattern in %s", id, entry.Name())
				}
				return nil
			}
			return fmt.Errorf("TaskHasSequentialID: no task .md file found in %s", tasksDir)
		},
	}
}

// TaskExistsInRepo asserts that a task file exists in ctx.RepoDir whose
// front matter contains all the provided field values.
// Required param: "title: <title>".
// Optional params: "priority: <priority>", "assignee: <name>", "created_by: <name>",
// "description: <text>".
func TaskExistsInRepo(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("task exists in repo (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := taskExistsInRepoDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("TaskExistsInRepo: %w", err)
			}
			if ctx.RepoDir == "" {
				return fmt.Errorf("TaskExistsInRepo: RepoDir not set; call ARepoWithNoTasks or ARepoWithTasks first")
			}

			tasksDir := filepath.Join(ctx.RepoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return fmt.Errorf("TaskExistsInRepo: read tasks dir: %w", err)
			}

			title := vals.Value("title")
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				content, err := os.ReadFile(filepath.Join(tasksDir, entry.Name()))
				if err != nil {
					continue
				}
				text := string(content)
				if !strings.Contains(text, "title: "+title) {
					continue
				}
				// Title matched — now assert all other specified fields.
				for _, field := range []string{"priority", "assignee", "created_by", "description"} {
					expected := vals.Value(field)
					if expected == "" {
						continue
					}
					needle := field + ": " + expected
					if !strings.Contains(text, needle) {
						return fmt.Errorf("TaskExistsInRepo: field %q: expected %q in task file %s:\n%s",
							field, expected, entry.Name(), text)
					}
				}
				// Status is tracked in transitions.log; new task files omit the
				// status field (defaulting to "todo"). Accept the assertion when:
				//   - expected status is "todo" and no "status:" key is present, OR
				//   - the literal "status: <expected>" line is present.
				if expectedStatus := vals.Value("status"); expectedStatus != "" {
					hasStatusField := strings.Contains(text, "\nstatus: ")
					literalMatch := strings.Contains(text, "status: "+expectedStatus)
					defaultTodo := expectedStatus == "todo" && !hasStatusField
					if !literalMatch && !defaultTodo {
						return fmt.Errorf("TaskExistsInRepo: field %q: expected %q in task file %s:\n%s",
							"status", expectedStatus, entry.Name(), text)
					}
				}
				return nil
			}

			return fmt.Errorf("TaskExistsInRepo: no task file found with title %q in %s", title, tasksDir)
		},
	}
}

var remoteRepoContainsTaskDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("title"),
)

// RemoteRepoContainsTask clones the bare remote (ctx.RemoteDir) into a temp
// directory and asserts that a task file containing the given title exists.
// Required param: "title: <title>".
func RemoteRepoContainsTask(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("remote repo contains task (%s)", strings.Join(params, ", ")),
		Run: func(ctx *WebContext) error {
			vals, err := remoteRepoContainsTaskDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("RemoteRepoContainsTask: %w", err)
			}
			if ctx.RemoteDir == "" {
				return fmt.Errorf("RemoteRepoContainsTask: RemoteDir not set; call ARepoWithRemote first")
			}

			cloneDir, err := os.MkdirTemp("", "kanban-remote-verify-*")
			if err != nil {
				return fmt.Errorf("RemoteRepoContainsTask: create temp dir: %w", err)
			}
			defer func() { _ = os.RemoveAll(cloneDir) }()

			cloneCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			var buf bytes.Buffer
			cmd := exec.CommandContext(cloneCtx, "git", "clone", ctx.RemoteDir, cloneDir)
			cmd.Stdout = &buf
			cmd.Stderr = &buf
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("RemoteRepoContainsTask: git clone: %w\n%s", err, buf.String())
			}

			tasksDir := filepath.Join(cloneDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return fmt.Errorf("RemoteRepoContainsTask: read tasks dir in clone: %w", err)
			}

			title := vals.Value("title")
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				content, err := os.ReadFile(filepath.Join(tasksDir, entry.Name()))
				if err != nil {
					continue
				}
				if strings.Contains(string(content), "title: "+title) {
					return nil
				}
			}

			return fmt.Errorf("RemoteRepoContainsTask: no task file with title %q found in remote clone %s", title, tasksDir)
		},
	}
}
