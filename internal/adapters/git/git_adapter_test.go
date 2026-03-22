package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gitadapter "github.com/kanban-tasks/kanban/internal/adapters/git"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// initRepo initialises a bare git repository in dir and configures user identity.
func initRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", dir},
		{"-C", dir, "config", "user.email", "test@example.com"},
		{"-C", dir, "config", "user.name", "Test User"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
}

// writeAndCommit creates a file and commits it in dir.
func writeAndCommit(t *testing.T, dir, filename, content, message string) {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", filename, err)
	}
	for _, args := range [][]string{
		{"-C", dir, "add", filename},
		{"-C", dir, "commit", "-m", message},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
}

// refSHA resolves a ref to its SHA in dir.
func refSHA(t *testing.T, dir, ref string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", ref).Output()
	if err != nil {
		t.Fatalf("rev-parse %s: %v", ref, err)
	}
	return strings.TrimSpace(string(out))
}

// Test Budget: 5 behaviors x 2 = 10 max unit tests (using 5)

func TestRepoRoot_ReturnsErrNotGitRepo_WhenOutsideGitRepo(t *testing.T) {
	dir := t.TempDir()
	// Change working dir to non-git directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	adapter := gitadapter.NewGitAdapter()
	_, gotErr := adapter.RepoRoot()
	if gotErr != ports.ErrNotGitRepo {
		t.Errorf("RepoRoot() error = %v, want ports.ErrNotGitRepo", gotErr)
	}
}

func TestCommitMessagesInRange_ReturnsMessagesInRange(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	writeAndCommit(t, dir, "a.txt", "a", "first commit")
	fromSHA := refSHA(t, dir, "HEAD")

	writeAndCommit(t, dir, "b.txt", "b", "second commit")
	writeAndCommit(t, dir, "c.txt", "c", "third commit")
	toSHA := refSHA(t, dir, "HEAD")

	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) }) //nolint:errcheck
	os.Chdir(dir)                           //nolint:errcheck

	adapter := gitadapter.NewGitAdapter()
	messages, err := adapter.CommitMessagesInRange(fromSHA, toSHA)
	if err != nil {
		t.Fatalf("CommitMessagesInRange: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("got %d messages, want 2: %v", len(messages), messages)
	}
	if messages[0] != "third commit" {
		t.Errorf("messages[0] = %q, want %q", messages[0], "third commit")
	}
	if messages[1] != "second commit" {
		t.Errorf("messages[1] = %q, want %q", messages[1], "second commit")
	}
}


func TestGetIdentity_ReturnsConfiguredNameAndEmail(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir) // configures user.name = "Test User", user.email = "test@example.com"

	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) }) //nolint:errcheck
	os.Chdir(dir)                           //nolint:errcheck

	adapter := gitadapter.NewGitAdapter()
	identity, err := adapter.GetIdentity()
	if err != nil {
		t.Fatalf("GetIdentity: %v", err)
	}
	if identity.Name != "Test User" {
		t.Errorf("identity.Name = %q, want %q", identity.Name, "Test User")
	}
	if identity.Email != "test@example.com" {
		t.Errorf("identity.Email = %q, want %q", identity.Email, "test@example.com")
	}
}

