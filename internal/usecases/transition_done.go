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

	var transitioned []string
	for _, id := range taskIDs {
		if _, findErr := u.tasks.FindByID(repoRoot, id); findErr != nil {
			continue
		}

		// Derive current status from transitions.log (authoritative source).
		currentStatus, statusErr := u.log.LatestStatus(repoRoot, id)
		if statusErr != nil {
			continue
		}
		if currentStatus != domain.StatusInProgress {
			continue
		}

		entry := domain.TransitionEntry{
			Timestamp: time.Now().UTC(),
			TaskID:    id,
			From:      domain.StatusInProgress,
			To:        domain.StatusDone,
			Author:    "ci-done",
			Trigger:   "ci-done",
		}
		if err := u.log.Append(repoRoot, entry); err != nil {
			return fmt.Errorf("append transition for %s: %w", id, err)
		}

		_, _ = fmt.Fprintf(u.output, "[kanban] %s moved  in-progress -> done\n", id)
		transitioned = append(transitioned, id)
	}

	if len(transitioned) == 0 {
		_, _ = fmt.Fprintf(u.output, "[kanban] no tasks to transition to done\n")
		return nil
	}

	transitionsLogPath := filepath.Join(repoRoot, ".kanban", "transitions.log")
	return u.git.CommitFiles(repoRoot, "chore(kanban): mark tasks done [skip ci]", []string{transitionsLogPath})
}

