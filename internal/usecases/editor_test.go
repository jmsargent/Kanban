package usecases_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmsargent/kanban/internal/usecases"
)

// Test Budget: 2 behaviors x 2 = 4 max unit tests (using 2)
// Behaviors:
//  1. OpenEditor runs the editor command with the file path as argument
//  2. OpenEditor returns an error when the editor command fails

// TestOpenEditor_RunsEditorWithFilePath verifies that OpenEditor invokes the
// given editor command with the temp file path as its argument. The observable
// outcome is that the editor script writes a sentinel value into the file.
func TestOpenEditor_RunsEditorWithFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "task.yaml")
	if err := os.WriteFile(tmpFile, []byte("title: \"\"\n"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	// Use a shell script as the "editor": it writes a known marker into the file.
	scriptPath := filepath.Join(tmpDir, "editor.sh")
	script := "#!/bin/sh\nprintf 'title: Sentinel\\n' > \"$1\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write editor script: %v", err)
	}

	if err := usecases.OpenEditor(scriptPath, tmpFile); err != nil {
		t.Fatalf("OpenEditor: %v", err)
	}

	got, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read tmp file: %v", err)
	}
	if string(got) != "title: Sentinel\n" {
		t.Errorf("expected file to contain sentinel written by editor, got: %q", string(got))
	}
}

// TestOpenEditor_ReturnsErrorWhenEditorFails verifies that OpenEditor propagates
// a non-zero exit code from the editor as an error.
func TestOpenEditor_ReturnsErrorWhenEditorFails(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "task.yaml")
	if err := os.WriteFile(tmpFile, []byte(""), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	scriptPath := filepath.Join(tmpDir, "fail-editor.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write editor script: %v", err)
	}

	err := usecases.OpenEditor(scriptPath, tmpFile)
	if err == nil {
		t.Error("expected an error when editor exits non-zero, got nil")
	}
}
