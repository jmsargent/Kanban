package usecases

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/jmsargent/kanban/internal/ports"
)

// TaskHistoryResult is the output of GetTaskHistory.Execute.
type TaskHistoryResult struct {
	TaskID  string
	Title   string
	Entries []ports.CommitEntry
}

// GetTaskHistory retrieves the git commit history for a task file.
type GetTaskHistory struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
	git    ports.GitPort
}

// NewGetTaskHistory constructs a GetTaskHistory use case.
func NewGetTaskHistory(config ports.ConfigRepository, tasks ports.TaskRepository, git ports.GitPort) *GetTaskHistory {
	return &GetTaskHistory{config: config, tasks: tasks, git: git}
}

// Execute returns the task header and its git commit history.
// Returns ErrNotInitialised when the repository has not been initialised.
// Returns ErrTaskNotFound when no task matches taskID.
// Returns an empty Entries slice (not an error) when no commits exist.
func (u *GetTaskHistory) Execute(repoRoot, taskID string) (TaskHistoryResult, error) {
	if _, err := u.config.Read(repoRoot); err != nil {
		return TaskHistoryResult{}, fmt.Errorf("read config: %w", err)
	}

	task, err := u.tasks.FindByID(repoRoot, taskID)
	if err != nil {
		if errors.Is(err, ports.ErrTaskNotFound) {
			return TaskHistoryResult{}, err
		}
		return TaskHistoryResult{}, fmt.Errorf("find task: %w", err)
	}

	filePath := filepath.Join(".kanban", "tasks", taskID+".md")
	entries, err := u.git.LogFile(repoRoot, filePath)
	if err != nil {
		return TaskHistoryResult{}, fmt.Errorf("git log file: %w", err)
	}

	if entries == nil {
		entries = []ports.CommitEntry{}
	}

	return TaskHistoryResult{
		TaskID:  task.ID,
		Title:   task.Title,
		Entries: entries,
	}, nil
}

