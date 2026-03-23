package dsl

// mermaid_steps.go — DSL step factories for the board-mermaid-export feature.
//
// Driving port: kanban binary invoked as subprocess via run().
// All steps follow the port-to-port principle: no direct calls to internal packages.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ---- Action steps ----

// DeveloperRunsKanbanBoardMermaid runs "kanban board --mermaid" and captures output/exit.
func DeveloperRunsKanbanBoardMermaid() Step {
	return Step{
		Description: "developer runs kanban board --mermaid",
		Run: func(ctx *Context) error {
			run(ctx, "board", "--mermaid")
			return nil
		},
	}
}

// DeveloperRunsKanbanBoardMermaidWithMe runs "kanban board --mermaid --me" and captures output/exit.
func DeveloperRunsKanbanBoardMermaidWithMe() Step {
	return Step{
		Description: "developer runs kanban board --mermaid --me",
		Run: func(ctx *Context) error {
			run(ctx, "board", "--mermaid", "--me")
			return nil
		},
	}
}

// DeveloperRunsKanbanBoardMermaidWithOut runs "kanban board --mermaid --out <abs-path>".
// filename is joined with ctx.repoDir to produce an absolute path.
func DeveloperRunsKanbanBoardMermaidWithOut(filename string) Step {
	return Step{
		Description: fmt.Sprintf("developer runs kanban board --mermaid --out %s", filename),
		Run: func(ctx *Context) error {
			outPath := filepath.Join(ctx.repoDir, filename)
			run(ctx, "board", "--mermaid", "--out", outPath)
			return nil
		},
	}
}

// DeveloperRunsKanbanBoardWithOut runs "kanban board --out <abs-path>" (without --mermaid).
// filename is joined with ctx.repoDir to produce an absolute path.
func DeveloperRunsKanbanBoardWithOut(filename string) Step {
	return Step{
		Description: fmt.Sprintf("developer runs kanban board --out %s (no --mermaid)", filename),
		Run: func(ctx *Context) error {
			outPath := filepath.Join(ctx.repoDir, filename)
			run(ctx, "board", "--out", outPath)
			return nil
		},
	}
}

// DeveloperRunsKanbanBoardMermaidAndJSON runs "kanban board --mermaid --json" and captures output/exit.
func DeveloperRunsKanbanBoardMermaidAndJSON() Step {
	return Step{
		Description: "developer runs kanban board --mermaid --json",
		Run: func(ctx *Context) error {
			run(ctx, "board", "--mermaid", "--json")
			return nil
		},
	}
}

// ---- Setup steps ----

// FileExistsWithContent writes content to filename (relative to ctx.repoDir).
func FileExistsWithContent(filename, content string) Step {
	return Step{
		Description: fmt.Sprintf("file %s exists with specific content", filename),
		Run: func(ctx *Context) error {
			path := filepath.Join(ctx.repoDir, filename)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("mkdir for %s: %w", filename, err)
			}
			return os.WriteFile(path, []byte(content), 0o644)
		},
	}
}

// FileExistsWithKanbanBlock writes a file at filename (relative to ctx.repoDir)
// containing a fenced Mermaid kanban block surrounded by other content.
// This satisfies the "existing kanban block" precondition for AC-11.
func FileExistsWithKanbanBlock(filename string) Step {
	content := "# My Project\n\nSome text above.\n\n```mermaid\nkanban\n  section To Do\n    TASK-OLD@{ label: \"Old task\" }\n```\n\nSome text below.\n"
	return FileExistsWithContent(filename, content)
}

// ---- Assertion steps ----

// StdoutStartsWithMermaidFence asserts ctx.lastStdout begins with "```mermaid".
func StdoutStartsWithMermaidFence() Step {
	return Step{
		Description: "stdout starts with mermaid fenced code block",
		Run: func(ctx *Context) error {
			if !strings.HasPrefix(ctx.lastStdout, "```mermaid") {
				return fmt.Errorf("expected stdout to start with \"```mermaid\"\nStdout:\n%s", ctx.lastStdout)
			}
			return nil
		},
	}
}

// StdoutContainsMermaidKanbanType asserts ctx.lastStdout contains "kanban" as a
// standalone diagram type declaration (line containing only "kanban").
func StdoutContainsMermaidKanbanType() Step {
	return Step{
		Description: "stdout contains kanban as the diagram type",
		Run: func(ctx *Context) error {
			for _, line := range strings.Split(ctx.lastStdout, "\n") {
				if strings.TrimSpace(line) == "kanban" {
					return nil
				}
			}
			return fmt.Errorf("expected stdout to contain \"kanban\" as diagram type\nStdout:\n%s", ctx.lastStdout)
		},
	}
}

