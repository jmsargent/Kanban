package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmsargent/kanban/internal/domain"
)

// TestWriteMermaidToFile_NewFile validates writeMermaidToFile when the target
// file does not exist.
// Test Budget: 2 behaviors (creates file with content | returns nil) x 2 = 4 max.
func TestWriteMermaidToFile_NewFile(t *testing.T) {
	t.Run("creates file containing the content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "BOARD.md")
		content := "```mermaid\nkanban\n```\n"

		_ = writeMermaidToFile(path, content)

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected file to be created at %s: %v", path, err)
		}
		if string(got) != content {
			t.Errorf("file content = %q, want %q", string(got), content)
		}
	})

	t.Run("returns nil error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "BOARD.md")

		err := writeMermaidToFile(path, "```mermaid\nkanban\n```\n")

		if err != nil {
			t.Errorf("writeMermaidToFile returned error %v, want nil", err)
		}
	})
}

// TestWriteMermaidToFile_ExistingFileNoBlock validates writeMermaidToFile when the
// target file exists but contains no Mermaid kanban block.
// Test Budget: 2 behaviors (returns ErrNoKanbanBlock | file content unchanged) x 2 = 4 max.
func TestWriteMermaidToFile_ExistingFileNoBlock(t *testing.T) {
	const originalContent = "# My Project\n\nSome markdown without a kanban block.\n"

	t.Run("returns ErrNoKanbanBlock", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "README.md")
		if err := os.WriteFile(path, []byte(originalContent), 0o644); err != nil {
			t.Fatalf("setup: write file: %v", err)
		}

		err := writeMermaidToFile(path, "```mermaid\nkanban\n```\n")

		if err != ErrNoKanbanBlock {
			t.Errorf("writeMermaidToFile returned %v, want ErrNoKanbanBlock", err)
		}
	})

	t.Run("file content is unchanged", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "README.md")
		if err := os.WriteFile(path, []byte(originalContent), 0o644); err != nil {
			t.Fatalf("setup: write file: %v", err)
		}

		_ = writeMermaidToFile(path, "```mermaid\nkanban\n```\n")

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read file after call: %v", err)
		}
		if string(got) != originalContent {
			t.Errorf("file content changed\nGot:  %q\nWant: %q", string(got), originalContent)
		}
	})
}

// TestWriteMermaidToFile_ExistingFileWithBlock validates writeMermaidToFile when the
// target file exists and contains a Mermaid kanban block.
// Test Budget: 3 behaviors (returns nil | replaces block content | preserves surrounding content) x 2 = 6 max.
func TestWriteMermaidToFile_ExistingFileWithBlock(t *testing.T) {
	const fileWithBlock = "# My Project\n\nSome text above.\n\n```mermaid\nkanban\n  section To Do\n    TASK-OLD@{ label: \"Old task\" }\n```\n\nSome text below.\n"
	const newBlock = "```mermaid\nkanban\n  section To Do\n    TASK-001@{ label: \"Fix login bug\" }\n```\n"

	setup := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "README.md")
		if err := os.WriteFile(path, []byte(fileWithBlock), 0o644); err != nil {
			t.Fatalf("setup: write file: %v", err)
		}
		return path
	}

	t.Run("returns nil when kanban block replaced", func(t *testing.T) {
		path := setup(t)

		err := writeMermaidToFile(path, newBlock)

		if err != nil {
			t.Errorf("writeMermaidToFile returned %v, want nil", err)
		}
	})

	t.Run("new block content appears in file", func(t *testing.T) {
		path := setup(t)

		_ = writeMermaidToFile(path, newBlock)

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read file after call: %v", err)
		}
		if !strings.Contains(string(got), "TASK-001") {
			t.Errorf("expected file to contain TASK-001 after replacement\nGot:\n%s", string(got))
		}
		if strings.Contains(string(got), "TASK-OLD") {
			t.Errorf("expected TASK-OLD to be absent after replacement\nGot:\n%s", string(got))
		}
	})

	t.Run("surrounding content is preserved", func(t *testing.T) {
		path := setup(t)

		_ = writeMermaidToFile(path, newBlock)

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read file after call: %v", err)
		}
		content := string(got)
		for _, want := range []string{"# My Project", "Some text above.", "Some text below."} {
			if !strings.Contains(content, want) {
				t.Errorf("expected file to contain %q after replacement\nGot:\n%s", want, content)
			}
		}
	})
}

