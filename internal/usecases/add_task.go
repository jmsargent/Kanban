package usecases

import (
	"errors"
	"fmt"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// AddTaskInput holds the caller-supplied fields for creating a new task.
type AddTaskInput struct {
	Title       string
	Priority    string
	Due         *time.Time
	Assignee    string
	Description string
	CreatedBy   string
}

// AddTask implements the create-task use case.
// Driving port entrypoint for "kanban add".
type AddTask struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
}

// NewAddTask constructs an AddTask use case with its required collaborators.
func NewAddTask(config ports.ConfigRepository, tasks ports.TaskRepository) *AddTask {
	return &AddTask{config: config, tasks: tasks}
}

// Execute validates input, generates an ID, persists the task, and returns it.
// Returns ErrInvalidInput when title is empty or due date is in the past.
// Returns ErrNotInitialised when the workspace has not been set up.
func (u *AddTask) Execute(repoRoot string, input AddTaskInput) (domain.Task, error) {
	if err := domain.ValidateNewTask(input.Title, input.Due); err != nil {
		return domain.Task{}, fmt.Errorf("%w: %s", ports.ErrInvalidInput, err.Error())
	}

	if _, err := u.config.Read(repoRoot); err != nil {
		if errors.Is(err, ports.ErrNotInitialised) {
			return domain.Task{}, ports.ErrNotInitialised
		}
		return domain.Task{}, fmt.Errorf("read config: %w", err)
	}

	id, err := u.tasks.NextID(repoRoot)
	if err != nil {
		return domain.Task{}, fmt.Errorf("next id: %w", err)
	}

	task := domain.Task{
		ID:          id,
		Title:       input.Title,
		Status:      domain.StatusTodo,
		Priority:    input.Priority,
		Due:         input.Due,
		Assignee:    input.Assignee,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
	}

	if err := u.tasks.Save(repoRoot, task); err != nil {
		return domain.Task{}, fmt.Errorf("save task: %w", err)
	}

	return task, nil
}
