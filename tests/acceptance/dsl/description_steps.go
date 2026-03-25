package dsl

// description_steps.go — DSL step factories for the add-task-description feature.
//
// Driving port: kanban binary invoked as subprocess via run().
// All steps follow the port-to-port principle: no direct calls to internal packages.
//
// IMPORTANT: task.Description is stored as the Markdown body of the task file
// (the plain text below the closing "---" front matter delimiter). Assertion
// steps in this file read the body section, NOT a "description:" YAML key.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ---- Editor script helpers (not Steps) ----

// EditorScriptThatSetsTitleAndDescription creates a temporary executable shell
// script that uses sed to replace the blank title and description fields in the
// YAML temp file. Returns the absolute path to the script. A t.Cleanup is
// registered to remove the script directory.
//
// The description field in the editFields temp file uses the key "description:"
// (this is the editor temp file, not the final task file — the final task file
// stores description as the Markdown body).
func EditorScriptThatSetsTitleAndDescription(ctx *Context, title, description string) (string, error) {
	scriptDir, err := os.MkdirTemp("", "kanban-editor-new-titledesc-*")
	if err != nil {
		return "", fmt.Errorf("create editor script dir: %w", err)
	}
	ctx.t.Cleanup(func() { _ = os.RemoveAll(scriptDir) })

	scriptPath := filepath.Join(scriptDir, "editor.sh")
	escTitle := strings.ReplaceAll(title, "/", "\\/")
	escDesc := strings.ReplaceAll(description, "/", "\\/")
	escDesc = strings.ReplaceAll(escDesc, "&", "\\&")
	script := fmt.Sprintf(
		"#!/bin/sh\n"+
			"sed -i.bak 's/^title: .*/title: %s/' \"$1\"\n"+
			"sed -i.bak 's/^description: .*/description: %s/' \"$1\"\n",
		escTitle, escDesc,
	)
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("write editor script: %w", err)
	}
	return scriptPath, nil
}

// ---- Template assertion steps ----

// TemplateHasBlankDescriptionField reads the captured template file at
// capturePath and asserts it contains a blank description field
// (description: "").
func TemplateHasBlankDescriptionField(capturePath string) Step {
	return Step{
		Description: "template file contains a blank description field",
		Run: func(ctx *Context) error {
			return assertTemplateContains(capturePath, `description: ""`, "blank description field")
		},
	}
}

// ---- Task body assertion steps ----

// TaskBodyContains reads the task file for the most recently created task
// (ctx.LastTaskID()) and asserts the Markdown body (text below the closing
// "---" delimiter) contains the given substring.
func TaskBodyContains(text string) Step {
	return Step{
		Description: fmt.Sprintf("task body contains %q", text),
		Run: func(ctx *Context) error {
			taskID := ctx.LastTaskID()
			if taskID == "" {
				return fmt.Errorf("no task ID in context — cannot read task body")
			}
			body, err := taskBody(ctx, taskID)
			if err != nil {
				return err
			}
			if !strings.Contains(body, text) {
				return fmt.Errorf("expected task body to contain %q\nBody:\n%s", text, body)
			}
			return nil
		},
	}
}

// TaskBodyIsEmpty asserts the task body (below closing "---") for the most
// recently created task is empty or contains only whitespace. Used when
// description was not set.
func TaskBodyIsEmpty() Step {
	return Step{
		Description: "task body is empty or whitespace only",
		Run: func(ctx *Context) error {
			taskID := ctx.LastTaskID()
			if taskID == "" {
				return fmt.Errorf("no task ID in context — cannot read task body")
			}
			body, err := taskBody(ctx, taskID)
			if err != nil {
				return err
			}
			if strings.TrimSpace(body) != "" {
				return fmt.Errorf("expected task body to be empty but got:\n%s", body)
			}
			return nil
		},
	}
}

// ---- CLI action steps ----

// IRunKanbanNewWithDescription runs "kanban new <title> --description <desc>".
func IRunKanbanNewWithDescription(title, description string) Step {
	return Step{
		Description: fmt.Sprintf("I run kanban new %q --description %q", title, description),
		Run: func(ctx *Context) error {
			run(ctx, "new", title, "--description", description)
			return nil
		},
	}
}

// ---- Internal helpers ----

// taskBody reads the Markdown body from a task file (text after the second
// "---" delimiter). Returns empty string if no body exists. The body may begin
// with a newline immediately after the closing delimiter — that leading newline
// is included in the returned string so callers can assert on exact content
// using strings.Contains.
func taskBody(ctx *Context, taskID string) (string, error) {
	// Task files are named <taskID>.md and reside in .kanban/tasks/. However,
	// the task file name includes a slugified title suffix, so we must scan
	// the directory for a file whose name starts with taskID.
	tasksDir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return "", fmt.Errorf("read .kanban/tasks/: %w", err)
	}

	var matchPath string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), taskID) && strings.HasSuffix(e.Name(), ".md") {
			matchPath = filepath.Join(tasksDir, e.Name())
			break
		}
	}
	if matchPath == "" {
		return "", fmt.Errorf("no task file found for %s in .kanban/tasks/", taskID)
	}

	content, err := os.ReadFile(matchPath)
	if err != nil {
		return "", fmt.Errorf("read task file %s: %w", matchPath, err)
	}

	// The file format is:
	//   ---
	//   <YAML front matter>
	//   ---
	//   <Markdown body>
	//
	// We find the second occurrence of "---" on its own line and return
	// everything after it.
	text := string(content)
	lines := strings.Split(text, "\n")
	delimCount := 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			delimCount++
			if delimCount == 2 {
				// Body starts on the next line.
				body := strings.Join(lines[i+1:], "\n")
				return body, nil
			}
		}
	}
	// No second delimiter found — no body.
	return "", nil
}
