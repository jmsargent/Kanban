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

// ---- DSL steps for the new-editor-mode feature ----
//
// These step factories extend the DSL to cover US-01: kanban new with no
// positional arguments opens $EDITOR with a blank task template.
//
// All steps invoke the kanban binary as a subprocess (driving port). The only
// direct filesystem access permitted is reading captured template files to
// verify template contents — an observable file output that the binary writes.

// ---- Action steps ----

// IRunKanbanNewInteractive runs "kanban new" with no positional arguments, with
// EDITOR set to the given executable shell script path. The script receives the
// temp file path as its first argument (standard editor contract).
func IRunKanbanNewInteractive(editorScript string) Step {
	return Step{
		Description: "I run kanban new with no arguments using a mock editor",
		Run: func(ctx *Context) error {
			return runKanbanNewWithEditor(ctx, editorScript)
		},
	}
}

// IRunKanbanNewInteractiveNoEditor runs "kanban new" with no positional
// arguments with EDITOR unset and a PATH that does not contain vi or any
// other editor. Git must remain available so the binary can locate the repo
// root and read the git identity before attempting to open the editor.
// Used to verify the "editor unavailable" error path.
func IRunKanbanNewInteractiveNoEditor() Step {
	return Step{
		Description: "I run kanban new with no arguments and no editor available",
		Run: func(ctx *Context) error {
			cmdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Locate the real git binary so we can make it available on PATH
			// without also making vi or other editors available.
			gitPath, err := exec.LookPath("git")
			if err != nil {
				ctx.t.Fatalf("setup: could not locate git binary: %v", err)
			}

			// Create a temp directory containing only a symlink to git.
			// This lets the kanban binary call git (required for RepoRoot and
			// GetIdentity) while no editor binary is on the PATH.
			gitOnlyDir, err := os.MkdirTemp("", "kanban-git-only-*")
			if err != nil {
				ctx.t.Fatalf("setup: create git-only dir: %v", err)
			}
			ctx.t.Cleanup(func() { _ = os.RemoveAll(gitOnlyDir) })

			if err := os.Symlink(gitPath, filepath.Join(gitOnlyDir, "git")); err != nil {
				ctx.t.Fatalf("setup: symlink git: %v", err)
			}

			binDir := filepath.Dir(ctx.binPath)
			env := filterEnv(ctx.env, "EDITOR", "PATH")
			env = append(env, "EDITOR=", "PATH="+binDir+string(os.PathListSeparator)+gitOnlyDir)

			cmd := exec.CommandContext(cmdCtx, ctx.binPath, "new")
			cmd.Dir = ctx.repoDir
			cmd.Env = env

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			captureResult(ctx, cmd.Run(), &stdout, &stderr)
			return nil
		},
	}
}

// ---- Editor script helpers (not Steps) ----

// EditorScriptThatSetsTitle creates a temporary executable shell script that
// uses sed to replace the blank title field in the temp file with the given
// title. Returns the absolute path to the script. A t.Cleanup is registered to
// remove the script directory.
func EditorScriptThatSetsTitle(ctx *Context, title string) (string, error) {
	scriptDir, err := os.MkdirTemp("", "kanban-editor-new-title-*")
	if err != nil {
		return "", fmt.Errorf("create editor script dir: %w", err)
	}
	ctx.t.Cleanup(func() { _ = os.RemoveAll(scriptDir) })

	scriptPath := filepath.Join(scriptDir, "editor.sh")
	escapedTitle := strings.ReplaceAll(title, "/", "\\/")
	script := fmt.Sprintf("#!/bin/sh\nsed -i.bak 's/^title: .*/title: %s/' \"$1\"\n", escapedTitle)
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("write editor script: %w", err)
	}
	return scriptPath, nil
}

