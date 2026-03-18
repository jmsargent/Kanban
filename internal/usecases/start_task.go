package usecases

import (
	"errors"
	"fmt"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
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
	log    ports.TransitionLogRepository
}

// NewStartTask constructs a StartTask use case.
func NewStartTask(config ports.ConfigRepository, tasks ports.TaskRepository, log ports.TransitionLogRepository) *StartTask {
	return &StartTask{config: config, tasks: tasks, log: log}
}

// Execute transitions the identified task from todo to in-progress.
// assignee is the git author email recorded in the TransitionEntry.Author field.
// The task file is never modified — status and author are tracked in transitions.log only.
// Returns (StartTaskResult{AlreadyInProgress: true}, nil) when the task is already in-progress per LatestStatus.
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

	// Derive current status from transitions log, not from task YAML.
	currentStatus, err := u.log.LatestStatus(repoRoot, taskID)
	if err != nil && !errors.Is(err, ports.ErrTaskNotFound) {
		return StartTaskResult{}, fmt.Errorf("latest status: %w", err)
	}
	// ErrTaskNotFound from LatestStatus means no log entries exist => task is todo.
	if errors.Is(err, ports.ErrTaskNotFound) {
		currentStatus = domain.StatusTodo
	}

	if currentStatus == domain.StatusInProgress {
		return StartTaskResult{AlreadyInProgress: true}, nil
	}

	if !domain.CanTransitionTo(currentStatus, domain.StatusInProgress) {
		return StartTaskResult{}, fmt.Errorf("task %s: %w", taskID, ports.ErrInvalidTransition)
	}

	// Preserve PreviousAssignee from whatever name was stored in YAML (may be empty
	// for tasks created after the log-based migration).
	previousAssignee := task.Assignee

	entry := domain.TransitionEntry{
		Timestamp: time.Now().UTC(),
		TaskID:    taskID,
		From:      currentStatus,
		To:        domain.StatusInProgress,
		Author:    assignee,
		Trigger:   "manual",
	}
	if err := u.log.Append(repoRoot, entry); err != nil {
		return StartTaskResult{}, fmt.Errorf("append transition: %w", err)
	}

	// Task file is NOT modified: status and author are tracked in transitions.log only.
	task.Status = domain.StatusInProgress
	result := StartTaskResult{Transitioned: true, Task: task}
	if previousAssignee != "" && previousAssignee != assignee {
		result.PreviousAssignee = previousAssignee
	}
	return result, nil
}
