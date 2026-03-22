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

// Test Budget: 6 behaviors x 2 = 12 max unit tests (using 7)

func TestStartTask_PersistsInProgressStatus_WhenTaskIsInTodo(t *testing.T) {
	// Behavior: StartTask calls tasks.Update() with status:in-progress, making
	// YAML the authoritative state source.
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo(task)

	uc := usecases.NewStartTask(cfg, tasks)
	result, err := uc.Execute(repoRoot, "TASK-001", "alice@example.com")

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
	// tasks.Update() must be called with status:in-progress to persist the change.
	if tasks.updated == nil {
		t.Fatal("expected TaskRepository.Update to be called, but it was not")
	}
	if tasks.updated.Status != domain.StatusInProgress {
		t.Errorf("expected Update called with status in-progress, got: %s", tasks.updated.Status)
	}
	if tasks.updated.Assignee != "alice@example.com" {
		t.Errorf("expected Update called with assignee alice@example.com, got: %q", tasks.updated.Assignee)
	}
}

func TestStartTask_ReturnsAlreadyInProgress_WhenTaskStatusIsInProgress(t *testing.T) {
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
	// No update must be issued for an already in-progress task.
	if tasks.updated != nil {
		t.Error("expected no TaskRepository.Update call when already in-progress")
	}
}

func TestStartTask_ReturnsError_ForInvalidPreconditions(t *testing.T) {
	// These three error paths share identical assertion shape: Execute returns a
	// sentinel error and no Update is issued. Parametrized as input variations of
	// one behavior (Mandate 5).
	cases := []struct {
		name    string
		cfg     ports.ConfigRepository
		task    *domain.Task // nil = empty repo (ErrTaskNotFound)
		taskID  string
		wantErr error
	}{
		{
			name:    "done task returns ErrInvalidTransition",
			cfg:     &fakeConfigRepo{readResult: ports.Config{Columns: []domain.Column{{Name: "done", Label: "Done"}}}},
			task:    &domain.Task{ID: "TASK-003", Title: "Completed task", Status: domain.StatusDone},
			taskID:  "TASK-003",
			wantErr: ports.ErrInvalidTransition,
		},
		{
			name:    "missing task returns ErrTaskNotFound",
			cfg:     &fakeConfigRepo{readResult: ports.Config{Columns: []domain.Column{{Name: "todo", Label: "TODO"}}}},
			task:    nil,
			taskID:  "TASK-999",
			wantErr: ports.ErrTaskNotFound,
		},
		{
			name:    "uninitialised repo returns ErrNotInitialised",
			cfg:     newFreshConfigRepo(),
			task:    nil,
			taskID:  "TASK-001",
			wantErr: ports.ErrNotInitialised,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := tmpRepo(t)
			var tasks *startTaskFakeRepo
			if tc.task != nil {
				tasks = newStartTaskFakeRepo(*tc.task)
			} else {
				tasks = newStartTaskFakeRepo()
			}

			uc := usecases.NewStartTask(tc.cfg, tasks)
			_, err := uc.Execute(repoRoot, tc.taskID, "")

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected %v, got: %v", tc.wantErr, err)
			}
			if tasks.updated != nil {
				t.Error("expected no Update call when precondition fails")
			}
		})
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

func TestStartTask_PreviousAssigneeIsPopulated_WhenTaskWasAssignedToDifferentDeveloper(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo, Assignee: "Alice"}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo(task)

	uc := usecases.NewStartTask(cfg, tasks)
	result, err := uc.Execute(repoRoot, "TASK-001", "Bob")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.PreviousAssignee != "Alice" {
		t.Errorf("expected PreviousAssignee to be 'Alice', got: %q", result.PreviousAssignee)
	}
}
