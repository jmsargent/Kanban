package usecases_test

import (
	"testing"
	"time"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 5 behaviors x 2 = 10 max unit tests (using 5)

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

func TestGetBoard_SortsTasksByCreatedAtAscending_WhenMultipleTasksInSameColumn(t *testing.T) {
	// Behavior: tasks within a column are ordered oldest-first by CreatedAt.
	repoRoot := tmpRepo(t)
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
		},
	}}
	tasks := &fakeTaskRepo{listAll: []domain.Task{
		{ID: "TASK-002", Title: "Newer Task", Status: domain.StatusTodo, CreatedAt: newer},
		{ID: "TASK-001", Title: "Older Task", Status: domain.StatusTodo, CreatedAt: older},
	}}

	uc := usecases.NewGetBoard(cfg, tasks)
	board, err := uc.Execute(repoRoot, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	todoTasks := board.Tasks[domain.StatusTodo]
	if len(todoTasks) != 2 {
		t.Fatalf("expected 2 tasks in TODO, got: %d", len(todoTasks))
	}
	if todoTasks[0].ID != "TASK-001" {
		t.Errorf("expected oldest task first, got: %s", todoTasks[0].ID)
	}
	if todoTasks[1].ID != "TASK-002" {
		t.Errorf("expected newer task second, got: %s", todoTasks[1].ID)
	}
}
