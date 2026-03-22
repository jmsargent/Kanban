package cli

import (
	"github.com/spf13/cobra"
)

// NewHookCommand builds the "kanban _hook" internal command family.
// These commands are invoked by git hooks installed during "kanban init".
func NewHookCommand() *cobra.Command {
	hook := &cobra.Command{
		Use:    "_hook",
		Short:  "Internal hook commands (invoked by git hooks)",
		Hidden: true,
	}
	hook.AddCommand(newCommitMsgHookCommand())
	return hook
}

// newCommitMsgHookCommand handles "kanban _hook commit-msg <msg-file>".
// No-op: kanban no longer performs automatic state transitions on commit.
// The hook ALWAYS exits 0.
func newCommitMsgHookCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "commit-msg <msg-file>",
		Short:  "No-op: automatic transitions have been removed",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
