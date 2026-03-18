package usecases_test

import (
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fake TransitionLogRepository ────────────────────────────────────────────

// fakeTransitionLog is an in-memory fake implementing ports.TransitionLogRepository.
// It maps taskID -> TaskStatus to simulate log entries.
type fakeTransitionLog struct {
	latestStatus map[string]domain.TaskStatus
}

func newFakeTransitionLog(statuses map[string]domain.TaskStatus) *fakeTransitionLog {
	if statuses == nil {
		statuses = make(map[string]domain.TaskStatus)
	}
	return &fakeTransitionLog{latestStatus: statuses}
}

func (f *fakeTransitionLog) Append(repoRoot string, entry domain.TransitionEntry) error {
	return nil
}

func (f *fakeTransitionLog) LatestStatus(repoRoot, taskID string) (domain.TaskStatus, error) {
	if status, ok := f.latestStatus[taskID]; ok {
		return status, nil
	}
	return domain.StatusTodo, nil
}

func (f *fakeTransitionLog) History(repoRoot, taskID string) ([]domain.TransitionEntry, error) {
	return []domain.TransitionEntry{}, nil
}

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
		{ID: "TASK-001", Title: "My task", Assignee: "dev@example.com"},
		{ID: "TASK-002", Title: "Unassigned task", Assignee: ""},
	}}
	log := newFakeTransitionLog(map[string]domain.TaskStatus{
		"TASK-001": domain.StatusTodo,
		"TASK-002": domain.StatusTodo,
	})

	uc := usecases.NewGetBoard(cfg, tasks, log)
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
		{ID: "TASK-001", Title: "Todo work"},
		{ID: "TASK-002", Title: "Active work"},
		{ID: "TASK-003", Title: "Finished work"},
	}}
	log := newFakeTransitionLog(map[string]domain.TaskStatus{
		"TASK-001": domain.StatusTodo,
		"TASK-002": domain.StatusInProgress,
		"TASK-003": domain.StatusDone,
	})

	uc := usecases.NewGetBoard(cfg, tasks, log)
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
	log := newFakeTransitionLog(nil)

	uc := usecases.NewGetBoard(cfg, tasks, log)
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

func TestGetBoard_PlacesTaskInTodo_WhenNoLogEntryExists(t *testing.T) {
	// Behavior: a task with no log entry defaults to TODO regardless of the
	// YAML status field. This enforces that GetBoard uses the log, not YAML.
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
			{Name: "in-progress", Label: "IN PROGRESS"},
			{Name: "done", Label: "DONE"},
		},
	}}
	// Task has no Status field in YAML (zero value), and no log entry.
	tasks := &fakeTaskRepo{listAll: []domain.Task{
		{ID: "TASK-001", Title: "New task"},
	}}
	// Empty log — no entries for TASK-001 — LatestStatus returns StatusTodo.
	log := newFakeTransitionLog(nil)

	uc := usecases.NewGetBoard(cfg, tasks, log)
	board, err := uc.Execute(repoRoot, "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(board.Tasks[domain.StatusTodo]) != 1 {
		t.Errorf("expected 1 task in TODO column, got: %d", len(board.Tasks[domain.StatusTodo]))
	}
}
