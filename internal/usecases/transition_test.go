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

// Test Budget: 5 behaviors x 2 = 10 max unit tests (using 6)

func TestTransitionToInProgress_DoesNotModifyTaskFile_WhenCommitReferencesTask(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: domain.StatusTodo}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(nil), out)
	err := uc.Execute(repoRoot, "TASK-001: start work")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// The hook must NOT modify the task file — status is tracked in transitions.log only.
	if tasks.updated != nil {
		t.Fatal("hook must not update the task file; status is tracked in transitions.log only")
	}
}

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
	// Task file must NOT be updated — status is tracked in transitions.log only.
	if tasks.updated != nil {
		t.Error("hook must not update the task file")
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
	task := domain.Task{ID: "TASK-002", Title: "API work"}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	// Log reports task already in-progress (authoritative source).
	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(map[string]domain.TaskStatus{"TASK-002": domain.StatusInProgress}), out)
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
	task := domain.Task{ID: "TASK-003", Title: "Finished work"}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	// Log reports task already done (authoritative source).
	uc := usecases.NewTransitionToInProgress(cfg, tasks, newFakeTransitionLog(map[string]domain.TaskStatus{"TASK-003": domain.StatusDone}), out)
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

func TestTransitionToInProgress_ReturnsError_WhenLogAppendFails(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Broken task"}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-[0-9]+`}}
	out := &strings.Builder{}

	// Execute() propagates append errors so the hook adapter can write a stderr warning.
	// The hook's RunE then writes the warning and returns nil (exit 0).
	uc := usecases.NewTransitionToInProgress(cfg, tasks, &errorOnAppendLog{}, out)
	err := uc.Execute(repoRoot, "TASK-001: some work")
	if err == nil {
		t.Error("expected error to be propagated when log.Append fails")
	}
}

// errorOnAppendLog is a fake TransitionLogRepository whose Append always fails.
type errorOnAppendLog struct{}

func (f *errorOnAppendLog) Append(_ string, _ domain.TransitionEntry) error {
	return errors.New("log unwritable")
}

func (f *errorOnAppendLog) LatestStatus(_, _ string) (domain.TaskStatus, error) {
	return domain.StatusTodo, nil // default todo so transition is attempted
}

func (f *errorOnAppendLog) History(_, _ string) ([]domain.TransitionEntry, error) {
	return nil, nil
}
