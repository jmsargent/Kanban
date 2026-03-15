package usecases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fakes ───────────────────────────────────────────────────────────────────

type fakeTaskRepo struct {
	nextID  string
	saved   *domain.Task
	saveErr error
	listAll []domain.Task
}

func (f *fakeTaskRepo) NextID(repoRoot string) (string, error) {
	return f.nextID, nil
}

func (f *fakeTaskRepo) Save(repoRoot string, task domain.Task) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = &task
	return nil
}

func (f *fakeTaskRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	return domain.Task{}, ports.ErrTaskNotFound
}

func (f *fakeTaskRepo) ListAll(repoRoot string) ([]domain.Task, error) {
	return f.listAll, nil
}

func (f *fakeTaskRepo) Update(repoRoot string, task domain.Task) error {
	f.saved = &task
	return nil
}

func (f *fakeTaskRepo) Delete(repoRoot, taskID string) error {
	return nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 4 behaviors x 2 = 8 max unit tests (using 4)

func TestAddTask_CreatesTaskWithTodoStatus_WhenTitleIsValid(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := &fakeTaskRepo{nextID: "TASK-001"}
	input := usecases.AddTaskInput{Title: "Fix login bug"}

	uc := usecases.NewAddTask(cfg, tasks)
	task, err := uc.Execute(repoRoot, input)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if task.ID != "TASK-001" {
		t.Errorf("expected task ID TASK-001, got: %s", task.ID)
	}
	if task.Title != "Fix login bug" {
		t.Errorf("expected title 'Fix login bug', got: %s", task.Title)
	}
	if task.Status != domain.StatusTodo {
		t.Errorf("expected status todo, got: %s", task.Status)
	}
	if tasks.saved == nil {
		t.Error("expected task to be saved via TaskRepository")
	}
}

func TestAddTask_ReturnsErrInvalidInput_WhenTitleIsEmpty(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := &fakeTaskRepo{nextID: "TASK-001"}
	input := usecases.AddTaskInput{Title: ""}

	uc := usecases.NewAddTask(cfg, tasks)
	_, err := uc.Execute(repoRoot, input)

	if !errors.Is(err, ports.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
	if tasks.saved != nil {
		t.Error("expected no task to be saved when title is empty")
	}
}

func TestAddTask_PersistsOptionalFields_WhenProvided(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := &fakeTaskRepo{nextID: "TASK-001"}
	due := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
	input := usecases.AddTaskInput{
		Title:    "API work",
		Priority: "P1",
		Due:      &due,
		Assignee: "Alice",
	}

	uc := usecases.NewAddTask(cfg, tasks)
	task, err := uc.Execute(repoRoot, input)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if task.Priority != "P1" {
		t.Errorf("expected priority P1, got: %s", task.Priority)
	}
	if task.Assignee != "Alice" {
		t.Errorf("expected assignee Alice, got: %s", task.Assignee)
	}
}

func TestAddTask_ReturnsErrNotInitialised_WhenKanbanNotSetUp(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := newFreshConfigRepo() // ErrNotInitialised
	tasks := &fakeTaskRepo{nextID: "TASK-001"}
	input := usecases.AddTaskInput{Title: "valid title"}

	uc := usecases.NewAddTask(cfg, tasks)
	_, err := uc.Execute(repoRoot, input)

	if !errors.Is(err, ports.ErrNotInitialised) {
		t.Errorf("expected ErrNotInitialised, got: %v", err)
	}
}