// TestWriteMermaidToFile_NonKanbanFenceBeforeKanbanBlock validates that a file
// containing a non-kanban mermaid fence before the target kanban block is handled
// correctly — the state machine resets on the non-kanban fence closure.
func TestWriteMermaidToFile_NonKanbanFenceBeforeKanbanBlock(t *testing.T) {
	const fileContent = "# Doc\n\n```mermaid\ngraph TD\n  A --> B\n```\n\n```mermaid\nkanban\n  section To Do\n    TASK-OLD@{ label: \"Old\" }\n```\n"
	const newBlock = "```mermaid\nkanban\n  section To Do\n    TASK-001@{ label: \"New\" }\n```\n"

	dir := t.TempDir()
	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := writeMermaidToFile(path, newBlock)

	if err != nil {
		t.Errorf("writeMermaidToFile returned %v, want nil", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read file: %v", readErr)
	}
	if strings.Contains(string(got), "TASK-OLD") {
		t.Errorf("old task still present after replacement\nGot:\n%s", string(got))
	}
	if !strings.Contains(string(got), "TASK-001") {
		t.Errorf("new task missing after replacement\nGot:\n%s", string(got))
	}
}

func boardWithOneTask() domain.Board {
	return domain.Board{
		Columns: []domain.Column{
			{Name: "todo", Label: "To Do"},
		},
		Tasks: map[domain.TaskStatus][]domain.Task{
			"todo": {{ID: "TASK-001", Title: "Fix login bug", Status: "todo"}},
		},
	}
}

// TestSanitiseMermaidTitle validates that sanitiseMermaidTitle replaces
// Mermaid-unsafe characters with safe substitutes.
// Test Budget: 2 behaviors (quote replacement, bracket replacement) x 2 = 4 tests max.
// Parametrised input variations count as one behavior each.
func TestSanitiseMermaidTitle(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "double quotes replaced with single quotes",
			input: `Fix "login" bug`,
			want:  "Fix 'login' bug",
		},
		{
			name:  "square brackets replaced with parentheses",
			input: "[urgent] task",
			want:  "(urgent) task",
		},
		{
			name:  "combined unsafe chars all replaced",
			input: `Fix "login" bug [urgent]`,
			want:  "Fix 'login' bug (urgent)",
		},
		{
			name:  "newlines replaced with spaces",
			input: "line1\nline2\r",
			want:  "line1 line2 ",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitiseMermaidTitle(tc.input)
			if got != tc.want {
				t.Errorf("sanitiseMermaidTitle(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestRenderBoardMermaid_StartsWithFencedBlock validates that the output begins
// with a fenced mermaid code block (```mermaid).
func TestRenderBoardMermaid_StartsWithFencedBlock(t *testing.T) {
	board := boardWithOneTask()

	result := renderBoardMermaid(board)

	if !strings.HasPrefix(result, "```mermaid\n") {
		t.Errorf("expected output to start with \"```mermaid\\n\", got:\n%s", result)
	}
}

// TestRenderBoardMermaid_ContainsKanbanDiagramType validates that the output
// contains "kanban" as a standalone diagram type declaration line.
func TestRenderBoardMermaid_ContainsKanbanDiagramType(t *testing.T) {
	board := boardWithOneTask()

	result := renderBoardMermaid(board)

	found := false
	for _, line := range strings.Split(result, "\n") {
		if strings.TrimSpace(line) == "kanban" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output to contain \"kanban\" as diagram type declaration\nGot:\n%s", result)
	}
}
