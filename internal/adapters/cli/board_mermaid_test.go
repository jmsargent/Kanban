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
