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

// startTaskFakeLog is a TransitionLogRepository fake that records Append calls
// and returns a configurable LatestStatus for a task.
type startTaskFakeLog struct {
	latestByID    map[string]domain.TaskStatus
	latestErr     error
	appended      []domain.TransitionEntry
	appendErr     error
}

func newStartTaskFakeLog() *startTaskFakeLog {
	return &startTaskFakeLog{latestByID: make(map[string]domain.TaskStatus)}
}

func (f *startTaskFakeLog) withLatestStatus(taskID string, status domain.TaskStatus) *startTaskFakeLog {
	f.latestByID[taskID] = status
	return f
}

func (f *startTaskFakeLog) Append(repoRoot string, entry domain.TransitionEntry) error {
	if f.appendErr != nil {
		return f.appendErr
	}
	f.appended = append(f.appended, entry)
	return nil
}

func (f *startTaskFakeLog) LatestStatus(repoRoot, taskID string) (domain.TaskStatus, error) {
	if f.latestErr != nil {
		return "", f.latestErr
	}
	status, ok := f.latestByID[taskID]
	if !ok {
		return "", ports.ErrTaskNotFound
	}
	return status, nil
}

func (f *startTaskFakeLog) History(repoRoot, taskID string) ([]domain.TransitionEntry, error) {
	return nil, nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 6 behaviors x 2 = 12 max unit tests (using 8)

func TestStartTask_AppendsTransitionEntry_WhenTaskIsInTodo(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo(task)
	log := newStartTaskFakeLog() // no latest status => ErrTaskNotFound => treat as todo

	uc := usecases.NewStartTask(cfg, tasks, log)
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
	if len(log.appended) != 1 {
		t.Fatalf("expected 1 Append call, got %d", len(log.appended))
	}
	entry := log.appended[0]
	if entry.TaskID != "TASK-001" {
		t.Errorf("expected entry.TaskID 'TASK-001', got %q", entry.TaskID)
	}
	if entry.From != domain.StatusTodo {
		t.Errorf("expected entry.From todo, got %q", entry.From)
	}
	if entry.To != domain.StatusInProgress {
		t.Errorf("expected entry.To in-progress, got %q", entry.To)
	}
	if entry.Author != "alice@example.com" {
		t.Errorf("expected entry.Author 'alice@example.com', got %q", entry.Author)
	}
	if entry.Trigger != "manual" {
		t.Errorf("expected entry.Trigger 'manual', got %q", entry.Trigger)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected entry.Timestamp to be non-zero")
	}
}

func TestStartTask_NeverCallsTaskUpdate_OnTransition(t *testing.T) {
	// Task file is never modified by StartTask — status and author go into
	// transitions.log only. Update must never be called.
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newStartTaskFakeRepo(task)
	log := newStartTaskFakeLog()

	uc := usecases.NewStartTask(cfg, tasks, log)
	_, err := uc.Execute(repoRoot, "TASK-001", "alice@example.com")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no TaskRepository.Update call: task file must not be modified by start")
	}
}

func TestStartTask_ReturnsAlreadyInProgress_WhenLatestStatusIsInProgress(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-002", Title: "Running task", Status: domain.StatusTodo}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "in-progress", Label: "In Progress"}},
	}}
	tasks := newStartTaskFakeRepo(task)
	log := newStartTaskFakeLog().withLatestStatus("TASK-002", domain.StatusInProgress)

	uc := usecases.NewStartTask(cfg, tasks, log)
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
	if len(log.appended) != 0 {
		t.Errorf("expected no Append call when task is already in-progress, got %d", len(log.appended))
	}
}

func TestStartTask_ReturnsError_ForInvalidPreconditions(t *testing.T) {
	// These three error paths share identical assertion shape: Execute returns a
	// sentinel error and no Append is issued. Parametrized as input variations of
	// one behavior (Mandate 5).
	cases := []struct {
		name      string
		cfg       ports.ConfigRepository
		task      *domain.Task // nil = empty repo (ErrTaskNotFound)
		logStatus domain.TaskStatus
		taskID    string
		wantErr   error
	}{
		{
			name:      "done task returns ErrInvalidTransition",
			cfg:       &fakeConfigRepo{readResult: ports.Config{Columns: []domain.Column{{Name: "done", Label: "Done"}}}},
			task:      &domain.Task{ID: "TASK-003", Title: "Completed task", Status: domain.StatusDone},
			logStatus: domain.StatusDone,
			taskID:    "TASK-003",
			wantErr:   ports.ErrInvalidTransition,
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
			log := newStartTaskFakeLog()
			if tc.logStatus != "" {
				log.withLatestStatus(tc.taskID, tc.logStatus)
			}

			uc := usecases.NewStartTask(tc.cfg, tasks, log)
			_, err := uc.Execute(repoRoot, tc.taskID, "")

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected %v, got: %v", tc.wantErr, err)
			}
			if len(log.appended) != 0 {
				t.Error("expected no Append call when precondition fails")
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
	log := newStartTaskFakeLog()

	uc := usecases.NewStartTask(cfg, tasks, log)
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
	log := newStartTaskFakeLog()

	uc := usecases.NewStartTask(cfg, tasks, log)
	result, err := uc.Execute(repoRoot, "TASK-001", "Bob")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.PreviousAssignee != "Alice" {
		t.Errorf("expected PreviousAssignee to be 'Alice', got: %q", result.PreviousAssignee)
	}
	// result.Task reflects the task as read from the repository (not updated to YAML).
	// Assignee tracking now goes into TransitionEntry.Author only.
}