// EditorScriptThatSetsFields creates a temporary executable shell script that
// sets title, priority, and assignee fields in the temp file. Returns the
// absolute path to the script. A t.Cleanup is registered to remove the script
// directory.
func EditorScriptThatSetsFields(ctx *Context, title, priority, assignee string) (string, error) {
	scriptDir, err := os.MkdirTemp("", "kanban-editor-new-fields-*")
	if err != nil {
		return "", fmt.Errorf("create editor script dir: %w", err)
	}
	ctx.t.Cleanup(func() { _ = os.RemoveAll(scriptDir) })

	scriptPath := filepath.Join(scriptDir, "editor.sh")
	escTitle := strings.ReplaceAll(title, "/", "\\/")
	escPriority := strings.ReplaceAll(priority, "/", "\\/")
	escAssignee := strings.ReplaceAll(assignee, "/", "\\/")
	script := fmt.Sprintf(
		"#!/bin/sh\n"+
			"sed -i.bak 's/^title: .*/title: %s/' \"$1\"\n"+
			"sed -i.bak 's/^priority: .*/priority: %s/' \"$1\"\n"+
			"sed -i.bak 's/^assignee: .*/assignee: %s/' \"$1\"\n",
		escTitle, escPriority, escAssignee,
	)
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("write editor script: %w", err)
	}
	return scriptPath, nil
}

// EditorScriptThatCapturesTemplate creates a temporary executable shell script
// that copies the temp file contents to a capture file before making no other
// changes. Returns the script path and the capture file path. The capture file
// can be read after the command completes to inspect the blank template that
// the binary prepared.
//
// Because the title remains blank, the binary will exit 2 (empty title). Tests
// using this helper should assert exit code 2 and then read capturePath to
// inspect the template structure.
func EditorScriptThatCapturesTemplate(ctx *Context) (scriptPath, capturePath string, err error) {
	scriptDir, err := os.MkdirTemp("", "kanban-editor-new-capture-*")
	if err != nil {
		return "", "", fmt.Errorf("create editor script dir: %w", err)
	}
	ctx.t.Cleanup(func() { _ = os.RemoveAll(scriptDir) })

	capturePath = filepath.Join(scriptDir, "captured-template.yaml")
	scriptPath = filepath.Join(scriptDir, "editor.sh")
	script := fmt.Sprintf("#!/bin/sh\ncp \"$1\" %q\n", capturePath)
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return "", "", fmt.Errorf("write capture script: %w", err)
	}
	return scriptPath, capturePath, nil
}

// EditorScriptThatLeavesBlankTitle creates a temporary executable shell script
// that makes no changes to the temp file — leaving the title field blank.
// Returns the absolute path to the script.
func EditorScriptThatLeavesBlankTitle(ctx *Context) (string, error) {
	scriptDir, err := os.MkdirTemp("", "kanban-editor-new-blank-*")
	if err != nil {
		return "", fmt.Errorf("create editor script dir: %w", err)
	}
	ctx.t.Cleanup(func() { _ = os.RemoveAll(scriptDir) })

	scriptPath := filepath.Join(scriptDir, "editor.sh")
	// `true` exits 0 without modifying the file — the title stays blank.
	script := "#!/bin/sh\ntrue\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("write editor script: %w", err)
	}
	return scriptPath, nil
}

// ---- Assertion steps ----

// TemplateHasBlankTitleField reads the captured template file at capturePath
// and asserts it contains a blank title field.
func TemplateHasBlankTitleField(capturePath string) Step {
	return Step{
		Description: "template file contains a blank title field",
		Run: func(ctx *Context) error {
			return assertTemplateContains(capturePath, `title: ""`, "blank title field")
		},
	}
}

// TemplateHasBlankPriorityField asserts the captured template contains a blank
// priority field.
func TemplateHasBlankPriorityField(capturePath string) Step {
	return Step{
		Description: "template file contains a blank priority field",
		Run: func(ctx *Context) error {
			return assertTemplateContainsAny(capturePath,
				[]string{`priority: ""`, `priority:`},
				"blank priority field")
		},
	}
}

