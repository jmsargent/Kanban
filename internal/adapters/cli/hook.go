package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewHookCommand builds the "kanban _hook" internal command family.
// These commands are invoked by git hooks installed during "kanban init".
func NewHookCommand(git ports.GitPort, config ports.ConfigRepository) *cobra.Command {
	hook := &cobra.Command{
		Use:    "_hook",
		Short:  "Internal hook commands (invoked by git hooks)",
		Hidden: true,
	}
	hook.AddCommand(newCommitMsgHookCommand(git, config))
	return hook
}

// newCommitMsgHookCommand handles "kanban _hook commit-msg <msg-file>".
// It reads the commit message, finds TASK-NNN references, and advances tasks.
func newCommitMsgHookCommand(git ports.GitPort, config ports.ConfigRepository) *cobra.Command {
	return &cobra.Command{
		Use:    "commit-msg <msg-file>",
		Short:  "Advance tasks referenced in a commit message",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			msgFile := args[0]
			data, err := os.ReadFile(msgFile)
			if err != nil {
				// Hook must not block commits; log and exit 0.
				fmt.Fprintf(os.Stderr, "kanban hook: cannot read commit-msg file: %v\n", err)
				return nil
			}

			repoRoot, err := git.RepoRoot()
			if err != nil {
				return nil // not in a git repo; silently skip
			}

			tasks := filesystem.NewTaskRepository()
			uc := usecases.NewTransitionTask(config, tasks)
			if err := uc.AdvanceByCommitMessage(repoRoot, string(data)); err != nil {
				// Log but do not block the commit.
				fmt.Fprintf(os.Stderr, "kanban hook: %v\n", err)
			}
			return nil
		},
	}
}
