package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewHookCommand builds the "kanban _hook" internal command family.
// These commands are invoked by git hooks installed during "kanban init".
func NewHookCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, log ports.TransitionLogRepository) *cobra.Command {
	hook := &cobra.Command{
		Use:    "_hook",
		Short:  "Internal hook commands (invoked by git hooks)",
		Hidden: true,
	}
	hook.AddCommand(newCommitMsgHookCommand(git, config, tasks, log))
	return hook
}

// newCommitMsgHookCommand handles "kanban _hook commit-msg <msg-file>".
// It reads the commit message, finds TASK-NNN references, and advances todo tasks to in-progress.
// The hook ALWAYS exits 0 — panics and errors are written to .kanban/hook.log.
func newCommitMsgHookCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, log ports.TransitionLogRepository) *cobra.Command {
	return &cobra.Command{
		Use:    "commit-msg <msg-file>",
		Short:  "Advance tasks referenced in a commit message",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			msgFile := args[0]

			repoRoot, err := git.RepoRoot()
			if err != nil {
				return nil // not in a git repo; silently skip
			}

			data, readErr := os.ReadFile(msgFile)
			if readErr != nil {
				appendHookLog(repoRoot, fmt.Sprintf("cannot read commit-msg file: %v", readErr))
				return nil
			}

			uc := usecases.NewTransitionToInProgress(config, tasks, log, cmd.OutOrStdout())
			if execErr := uc.Execute(repoRoot, string(data)); execErr != nil {
				appendHookLog(repoRoot, fmt.Sprintf("transition error: %v", execErr))
			}
			return nil
		},
	}
}

// appendHookLog appends msg to {repoRoot}/.kanban/hook.log in append-only mode.
// Silently discards write errors to prevent blocking the commit.
func appendHookLog(repoRoot, msg string) {
	logPath := filepath.Join(repoRoot, ".kanban", "hook.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, "[%s] %s\n", time.Now().UTC().Format(time.RFC3339), msg)
}
