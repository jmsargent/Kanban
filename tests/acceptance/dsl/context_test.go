package dsl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewContext_KanbanBinEnvironmentVariable(t *testing.T) {
	oldBin := os.Getenv("KANBAN_BIN")
	defer func() {
		_ = os.Setenv("KANBAN_BIN", oldBin)
	}()

	t.Run("uses KANBAN_BIN when set", func(t *testing.T) {
		if err := os.Setenv("KANBAN_BIN", "/custom/kanban"); err != nil {
			t.Fatalf("failed to set KANBAN_BIN: %v", err)
		}
		ctx := NewContext(t)
		if ctx.binPath != "/custom/kanban" {
			t.Errorf("expected binPath to be /custom/kanban, got %q", ctx.binPath)
		}
	})

	t.Run("resolves default path when KANBAN_BIN is empty", func(t *testing.T) {
		if err := os.Setenv("KANBAN_BIN", ""); err != nil {
			t.Fatalf("failed to clear KANBAN_BIN: %v", err)
		}
		ctx := NewContext(t)
		if ctx.binPath == "" {
			t.Error("expected binPath to be resolved, but got empty string")
		}
		if !filepath.IsAbs(ctx.binPath) {
			t.Errorf("expected absolute path, got %q", ctx.binPath)
		}
	})
}
