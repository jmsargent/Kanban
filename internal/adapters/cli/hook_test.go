package cli_test

// Test Budget: 2 behaviors x 2 = 4 max unit tests (using 2)
// Behavior 1: _hook commit-msg is a no-op — exits 0, no stdout, no stderr.
// Behavior 2: install-hook exits 1 with deprecation message.

import (
	"bytes"
	"strings"
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

func TestInstallHookCommand_ExitsOneWithDeprecationMessage(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	capturedCode := 0
	cli.SetOsExit(func(code int) { capturedCode = code })
	t.Cleanup(func() { cli.SetOsExit(nil) })

	cmd := cli.NewInstallHookCommand()
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	root := &cobra.Command{Use: "kanban", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(cmd)
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"install-hook"})

	_ = root.Execute()

	if capturedCode != 1 {
		t.Errorf("expected exit code 1, got %d", capturedCode)
	}
	combined := outBuf.String() + errBuf.String()
	if !strings.Contains(combined, "install-hook has been removed") {
		t.Errorf("expected deprecation message in output, got: %q", combined)
	}
}