// TemplateHasBlankAssigneeField asserts the captured template contains a blank
// assignee field.
func TemplateHasBlankAssigneeField(capturePath string) Step {
	return Step{
		Description: "template file contains a blank assignee field",
		Run: func(ctx *Context) error {
			return assertTemplateContainsAny(capturePath,
				[]string{`assignee: ""`, `assignee:`},
				"blank assignee field")
		},
	}
}

// TemplateHasTitleRequiredComment asserts the captured template contains at
// least one comment line indicating that the title field is required.
func TemplateHasTitleRequiredComment(capturePath string) Step {
	return Step{
		Description: "template file contains a comment stating title is required",
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(capturePath)
			if err != nil {
				return fmt.Errorf("read captured template at %s: %w", capturePath, err)
			}
			for _, line := range strings.Split(string(content), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "#") &&
					strings.Contains(strings.ToLower(line), "title") &&
					strings.Contains(strings.ToLower(line), "required") {
					return nil
				}
			}
			return fmt.Errorf("template has no comment indicating title is required\nContent:\n%s", string(content))
		},
	}
}

// TemplateHasNoDueField asserts the captured template does not contain a due
// field (per design decision D4).
func TemplateHasNoDueField(capturePath string) Step {
	return Step{
		Description: "template file does not contain a due field",
		Run: func(ctx *Context) error {
			content, err := os.ReadFile(capturePath)
			if err != nil {
				return fmt.Errorf("read captured template at %s: %w", capturePath, err)
			}
			for _, line := range strings.Split(string(content), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "due:") {
					return fmt.Errorf("template unexpectedly contains due field\nContent:\n%s", string(content))
				}
			}
			return nil
		},
	}
}

// TaskCreatedWithTitle asserts a task file exists in .kanban/tasks/ containing
// the given title in its YAML front matter.
func TaskCreatedWithTitle(title string) Step {
	return Step{
		Description: fmt.Sprintf("a task file exists with title %q", title),
		Run: func(ctx *Context) error {
			tasksDir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return fmt.Errorf("read .kanban/tasks/: %w", err)
			}
			for _, e := range entries {
				if !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				content, err := os.ReadFile(filepath.Join(tasksDir, e.Name()))
				if err != nil {
					continue
				}
				if strings.Contains(string(content), "title: "+title) {
					return nil
				}
			}
			return fmt.Errorf("no task file with title %q found in .kanban/tasks/\nFiles: %v", title, taskFileNames(entries))
		},
	}
}

// TaskCreatedWithFields asserts a task file exists in .kanban/tasks/ containing
// the given title, priority, and assignee values in its YAML front matter.
func TaskCreatedWithFields(title, priority, assignee string) Step {
	return Step{
		Description: fmt.Sprintf("a task file exists with title=%q priority=%q assignee=%q", title, priority, assignee),
		Run: func(ctx *Context) error {
			tasksDir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				return fmt.Errorf("read .kanban/tasks/: %w", err)
			}
			for _, e := range entries {
				if !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				content, err := os.ReadFile(filepath.Join(tasksDir, e.Name()))
				if err != nil {
					continue
				}
				s := string(content)
				if strings.Contains(s, "title: "+title) &&
					strings.Contains(s, "priority: "+priority) &&
					strings.Contains(s, "assignee: "+assignee) {
					return nil
				}
			}
			return fmt.Errorf(
				"no task file with title=%q priority=%q assignee=%q found in .kanban/tasks/\nFiles: %v",
				title, priority, assignee, taskFileNames(entries),
			)
		},
	}
}

