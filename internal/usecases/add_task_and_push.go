package usecases

import (
	"fmt"
	"path/filepath"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// TaskExecutor is the driving-port interface shared by AddTask and
// AddTaskAndPush. Primary adapters (web handler, CLI) depend on this
// abstraction rather than concrete use-case types.
type TaskExecutor interface {
	Execute(repoRoot string, input AddTaskInput) (domain.Task, error)
}

// commitGitPort is the subset of RemoteGitPort required for committing and
// pushing a new task file. Defined here so AddTaskAndPush only depends on the
// operations it actually uses.
type commitGitPort interface {
	Add(repoDir, path string) error
	Commit(repoDir, message string) error
	Push(repoDir string) error
}

// AddTaskAndPush orchestrates creating a task file and pushing it to the
// configured remote: AddTask → git add → git commit → git push.
type AddTaskAndPush struct {
	addTask *AddTask
	git     commitGitPort
}

// NewAddTaskAndPush constructs an AddTaskAndPush use case.
func NewAddTaskAndPush(config ports.ConfigRepository, tasks ports.TaskRepository, git commitGitPort) *AddTaskAndPush {
	return &AddTaskAndPush{
		addTask: NewAddTask(config, tasks),
		git:     git,
	}
}

// Execute creates the task, stages its file, commits, and pushes to origin.
// Returns the created task on success.
func (u *AddTaskAndPush) Execute(repoRoot string, input AddTaskInput) (domain.Task, error) {
	task, err := u.addTask.Execute(repoRoot, input)
	if err != nil {
		return domain.Task{}, err
	}

	taskPath := filepath.Join(repoRoot, ".kanban", "tasks", task.ID+".md")
	if err := u.git.Add(repoRoot, taskPath); err != nil {
		return domain.Task{}, fmt.Errorf("git add task file: %w", err)
	}

	message := fmt.Sprintf("feat: add task %s", task.ID)
	if err := u.git.Commit(repoRoot, message); err != nil {
		return domain.Task{}, fmt.Errorf("git commit task: %w", err)
	}

	if err := u.git.Push(repoRoot); err != nil {
		return domain.Task{}, fmt.Errorf("git push: %w", err)
	}

	return task, nil
}
