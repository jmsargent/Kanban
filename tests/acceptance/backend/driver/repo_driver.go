package driver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// RepoDriver creates and manages temporary git repositories for acceptance tests.
// It handles repo init, .kanban/tasks/ structure, and optional bare remote for
// push verification.
type RepoDriver struct {
	t       *testing.T
	repoDir string
	bareDir string
}

// NewRepoDriver constructs a RepoDriver with a temp git repo containing a
// baseline commit and .kanban/tasks/ directory.
func NewRepoDriver(t *testing.T) *RepoDriver {
	t.Helper()
	dir := t.TempDir()

	d := &RepoDriver{t: t, repoDir: dir}
	d.mustGit("init")
	d.mustGit("config", "user.email", "test@example.com")
	d.mustGit("config", "user.name", "Test User")

	// Create .kanban structure
	tasksDir := filepath.Join(dir, ".kanban", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("create .kanban/tasks/: %v", err)
	}

	// Write default config
	configPath := filepath.Join(dir, ".kanban", "config")
	configContent := `ci_task_pattern: "TASK-"
columns:
  - name: todo
    label: Todo
  - name: in-progress
    label: Doing
  - name: done
    label: Done
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Baseline commit
	d.mustGit("add", ".")
	d.mustGit("commit", "-m", "Initial commit with .kanban structure")

	return d
}

// RepoDir returns the absolute path to the temporary git repository.
func (d *RepoDriver) RepoDir() string {
	return d.repoDir
}

// BareDir returns the absolute path to the bare remote repository, or empty
// if no remote was configured.
func (d *RepoDriver) BareDir() string {
	return d.bareDir
}

// SetupBareRemote creates a bare git repository and adds it as "origin" to
// the working repo. Used for push verification tests.
func (d *RepoDriver) SetupBareRemote() {
	d.t.Helper()
	d.bareDir = d.t.TempDir()

	bareGit := func(args ...string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = d.bareDir
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			d.t.Fatalf("git %s in bare: %v\n%s", strings.Join(args, " "), err, buf.String())
		}
	}

	bareGit("init", "--bare")
	d.mustGit("remote", "add", "origin", d.bareDir)
	d.mustGit("push", "-u", "origin", "HEAD")
}

// SeedTask writes a task file to .kanban/tasks/ and commits it.
func (d *RepoDriver) SeedTask(id, title, status, assignee string) {
	d.SeedTaskWithDate(id, title, status, assignee, "")
}

// SeedTaskWithDate writes a task file with an optional created_at field and commits it.
// createdAt should be an RFC3339 string (e.g. "2024-01-01T00:00:00Z") or empty to omit.
func (d *RepoDriver) SeedTaskWithDate(id, title, status, assignee, createdAt string) {
	d.t.Helper()
	if status == "" {
		status = "todo"
	}
	body := fmt.Sprintf("id: %s\ntitle: %s\nstatus: %s\n", id, title, status)
	if assignee != "" {
		body += fmt.Sprintf("assignee: %s\n", assignee)
	}
	if createdAt != "" {
		body += fmt.Sprintf("created_at: %s\n", createdAt)
	}
	content := fmt.Sprintf("---\n%s---\n", body)

	taskPath := filepath.Join(d.repoDir, ".kanban", "tasks", id+".md")
	if err := os.WriteFile(taskPath, []byte(content), 0644); err != nil {
		d.t.Fatalf("write task file %s: %v", id, err)
	}
	d.mustGit("add", taskPath)
	d.mustGit("commit", "-m", fmt.Sprintf("Add task %s", id))
}

// TaskFileExists returns true if the task file for the given ID exists.
func (d *RepoDriver) TaskFileExists(id string) bool {
	taskPath := filepath.Join(d.repoDir, ".kanban", "tasks", id+".md")
	_, err := os.Stat(taskPath)
	return err == nil
}

// mustGit runs a git command in the repo directory and fails the test on error.
func (d *RepoDriver) mustGit(args ...string) string {
	d.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = d.repoDir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		d.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, buf.String())
	}
	return strings.TrimSpace(buf.String())
}
