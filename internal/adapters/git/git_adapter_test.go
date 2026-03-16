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

// Test Budget: 4 behaviors x 2 = 8 max unit tests (using 4)

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

func TestCommitFiles_CreatesCommitAnnotatedWithSkipCI(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	// Need at least one existing commit so we can add another
	writeAndCommit(t, dir, "init.txt", "init", "initial commit")

	// Create a new file to be staged and committed via the adapter
	newFile := filepath.Join(dir, "task.md")
	if err := os.WriteFile(newFile, []byte("# task"), 0o644); err != nil {
		t.Fatalf("write task.md: %v", err)
	}

	adapter := gitadapter.NewGitAdapter()
	err := adapter.CommitFiles(dir, "add task card", []string{"task.md"})
	if err != nil {
		t.Fatalf("CommitFiles: %v", err)
	}

	// Read the last commit message from the repo
	out, err := exec.Command("git", "-C", dir, "log", "-1", "--format=%s").Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	subject := strings.TrimSpace(string(out))
	if !strings.Contains(subject, "[skip ci]") {
		t.Errorf("commit subject %q does not contain [skip ci]", subject)
	}
	if !strings.Contains(subject, "add task card") {
		t.Errorf("commit subject %q does not contain original message", subject)
	}
}

func TestInstallHook_WritesExecutableCommitMsgScript(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	adapter := gitadapter.NewGitAdapter()
	if err := adapter.InstallHook(dir); err != nil {
		t.Fatalf("InstallHook: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "commit-msg")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook not written: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("hook file mode %o is not executable", info.Mode())
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	script := string(content)
	if !strings.HasPrefix(script, "#!/bin/sh") {
		t.Errorf("hook does not start with #!/bin/sh:\n%s", script)
	}

	exe, _ := os.Executable()
	if !strings.Contains(script, exe+" _hook commit-msg") {
		t.Errorf("hook does not use absolute path %q:\n%s", exe, script)
	}
}
