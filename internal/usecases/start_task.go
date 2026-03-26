package usecases

import (
	"errors"
	"fmt"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// StartTaskResult communicates the outcome of a StartTask.Execute call.
type StartTaskResult struct {
	Transitioned      bool
	AlreadyInProgress bool
	Task              domain.Task
	PreviousAssignee  string // non-empty when task YAML had a different assignee before this start
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
// assignee is the git author email written to the task file's assignee field.
// Status is persisted to the task file via TaskRepository.Update() — YAML is
// the authoritative state source.
// Returns (StartTaskResult{AlreadyInProgress: true}, nil) when the task is already in-progress per task.Status.
// Returns a wrapped ErrInvalidTransition when the task is done.
// Returns ErrTaskNotFound when the task ID does not exist.
// Returns ErrNotInitialised when the repository has not been initialised.
func (u *StartTask) Execute(repoRoot, taskID, assignee string) (StartTaskResult, error) {
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

	currentStatus := task.Status
	if currentStatus == "" {
		currentStatus = domain.StatusTodo
	}

	if currentStatus == domain.StatusInProgress {
		return StartTaskResult{AlreadyInProgress: true}, nil
	}

	if !domain.CanTransitionTo(currentStatus, domain.StatusInProgress) {
		return StartTaskResult{}, fmt.Errorf("task %s: %w", taskID, ports.ErrInvalidTransition)
	}

	previousAssignee := task.Assignee

	task.Status = domain.StatusInProgress
	task.Assignee = assignee

	if err := u.tasks.Update(repoRoot, task); err != nil {
		return StartTaskResult{}, fmt.Errorf("update task: %w", err)
	}

	result := StartTaskResult{Transitioned: true, Task: task}
	if previousAssignee != "" && previousAssignee != assignee {
		result.PreviousAssignee = previousAssignee
	}
	return result, nil
}
