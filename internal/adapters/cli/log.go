package cli

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// transitionPattern matches commit messages of the form "TASK-NNN: from->to [trigger:label]".
var transitionPattern = regexp.MustCompile(`[A-Z]+-\d+:\s*(\S+->[\w-]+)\s*\[trigger:(\w+)\]`)

// renderLogEntry formats a CommitEntry as a structured log line.
// Format: <ISO8601-timestamp> <from->to> <author-email> [trigger:<label>] commit:<sha7>
// Falls back to raw message fields when the commit message does not match the transition format.
func renderLogEntry(entry ports.CommitEntry) string {
	ts := entry.Timestamp.UTC().Format("2006-01-02T15:04:05Z")

	matches := transitionPattern.FindStringSubmatch(entry.Message)
	if len(matches) == 3 {
		transition := matches[1]
		trigger := matches[2]
		sha7 := entry.SHA
		if len(sha7) > 7 {
			sha7 = sha7[:7]
		}
		return fmt.Sprintf("  %s %s %s [trigger:%s] commit:%s", ts, transition, entry.Author, trigger, sha7)
	}

	// Fallback for commits not written by kanban start (e.g. initial add commit).
	sha7 := entry.SHA
	if len(sha7) > 7 {
		sha7 = sha7[:7]
	}
	return fmt.Sprintf("  %s %s %s commit:%s", ts, entry.Author, entry.Message, sha7)
}

// NewLogCommand builds the "kanban log <task-id>" cobra command.
// It displays the task header and its transition history.
func NewLogCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, log ports.TransitionLogRepository) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "log <task-id>",
		Short:         "Show the git commit history for a task",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			exitError := func(msg string) error {
				writeLine(errOut, msg)
				osExit(1)
				return nil
			}

			repoRoot, err := git.RepoRoot()
			if err != nil {
				return exitError("Not a git repository. Run 'kanban init' first.")
			}

			uc := usecases.NewGetTaskHistory(config, tasks, git, log)
			result, err := uc.Execute(repoRoot, taskID)
			if err != nil {
				if errors.Is(err, ports.ErrTaskNotFound) {
					return exitError(fmt.Sprintf("Task %s not found. Run 'kanban board' to see valid task IDs.", taskID))
				}
				if errors.Is(err, ports.ErrNotInitialised) {
					return exitError("kanban not initialised. Run 'kanban init' first.")
				}
				return exitError(fmt.Sprintf("Error: %v", err))
			}

			writeLine(out, fmt.Sprintf("%s: %s", result.TaskID, result.Title))

			if len(result.Entries) == 0 {
				writeLine(out, "No transitions recorded yet.")
				return nil
			}

			for _, entry := range result.Entries {
				writeLine(out, renderLogEntry(entry))
			}
			return nil
		},
	}

	return cmd
}
