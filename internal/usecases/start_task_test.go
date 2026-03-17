package usecases_test

import (
	"errors"
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fakes ───────────────────────────────────────────────────────────────────

// startTaskFakeRepo is a TaskRepository fake that supports configurable FindByID
// and records Update calls for assertion.
type startTaskFakeRepo struct {
	byID      map[string]domain.Task
	findErr   error
	updated   *domain.Task
	updateErr error
}

func newStartTaskFakeRepo(tasks ...domain.Task) *startTaskFakeRepo {
	repo := &startTaskFakeRepo{byID: make(map[string]domain.Task)}
	for _, t := range tasks {
		repo.byID[t.ID] = t
	}
	return repo
}

func (f *startTaskFakeRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	if f.findErr != nil {
		return domain.Task{}, f.findErr
	}
	t, ok := f.byID[taskID]
	if !ok {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	return t, nil
}

func (f *startTaskFakeRepo) Update(repoRoot string, task domain.Task) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = &task
	return nil
}

func (f *startTaskFakeRepo) Save(repoRoot string, task domain.Task) error  { return nil }
func (f *startTaskFakeRepo) ListAll(repoRoot string) ([]domain.Task, error) { return nil, nil }
func (f *startTaskFakeRepo) Delete(repoRoot, taskID string) error           { return nil }
func (f *startTaskFakeRepo) NextID(repoRoot string) (string, error)         { return "", nil }

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 5 behaviors x 2 = 10 max unit tests (using 5)

func TestStartTask_TransitionsTodoToInProgress_WhenTaskIsInTodo(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo(task)

	uc := usecases.NewStartTask(cfg, tasks)
	result, err := uc.Execute(repoRoot, "TASK-001", "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result.Transitioned {
		t.Error("expected result.Transitioned to be true")
	}
	if result.AlreadyInProgress {
		t.Error("expected result.AlreadyInProgress to be false")
	}
	if result.Task.Status != domain.StatusInProgress {
		t.Errorf("expected task status in-progress, got: %s", result.Task.Status)
	}
	if tasks.updated == nil {
		t.Error("expected task to be persisted via Update")
	}
	if tasks.updated != nil && tasks.updated.Status != domain.StatusInProgress {
		t.Errorf("expected persisted status in-progress, got: %s", tasks.updated.Status)
	}
}

func TestStartTask_ReturnsAlreadyInProgress_WhenTaskIsAlreadyInProgress(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-002", Title: "Running task", Status: domain.StatusInProgress}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "in-progress", Label: "In Progress"}},
	}}
	tasks := newStartTaskFakeRepo(task)

	uc := usecases.NewStartTask(cfg, tasks)
	result, err := uc.Execute(repoRoot, "TASK-002", "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result.AlreadyInProgress {
		t.Error("expected result.AlreadyInProgress to be true")
	}
	if result.Transitioned {
		t.Error("expected result.Transitioned to be false")
	}
	if tasks.updated != nil {
		t.Error("expected no Update call when task is already in-progress")
	}
}

func TestStartTask_ReturnsErrInvalidTransition_WhenTaskIsDone(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-003", Title: "Completed task", Status: domain.StatusDone}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "done", Label: "Done"}},
	}}
	tasks := newStartTaskFakeRepo(task)

	uc := usecases.NewStartTask(cfg, tasks)
	_, err := uc.Execute(repoRoot, "TASK-003", "")

	if !errors.Is(err, ports.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no Update call when transition is invalid")
	}
}

func TestStartTask_ReturnsErrTaskNotFound_WhenTaskIDDoesNotExist(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo() // empty — no tasks

	uc := usecases.NewStartTask(cfg, tasks)
	_, err := uc.Execute(repoRoot, "TASK-999", "")

	if !errors.Is(err, ports.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestStartTask_ReturnsErrNotInitialised_WhenRepoNotInitialised(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := newFreshConfigRepo() // ErrNotInitialised
	tasks := newStartTaskFakeRepo()

	uc := usecases.NewStartTask(cfg, tasks)
	_, err := uc.Execute(repoRoot, "TASK-001", "")

	if !errors.Is(err, ports.ErrNotInitialised) {
		t.Errorf("expected ErrNotInitialised, got: %v", err)
	}
}

func TestStartTask_SetsAssignee_WhenTaskIsStarted(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo(task)

	uc := usecases.NewStartTask(cfg, tasks)
	result, err := uc.Execute(repoRoot, "TASK-001", "Alice")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Task.Assignee != "Alice" {
		t.Errorf("expected result.Task.Assignee to be 'Alice', got: %q", result.Task.Assignee)
	}
	if tasks.updated == nil || tasks.updated.Assignee != "Alice" {
		t.Error("expected persisted task to have Assignee = 'Alice'")
	}
}

func TestStartTask_PreviousAssigneeIsEmpty_WhenTaskWasUnassigned(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo, Assignee: ""}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo(task)

	uc := usecases.NewStartTask(cfg, tasks)
	result, err := uc.Execute(repoRoot, "TASK-001", "Bob")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.PreviousAssignee != "" {
		t.Errorf("expected PreviousAssignee to be empty, got: %q", result.PreviousAssignee)
	}
}
