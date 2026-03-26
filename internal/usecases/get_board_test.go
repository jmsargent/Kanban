package usecases_test

import (
	"testing"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 4 behaviors x 2 = 8 max unit tests (using 4)

func TestGetBoard_ReturnsUnassignedCount_WhenFilterActiveAndUnassignedTasksExist(t *testing.T) {
	// Behavior: when a filterAssignee is set, tasks with no assignee are excluded
	// from the board but their count is surfaced in Board.UnassignedCount so the
	// CLI adapter can warn the developer.
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
			{Name: "in-progress", Label: "IN PROGRESS"},
			{Name: "done", Label: "DONE"},
		},
	}}
	tasks := &fakeTaskRepo{listAll: []domain.Task{
		{ID: "TASK-001", Title: "My task", Assignee: "dev@example.com", Status: domain.StatusTodo},
		{ID: "TASK-002", Title: "Unassigned task", Assignee: "", Status: domain.StatusTodo},
	}}

	uc := usecases.NewGetBoard(cfg, tasks)
	board, err := uc.Execute(repoRoot, "dev@example.com")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if board.UnassignedCount != 1 {
		t.Errorf("expected UnassignedCount=1, got: %d", board.UnassignedCount)
	}
	// Unassigned task must not appear in board columns.
	if len(board.Tasks[domain.StatusTodo]) != 1 {
		t.Errorf("expected 1 task in TODO (my task only), got: %d", len(board.Tasks[domain.StatusTodo]))
	}
}

func TestGetBoard_GroupsTasksByStatus_WhenTasksExist(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
			{Name: "in-progress", Label: "IN PROGRESS"},
			{Name: "done", Label: "DONE"},
		},
	}}
	tasks := &fakeTaskRepo{listAll: []domain.Task{
		{ID: "TASK-001", Title: "Todo work", Status: domain.StatusTodo},
		{ID: "TASK-002", Title: "Active work", Status: domain.StatusInProgress},
		{ID: "TASK-003", Title: "Finished work", Status: domain.StatusDone},
	}}

	uc := usecases.NewGetBoard(cfg, tasks)
	board, err := uc.Execute(repoRoot, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(board.Columns) != 3 {
		t.Errorf("expected 3 columns, got: %d", len(board.Columns))
	}
	if len(board.Tasks[domain.StatusTodo]) != 1 {
		t.Errorf("expected 1 todo task, got: %d", len(board.Tasks[domain.StatusTodo]))
	}
	if len(board.Tasks[domain.StatusInProgress]) != 1 {
		t.Errorf("expected 1 in-progress task, got: %d", len(board.Tasks[domain.StatusInProgress]))
	}
	if len(board.Tasks[domain.StatusDone]) != 1 {
		t.Errorf("expected 1 done task, got: %d", len(board.Tasks[domain.StatusDone]))
	}
}

func TestGetBoard_ReturnsEmptyBoard_WhenNoTasksExist(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
			{Name: "in-progress", Label: "IN PROGRESS"},
			{Name: "done", Label: "DONE"},
		},
	}}
	tasks := &fakeTaskRepo{listAll: []domain.Task{}}

	uc := usecases.NewGetBoard(cfg, tasks)
	board, err := uc.Execute(repoRoot, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(board.Columns) != 3 {
		t.Errorf("expected 3 columns even when empty, got: %d", len(board.Columns))
	}
	total := len(board.Tasks[domain.StatusTodo]) +
		len(board.Tasks[domain.StatusInProgress]) +
		len(board.Tasks[domain.StatusDone])
	if total != 0 {
		t.Errorf("expected 0 tasks on empty board, got: %d", total)
	}
}

func TestGetBoard_PlacesTaskInTodo_WhenStatusFieldIsEmpty(t *testing.T) {
	// Behavior: a task with an empty Status YAML field defaults to TODO.
	// GetBoard reads task.Status directly; empty string maps to StatusTodo.
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
			{Name: "in-progress", Label: "IN PROGRESS"},
			{Name: "done", Label: "DONE"},
		},
	}}
	// Task has no Status field in YAML (zero value).
	tasks := &fakeTaskRepo{listAll: []domain.Task{
		{ID: "TASK-001", Title: "New task"},
	}}

	uc := usecases.NewGetBoard(cfg, tasks)
	board, err := uc.Execute(repoRoot, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(board.Tasks[domain.StatusTodo]) != 1 {
		t.Errorf("expected 1 task in TODO column, got: %d", len(board.Tasks[domain.StatusTodo]))
	}
}
