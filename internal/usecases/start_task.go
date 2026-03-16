package usecases

import (
	"errors"
	"fmt"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// StartTaskResult communicates the outcome of a StartTask.Execute call.
type StartTaskResult struct {
	Transitioned      bool
	AlreadyInProgress bool
	Task              domain.Task
}

// StartTask implements the use case: explicitly start a task (todo -> in-progress).
type StartTask struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
}

// NewStartTask constructs a StartTask use case.
func NewStartTask(config ports.ConfigRepository, tasks ports.TaskRepository) *StartTask {
	return &StartTask{config: config, tasks: tasks}
}

// Execute transitions the identified task from todo to in-progress.
// Returns (StartTaskResult{AlreadyInProgress: true}, nil) when the task is already in-progress.
// Returns a wrapped ErrInvalidTransition when the task is done.
// Returns ErrTaskNotFound when the task ID does not exist.
// Returns ErrNotInitialised when the repository has not been initialised.
func (u *StartTask) Execute(repoRoot, taskID string) (StartTaskResult, error) {
	if _, err := u.config.Read(repoRoot); err != nil {
		return StartTaskResult{}, fmt.Errorf("read config: %w", err)
	}

	task, err := u.tasks.FindByID(repoRoot, taskID)
	if err != nil {
		if errors.Is(err, ports.ErrTaskNotFound) {
			return StartTaskResult{}, err
		}
		return StartTaskResult{}, fmt.Errorf("find task: %w", err)
	}

	if task.Status == domain.StatusInProgress {
		return StartTaskResult{AlreadyInProgress: true}, nil
	}

	if !domain.CanTransitionTo(task.Status, domain.StatusInProgress) {
		return StartTaskResult{}, fmt.Errorf("task %s: %w", taskID, ports.ErrInvalidTransition)
	}

	task.Status = domain.StatusInProgress
	if err := u.tasks.Update(repoRoot, task); err != nil {
		return StartTaskResult{}, fmt.Errorf("update task: %w", err)
	}

	return StartTaskResult{Transitioned: true, Task: task}, nil
}
