package usecases_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fakes for TransitionToInProgress ────────────────────────────────────────

type transitionTaskRepo struct {
	tasks     map[string]domain.Task
	findErr   error
	updateErr error
	updated   *domain.Task
}

func newTransitionTaskRepo(tasks ...domain.Task) *transitionTaskRepo {
	r := &transitionTaskRepo{tasks: make(map[string]domain.Task)}
	for _, t := range tasks {
		r.tasks[t.ID] = t
	}
	return r
}

func (r *transitionTaskRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	if r.findErr != nil {
		return domain.Task{}, r.findErr
	}
	t, ok := r.tasks[taskID]
	if !ok {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	return t, nil
}

func (r *transitionTaskRepo) Save(repoRoot string, task domain.Task) error { return nil }

func (r *transitionTaskRepo) ListAll(repoRoot string) ([]domain.Task, error) {
	return nil, nil
}

func (r *transitionTaskRepo) Update(repoRoot string, task domain.Task) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updated = &task
	r.tasks[task.ID] = task
	return nil
}

func (r *transitionTaskRepo) Delete(repoRoot, taskID string) error { return nil }

func (r *transitionTaskRepo) NextID(repoRoot string) (string, error) { return "TASK-001", nil }

// ─── Tests ────────────────────────────────────────────────────────────────────

// Test Budget: 5 behaviors x 2 = 10 max unit tests (using 5)

func TestTransitionToInProgress_TransitionsTodoTask_WhenCommitReferencesIt(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusTodo}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	err := uc.Execute(repoRoot, "TASK-001: start OAuth work")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated == nil {
		t.Fatal("expected task to be updated via TaskRepository")
	}
	if tasks.updated.Status != domain.StatusInProgress {
		t.Errorf("expected status in-progress, got: %s", tasks.updated.Status)
	}
	if !strings.Contains(out.String(), "kanban: TASK-001") {
		t.Errorf("expected transition output to contain 'kanban: TASK-001', got: %q", out.String())
	}
	if !strings.Contains(out.String(), "in-progress") {
		t.Errorf("expected transition output to contain 'in-progress', got: %q", out.String())
	}
}

func TestTransitionToInProgress_SkipsTask_WhenAlreadyInProgress(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-002", Title: "API work", Status: domain.StatusInProgress}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	err := uc.Execute(repoRoot, "TASK-002: add throttle middleware")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no update for task already in-progress")
	}
	if strings.Contains(out.String(), "kanban:") {
		t.Errorf("expected no transition output for in-progress task, got: %q", out.String())
	}
}

func TestTransitionToInProgress_SkipsTask_WhenDone(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-003", Title: "Finished work", Status: domain.StatusDone}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	err := uc.Execute(repoRoot, "TASK-003: minor cleanup")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no update for task that is done")
	}
	if strings.Contains(out.String(), "kanban:") {
		t.Errorf("expected no transition output for done task, got: %q", out.String())
	}
}

func TestTransitionToInProgress_HandlesNotFound_WhenTaskMissing(t *testing.T) {
	repoRoot := tmpRepo(t)
	tasks := newTransitionTaskRepo() // empty — no tasks
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	err := uc.Execute(repoRoot, "TASK-099: some work")

	if err != nil {
		t.Fatalf("expected no error (missing task should not fail), got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no update for non-existent task")
	}
}

func TestTransitionToInProgress_ProducesNoOutput_WhenNoTaskMatchInMessage(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Silent task", Status: domain.StatusTodo}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	err := uc.Execute(repoRoot, "fix typo in README")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no update when message has no task reference")
	}
	if out.String() != "" {
		t.Errorf("expected no output when message has no task reference, got: %q", out.String())
	}
}

func TestTransitionToInProgress_IsNoOp_WhenNotInitialised(t *testing.T) {
	repoRoot := tmpRepo(t)
	tasks := newTransitionTaskRepo()
	cfg := newFreshConfigRepo() // ErrNotInitialised
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	err := uc.Execute(repoRoot, "TASK-001: some work")

	if err != nil {
		t.Fatalf("expected no error when kanban not initialised, got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no update when kanban not initialised")
	}
}

func TestTransitionToInProgress_ReturnsNilOnUnrecoverableError_WhenUpdateFails(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Broken task", Status: domain.StatusTodo}
	tasks := newTransitionTaskRepo(task)
	tasks.updateErr = errors.New("disk full")
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	// Errors must be swallowed (hook must never fail)
	_ = uc.Execute(repoRoot, "TASK-001: some work")
	// No assertion on error — hook always exits 0, errors go to hook.log
}
