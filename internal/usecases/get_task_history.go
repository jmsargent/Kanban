package usecases

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/kanban-tasks/kanban/internal/ports"
)

// TaskHistoryResult is the output of GetTaskHistory.Execute.
type TaskHistoryResult struct {
	TaskID  string
	Title   string
	Entries []ports.CommitEntry
}

// GetTaskHistory retrieves the transition history for a task, merging entries
// from the transitions log (authoritative for state changes) with the git
// commit history of the task file (for commit-triggered transitions).
type GetTaskHistory struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
	git    ports.GitPort
	log    ports.TransitionLogRepository
}

// NewGetTaskHistory constructs a GetTaskHistory use case.
func NewGetTaskHistory(config ports.ConfigRepository, tasks ports.TaskRepository, git ports.GitPort, log ports.TransitionLogRepository) *GetTaskHistory {
	return &GetTaskHistory{config: config, tasks: tasks, git: git, log: log}
}

// Execute returns the task header and its transition history.
// Returns ErrNotInitialised when the repository has not been initialised.
// Returns ErrTaskNotFound when no task matches taskID.
// Returns an empty Entries slice (not an error) when no transitions exist.
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

	// Collect entries from transitions.log (authoritative for state transitions).
	logEntries, err := u.log.History(repoRoot, taskID)
	if err != nil {
		return TaskHistoryResult{}, fmt.Errorf("read transition history: %w", err)
	}

	// Collect entries from git log of the task file (legacy and commit-triggered).
	filePath := filepath.Join(".kanban", "tasks", taskID+".md")
	gitEntries, err := u.git.LogFile(repoRoot, filePath)
	if err != nil {
		return TaskHistoryResult{}, fmt.Errorf("git log file: %w", err)
	}

	// Merge: convert log entries to CommitEntry format and combine with git entries.
	var combined []ports.CommitEntry
	for _, e := range logEntries {
		combined = append(combined, ports.CommitEntry{
			SHA:       "",
			Timestamp: e.Timestamp,
			Author:    e.Author,
			Message:   fmt.Sprintf("%s: %s->%s [trigger:%s]", taskID, string(e.From), string(e.To), e.Trigger),
		})
	}
	combined = append(combined, gitEntries...)

	// Sort oldest first, deduplicate by (timestamp, message) proximity.
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Timestamp.Before(combined[j].Timestamp)
	})
	combined = deduplicateByTimestamp(combined)

	if combined == nil {
		combined = []ports.CommitEntry{}
	}

	return TaskHistoryResult{
		TaskID:  task.ID,
		Title:   task.Title,
		Entries: combined,
	}, nil
}

// deduplicateByTimestamp removes entries where a log entry and a git entry
// occurred within 2 seconds of each other (same transition, different source).
// The log entry (no SHA) takes precedence over the git entry.
func deduplicateByTimestamp(entries []ports.CommitEntry) []ports.CommitEntry {
	if len(entries) <= 1 {
		return entries
	}
	result := []ports.CommitEntry{entries[0]}
	for i := 1; i < len(entries); i++ {
		prev := result[len(result)-1]
		curr := entries[i]
		diff := curr.Timestamp.Sub(prev.Timestamp)
		if diff < 0 {
			diff = -diff
		}
		if diff < 2*time.Second {
			// Keep the one with richer info: prefer log entry (has from->to) over plain git entry.
			if isTransitionMessage(prev.Message) {
				// prev is already the log entry; skip curr
				continue
			}
			if isTransitionMessage(curr.Message) {
				// curr is the log entry; replace prev
				result[len(result)-1] = curr
				continue
			}
		}
		result = append(result, curr)
	}
	return result
}

// isTransitionMessage returns true if the commit message contains a state
// transition arrow (e.g. "todo->in-progress").
func isTransitionMessage(msg string) bool {
	for i := 0; i < len(msg)-2; i++ {
		if msg[i] == '-' && msg[i+1] == '>' {
			return true
		}
	}
	return false
}

