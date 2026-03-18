package domain

import "time"

// TransitionEntry records a single status change for a task in the audit trail.
// It is derived from git history or the transitions log.
type TransitionEntry struct {
	Timestamp time.Time
	TaskID    string
	From      TaskStatus
	To        TaskStatus
	Author    string
	Trigger   string // "manual" | "commit:<sha7>" | "ci-done:<sha7>"
}
