package filesystem

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// TransitionLogAdapter implements ports.TransitionLogRepository using a plain
// append-only text file at <repoRoot>/.kanban/transitions.log.
//
// Each line has five space-separated fields:
//
//	<ISO8601_UTC> <TASK-ID> <from>-><to> <author_email> <trigger>
//
// Concurrent appends are safe: each write acquires an exclusive file lock
// (syscall.LOCK_EX) before writing and releases it immediately after.
type TransitionLogAdapter struct{}

// NewTransitionLogAdapter constructs a TransitionLogAdapter.
func NewTransitionLogAdapter() *TransitionLogAdapter {
	return &TransitionLogAdapter{}
}

func transitionsLogFilePath(repoRoot string) string {
	return filepath.Join(repoRoot, ".kanban", "transitions.log")
}

// Append records a new transition entry. It creates the file and parent
// directories if they do not exist. Writes are serialised with an exclusive
// flock so concurrent callers never interleave bytes.
func (a *TransitionLogAdapter) Append(repoRoot string, entry domain.TransitionEntry) error {
	dir := filepath.Join(repoRoot, ".kanban")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create .kanban dir: %w", err)
	}

	logPath := transitionsLogFilePath(repoRoot)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open transitions.log: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock transitions.log: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	line := fmt.Sprintf("%s %s %s->%s %s %s\n",
		entry.Timestamp.UTC().Format(time.RFC3339),
		entry.TaskID,
		string(entry.From),
		string(entry.To),
		entry.Author,
		entry.Trigger,
	)

	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write transitions.log: %w", err)
	}
	return nil
}

// LatestStatus returns the most recent TaskStatus recorded for taskID.
// Returns (StatusTodo, nil) when no entries exist for the task — missing
// entry means the task has never transitioned from its implicit default state.
func (a *TransitionLogAdapter) LatestStatus(repoRoot, taskID string) (domain.TaskStatus, error) {
	entries, err := a.readAllForTask(repoRoot, taskID)
	if err != nil {
		return domain.StatusTodo, err
	}
	if len(entries) == 0 {
		return domain.StatusTodo, nil
	}
	return entries[len(entries)-1].To, nil
}

// History returns all recorded transitions for taskID, oldest first.
// Returns an empty slice (not an error) when no entries exist.
func (a *TransitionLogAdapter) History(repoRoot, taskID string) ([]domain.TransitionEntry, error) {
	return a.readAllForTask(repoRoot, taskID)
}

// readAllForTask reads the log file and returns all entries for taskID.
func (a *TransitionLogAdapter) readAllForTask(repoRoot, taskID string) ([]domain.TransitionEntry, error) {
	logPath := transitionsLogFilePath(repoRoot)
	f, err := os.Open(logPath)
	if os.IsNotExist(err) {
		return []domain.TransitionEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open transitions.log: %w", err)
	}
	defer func() { _ = f.Close() }()

	var entries []domain.TransitionEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		entry, err := parseLogLine(line)
		if err != nil {
			continue // skip malformed lines
		}
		if entry.TaskID == taskID {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan transitions.log: %w", err)
	}
	return entries, nil
}

// parseLogLine parses a single transitions.log line into a TransitionEntry.
// Format: <ISO8601_UTC> <TASK-ID> <from>-><to> <author_email> <trigger>
func parseLogLine(line string) (domain.TransitionEntry, error) {
	fields := strings.Fields(line)
	if len(fields) != 5 {
		return domain.TransitionEntry{}, fmt.Errorf("malformed log line: %q", line)
	}

	ts, err := time.Parse(time.RFC3339, fields[0])
	if err != nil {
		return domain.TransitionEntry{}, fmt.Errorf("parse timestamp %q: %w", fields[0], err)
	}

	transition := fields[2]
	parts := strings.SplitN(transition, "->", 2)
	if len(parts) != 2 {
		return domain.TransitionEntry{}, fmt.Errorf("malformed transition %q", transition)
	}

	return domain.TransitionEntry{
		Timestamp: ts,
		TaskID:    fields[1],
		From:      domain.TaskStatus(parts[0]),
		To:        domain.TaskStatus(parts[1]),
		Author:    fields[3],
		Trigger:   fields[4],
	}, nil
}

// ensure compile-time interface compliance
var _ ports.TransitionLogRepository = (*TransitionLogAdapter)(nil)
