package usecases_test

import (
	"errors"
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fake ────────────────────────────────────────────────────────────────────

// completeTaskFakeRepo is an in-memory TaskRepository for CompleteTask tests.
type completeTaskFakeRepo struct {
	byID        map[string]domain.Task
	updated     *domain.Task
	updateCalls int
}

func newCompleteTaskFakeRepo(tasks ...domain.Task) *completeTaskFakeRepo {
	r := &completeTaskFakeRepo{byID: make(map[string]domain.Task)}
	for _, t := range tasks {
		r.byID[t.ID] = t
	}
	return r
}

func (r *completeTaskFakeRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	t, ok := r.byID[taskID]
	if !ok {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	return t, nil
}

func (r *completeTaskFakeRepo) Update(repoRoot string, task domain.Task) error {
	r.updateCalls++
	r.updated = &task
	r.byID[task.ID] = task
	return nil
}

func (r *completeTaskFakeRepo) Save(repoRoot string, task domain.Task) error  { return nil }
func (r *completeTaskFakeRepo) ListAll(repoRoot string) ([]domain.Task, error) { return nil, nil }
func (r *completeTaskFakeRepo) Delete(repoRoot, taskID string) error           { return nil }
func (r *completeTaskFakeRepo) NextID(repoRoot string) (string, error)         { return "", nil }

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 4 behaviors x 2 = 8 max unit tests (using 4)
// Behaviors:
//  1. In-progress task transitions to done: Update called, From=in-progress, To=done
//  2. Todo task transitions to done: Update called, From=todo, To=done
//  3. Already-done task: AlreadyDone=true, no Update call
//  4. Non-existent task: ErrTaskNotFound returned

func TestCompleteTask_TransitionsToDone(t *testing.T) {
	cases := []struct {
		name       string
		fromStatus domain.TaskStatus
	}{
		{"in-progress task transitions to done", domain.StatusInProgress},
		{"todo task transitions to done", domain.StatusTodo},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := tmpRepo(t)
			task := domain.Task{ID: "TASK-001", Title: "Fix bug", Status: tc.fromStatus}
			repo := newCompleteTaskFakeRepo(task)

			uc := usecases.NewCompleteTask(repo)
			result, err := uc.Execute(repoRoot, "TASK-001")

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if result.AlreadyDone {
				t.Error("expected AlreadyDone to be false")
			}
			if result.From != tc.fromStatus {
				t.Errorf("expected From=%s, got: %s", tc.fromStatus, result.From)
			}
			if result.To != domain.StatusDone {
				t.Errorf("expected To=done, got: %s", result.To)
			}
			if repo.updateCalls != 1 {
				t.Errorf("expected exactly 1 Update call, got: %d", repo.updateCalls)
			}
			if repo.updated == nil || repo.updated.Status != domain.StatusDone {
				t.Error("expected Update to be called with task status=done")
			}
		})
	}
}

func TestCompleteTask_ReturnsAlreadyDone_WhenTaskIsAlreadyDone(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-002", Title: "Deployed", Status: domain.StatusDone}
	repo := newCompleteTaskFakeRepo(task)

	uc := usecases.NewCompleteTask(repo)
	result, err := uc.Execute(repoRoot, "TASK-002")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result.AlreadyDone {
		t.Error("expected AlreadyDone to be true")
	}
	if repo.updateCalls != 0 {
		t.Errorf("expected no Update call for already-done task, got: %d", repo.updateCalls)
	}
}

func TestCompleteTask_ReturnsErrTaskNotFound_WhenTaskDoesNotExist(t *testing.T) {
	repoRoot := tmpRepo(t)
	repo := newCompleteTaskFakeRepo() // empty repo

	uc := usecases.NewCompleteTask(repo)
	_, err := uc.Execute(repoRoot, "TASK-999")

	if !errors.Is(err, ports.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}
