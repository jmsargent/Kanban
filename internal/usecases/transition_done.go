package usecases

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// TransitionToDone advances in-progress tasks to done when they are referenced
// in the commit messages of the CI pipeline range.
type TransitionToDone struct {
	git    ports.GitPort
	tasks  ports.TaskRepository
	config ports.ConfigRepository
	log    ports.TransitionLogRepository
	output io.Writer
}

// NewTransitionToDone constructs a TransitionToDone use case.
func NewTransitionToDone(
	git ports.GitPort,
	tasks ports.TaskRepository,
	config ports.ConfigRepository,
	log ports.TransitionLogRepository,
	output io.Writer,
) *TransitionToDone {
	return &TransitionToDone{git: git, tasks: tasks, config: config, log: log, output: output}
}

// Execute reads commits in the range [from, to], finds task IDs referenced in
// those messages, and advances any in-progress tasks to done.
// If no tasks are advanced, no commit is made.
func (u *TransitionToDone) Execute(repoRoot, from, to string) error {
	cfg, err := u.config.Read(repoRoot)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	pattern := cfg.CITaskPattern
	if pattern == "" {
		pattern = `TASK-\d+`
	}
	taskPattern := regexp.MustCompile(pattern)

	messages, err := u.git.CommitMessagesInRange(from, to)
	if err != nil {
		return fmt.Errorf("commit messages: %w", err)
	}

	seen := make(map[string]bool)
	var taskIDs []string
	for _, msg := range messages {
		for _, id := range taskPattern.FindAllString(msg, -1) {
			if !seen[id] {
				seen[id] = true
				taskIDs = append(taskIDs, id)
			}
		}
	}

	var updatedPaths []string
	for _, id := range taskIDs {
		task, findErr := u.tasks.FindByID(repoRoot, id)
		if findErr != nil {
			continue
		}
		if task.Status != domain.StatusInProgress {
			continue
		}
		task.Status = domain.StatusDone
		if updateErr := u.tasks.Update(repoRoot, task); updateErr != nil {
			return fmt.Errorf("update task %s: %w", id, updateErr)
		}

		// Record in the log (authoritative for GetBoard status). Non-fatal.
		entry := domain.TransitionEntry{
			Timestamp: time.Now().UTC(),
			TaskID:    id,
			From:      domain.StatusInProgress,
			To:        domain.StatusDone,
			Author:    "ci-done",
			Trigger:   "ci-done",
		}
		_ = u.log.Append(repoRoot, entry)

		_, _ = fmt.Fprintf(u.output, "[kanban] %s moved  in-progress -> done\n", id)
		updatedPaths = append(updatedPaths, taskFilePath(repoRoot, id))
	}

	if len(updatedPaths) == 0 {
		return nil
	}
	return u.git.CommitFiles(repoRoot, "chore(kanban): mark tasks done [skip ci]", updatedPaths)
}

// taskFilePath returns the filesystem path for a task file given its ID and repoRoot.
func taskFilePath(repoRoot, taskID string) string {
	return filepath.Join(repoRoot, ".kanban", "tasks", taskID+".md")
}
