package cli_test

// Test Budget: 1 behavior x 2 = 2 max unit tests (using 1)
// Behavior: _hook commit-msg is a no-op — exits 0, no stdout, no stderr.

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
)

func TestHookCommand_CommitMsg_IsNoOp(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	cmd := cli.NewHookCommand()
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	root := &cobra.Command{Use: "kanban", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(cmd)
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"_hook", "commit-msg", "any-file"})

	err := root.Execute()

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("expected no stdout output, got: %q", outBuf.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected no stderr output, got: %q", errBuf.String())
	}
}