// StdoutContainsMermaidSection asserts ctx.lastStdout contains a Mermaid section
// header for the given label (e.g. "section To Do").
func StdoutContainsMermaidSection(label string) Step {
	return Step{
		Description: fmt.Sprintf("stdout contains mermaid section %q", label),
		Run: func(ctx *Context) error {
			needle := "section " + label
			if !strings.Contains(ctx.lastStdout, needle) {
				return fmt.Errorf("expected stdout to contain %q\nStdout:\n%s", needle, ctx.lastStdout)
			}
			return nil
		},
	}
}

// StdoutContainsMermaidNode asserts ctx.lastStdout contains a Mermaid node entry
// for the given task ID (e.g. "TASK-001@{").
func StdoutContainsMermaidNode(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("stdout contains mermaid node for %s", taskID),
		Run: func(ctx *Context) error {
			needle := taskID + "@{"
			if !strings.Contains(ctx.lastStdout, needle) {
				return fmt.Errorf("expected stdout to contain mermaid node %q\nStdout:\n%s", needle, ctx.lastStdout)
			}
			return nil
		},
	}
}

// StdoutDoesNotContainMermaidNode asserts ctx.lastStdout does not contain a
// Mermaid node entry for the given task ID.
func StdoutDoesNotContainMermaidNode(taskID string) Step {
	return Step{
		Description: fmt.Sprintf("stdout does not contain mermaid node for %s", taskID),
		Run: func(ctx *Context) error {
			needle := taskID + "@{"
			if strings.Contains(ctx.lastStdout, needle) {
				return fmt.Errorf("expected stdout NOT to contain mermaid node %q\nStdout:\n%s", needle, ctx.lastStdout)
			}
			return nil
		},
	}
}

// StdoutIsEmpty asserts ctx.lastStdout is empty (no bytes written to stdout).
func StdoutIsEmpty() Step {
	return Step{
		Description: "stdout is empty",
		Run: func(ctx *Context) error {
			if ctx.lastStdout != "" {
				return fmt.Errorf("expected stdout to be empty\nStdout:\n%s", ctx.lastStdout)
			}
			return nil
		},
	}
}

// FileExists asserts the file at filename (relative to ctx.repoDir) exists.
func FileExists(filename string) Step {
	return Step{
		Description: fmt.Sprintf("file %s exists", filename),
		Run: func(ctx *Context) error {
			path := filepath.Join(ctx.repoDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return fmt.Errorf("expected file %s to exist but it does not", path)
			}
			return nil
		},
	}
}

// FileContainsText asserts the file at filename (relative to ctx.repoDir) contains text.
func FileContainsText(filename, text string) Step {
	return Step{
		Description: fmt.Sprintf("file %s contains %q", filename, text),
		Run: func(ctx *Context) error {
			path := filepath.Join(ctx.repoDir, filename)
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read file %s: %w", path, err)
			}
			if !strings.Contains(string(content), text) {
				return fmt.Errorf("expected file %s to contain %q\nContent:\n%s", filename, text, string(content))
			}
			return nil
		},
	}
}

// FileDoesNotContainText asserts the file at filename does not contain text.
func FileDoesNotContainText(filename, text string) Step {
	return Step{
		Description: fmt.Sprintf("file %s does not contain %q", filename, text),
		Run: func(ctx *Context) error {
			path := filepath.Join(ctx.repoDir, filename)
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read file %s: %w", path, err)
			}
			if strings.Contains(string(content), text) {
				return fmt.Errorf("expected file %s NOT to contain %q\nContent:\n%s", filename, text, string(content))
			}
			return nil
		},
	}
}

// FileContentEquals asserts the file at filename has exactly the expected content.
func FileContentEquals(filename, expected string) Step {
	return Step{
		Description: fmt.Sprintf("file %s has expected content", filename),
		Run: func(ctx *Context) error {
			path := filepath.Join(ctx.repoDir, filename)
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read file %s: %w", path, err)
			}
			if string(content) != expected {
				return fmt.Errorf("file %s content mismatch\nExpected:\n%s\nActual:\n%s", filename, expected, string(content))
			}
			return nil
		},
	}
}
