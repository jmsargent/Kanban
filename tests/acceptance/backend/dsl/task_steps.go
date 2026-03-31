package dsl

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
