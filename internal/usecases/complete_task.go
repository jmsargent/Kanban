package usecases

import (
	"errors"
	"fmt"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// CompleteTaskResult communicates the outcome of a CompleteTask.Execute call.
type CompleteTaskResult struct {
	From      domain.TaskStatus
	To        domain.TaskStatus
	AlreadyDone bool
}

// CompleteTask implements the use case: explicitly move a task to done.
// It reads the task status from the task file and writes the updated status
// back via TaskRepository.Update(). No git interactions are performed (C-03).
type CompleteTask struct {
	tasks ports.TaskRepository
}

// NewCompleteTask constructs a CompleteTask use case.
func NewCompleteTask(tasks ports.TaskRepository) *CompleteTask {
	return &CompleteTask{tasks: tasks}
}

// Execute transitions the identified task to done.
// Returns CompleteTaskResult{AlreadyDone: true} when task is already done (idempotent).
// Returns ErrTaskNotFound when the task ID does not exist.
func (u *CompleteTask) Execute(repoRoot, taskID string) (CompleteTaskResult, error) {
	task, err := u.tasks.FindByID(repoRoot, taskID)
	if err != nil {
		if errors.Is(err, ports.ErrTaskNotFound) {
			return CompleteTaskResult{}, err
		}
		return CompleteTaskResult{}, fmt.Errorf("find task: %w", err)
	}

	if task.Status == domain.StatusDone {
		return CompleteTaskResult{AlreadyDone: true}, nil
	}

	from := task.Status
	task.Status = domain.StatusDone

	if err := u.tasks.Update(repoRoot, task); err != nil {
		return CompleteTaskResult{}, fmt.Errorf("update task: %w", err)
	}

	return CompleteTaskResult{From: from, To: domain.StatusDone}, nil
}
