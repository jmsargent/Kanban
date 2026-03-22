package usecases

import (
	"os"
	"os/exec"
)

// OpenEditor opens the file at filePath in the given editor command.
// editorCmd is the path or name of the editor executable.
// Returns an error if the editor exits non-zero.
func OpenEditor(editorCmd, filePath string) error {
	cmd := exec.Command(editorCmd, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// resolveEditor returns the editor command to use, preferring $EDITOR then vi.
func resolveEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	return editor
}