// NoTaskFileCreated asserts .kanban/tasks/ contains no .md files. Used to
// verify that a failed editor session did not leave a partial task file.
func NoTaskFileCreated() Step {
	return Step{
		Description: "no task file was created in the workspace",
		Run: func(ctx *Context) error {
			tasksDir := filepath.Join(ctx.repoDir, ".kanban", "tasks")
			entries, err := os.ReadDir(tasksDir)
			if err != nil {
				// Directory not existing means no task files — acceptable.
				return nil
			}
			var mdFiles []string
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".md") {
					mdFiles = append(mdFiles, e.Name())
				}
			}
			if len(mdFiles) > 0 {
				return fmt.Errorf("expected no task files but found: %v", mdFiles)
			}
			return nil
		},
	}
}

// NoTempFileFromNewEditor asserts no .tmp files remain in .kanban/tasks/.
// WriteTempNew writes to os.TempDir (not .kanban/tasks/), so this step
// is a proxy for the general "no partial files remain" guarantee provided by
// the existing NoTempFilesRemain step.
func NoTempFileFromNewEditor() Step {
	return NoTempFilesRemain()
}

// SuccessMessageMatchesNewWithTitle asserts the stdout output contains the
// standard success line format produced by "kanban new <title>": the word
// "Created", a TASK-NNN identifier, and the task title.
func SuccessMessageMatchesNewWithTitle(title string) Step {
	return Step{
		Description: fmt.Sprintf("success message matches kanban new output format for title %q", title),
		Run: func(ctx *Context) error {
			if !strings.Contains(ctx.lastStdout, title) {
				return fmt.Errorf("expected stdout to contain task title %q\nStdout:\n%s", title, ctx.lastStdout)
			}
			if ctx.lastTaskID == "" {
				return fmt.Errorf("expected a TASK-NNN in stdout but found none\nStdout:\n%s", ctx.lastStdout)
			}
			if !strings.Contains(ctx.lastStdout, "Created") {
				return fmt.Errorf("expected stdout to contain \"Created\"\nStdout:\n%s", ctx.lastStdout)
			}
			return nil
		},
	}
}

// HintMessagePresent asserts stdout contains the standard hint line directing
// the developer to reference the task ID in their next commit.
func HintMessagePresent() Step {
	return Step{
		Description: "stdout contains a hint to reference the task ID in the next commit",
		Run: func(ctx *Context) error {
			if !strings.Contains(ctx.lastStdout, "Hint") {
				return fmt.Errorf("expected a Hint line in stdout\nStdout:\n%s", ctx.lastStdout)
			}
			return nil
		},
	}
}

// ---- internal helpers ----

// runKanbanNewWithEditor executes "kanban new" (no positional args) with EDITOR
// set to scriptPath, capturing output into ctx fields.
func runKanbanNewWithEditor(ctx *Context, scriptPath string) error {
	cmdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := append(filterEnv(ctx.env, "EDITOR"), "EDITOR="+scriptPath)

	cmd := exec.CommandContext(cmdCtx, ctx.binPath, "new")
	cmd.Dir = ctx.repoDir
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	captureResult(ctx, cmd.Run(), &stdout, &stderr)
	return nil
}

// assertTemplateContains reads the file at path and asserts it contains the
// given substr. Returns a descriptive error if not found.
func assertTemplateContains(path, substr, description string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read captured template at %s: %w", path, err)
	}
	if !strings.Contains(string(content), substr) {
		return fmt.Errorf("template missing %s (expected %q)\nContent:\n%s", description, substr, string(content))
	}
	return nil
}

// assertTemplateContainsAny reads the file at path and asserts it contains at
// least one of the candidate strings. Returns a descriptive error if none found.
func assertTemplateContainsAny(path string, candidates []string, description string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read captured template at %s: %w", path, err)
	}
	for _, c := range candidates {
		if strings.Contains(string(content), c) {
			return nil
		}
	}
	return fmt.Errorf("template missing %s (expected one of %v)\nContent:\n%s", description, candidates, string(content))
}

// taskFileNames returns a slice of file names from a directory entry slice.
func taskFileNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
