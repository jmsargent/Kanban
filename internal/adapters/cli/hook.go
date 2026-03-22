package cli

import (
	"fmt"

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

// NewInstallHookCommand returns a deprecated stub for "kanban install-hook".
// It exits 1 with a removal notice. Hidden so it does not appear in --help.
func NewInstallHookCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "install-hook",
		Short:  "Deprecated: kanban no longer manages git hooks",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.ErrOrStderr(), "install-hook has been removed; kanban no longer manages git hooks")
			osExit(1)
			return nil
		},
	}
}
