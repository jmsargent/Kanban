package cli

import (
	"strings"
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
)

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
