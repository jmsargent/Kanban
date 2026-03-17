package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kanban-tasks/kanban/internal/ports"
)

// GitAdapter implements ports.GitPort via real git CLI invocations and
// direct filesystem operations on the .git/ directory.
type GitAdapter struct{}

// NewGitAdapter constructs a GitAdapter.
func NewGitAdapter() *GitAdapter {
	return &GitAdapter{}
}

// RepoRoot returns the absolute path to the repository root.
// Returns ports.ErrNotGitRepo when the working directory is not inside a git repo.
func (a *GitAdapter) RepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", ports.ErrNotGitRepo
	}
	return strings.TrimSpace(string(out)), nil
}

// CommitMessagesInRange returns commit messages for commits reachable from `to`
// but not from `from` (equivalent to git log from..to --format=%B).
func (a *GitAdapter) CommitMessagesInRange(from, to string) ([]string, error) {
	out, err := exec.Command("git", "log", from+".."+to, "--format=%B").Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	var messages []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			messages = append(messages, line)
		}
	}
	return messages, nil
}

// CommitFiles stages the given paths and creates a commit with the supplied message.
// "[skip ci]" is always appended to the message to prevent CI recursion.
func (a *GitAdapter) CommitFiles(repoRoot, message string, paths []string) error {
	addArgs := append([]string{"add", "--"}, paths...)
	if err := runGitIn(repoRoot, addArgs...); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	annotated := message + " [skip ci]"
	if err := runGitIn(repoRoot, "commit", "-m", annotated); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// InstallHook writes the kanban commit-msg hook script to .git/hooks/commit-msg
// and sets it executable (0755). Overwrites any existing hook.
// The hook uses the absolute path of the currently running kanban binary so it
// works regardless of $PATH.
func (a *GitAdapter) InstallHook(repoRoot string) error {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		exe = "kanban" // fallback to PATH lookup
	}

	hookPath := filepath.Join(hooksDir, "commit-msg")
	script := fmt.Sprintf("#!/bin/sh\n%s _hook commit-msg \"$1\"\n", exe)
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		return fmt.Errorf("write hook: %w", err)
	}
	return nil
}

// AppendToGitignore adds entry to .gitignore at the repository root if not already present.
// Creates .gitignore if it does not exist.
func (a *GitAdapter) AppendToGitignore(repoRoot, entry string) error {
	gitignorePath := filepath.Join(repoRoot, ".gitignore")

	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == entry {
			return nil
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer func() { _ = f.Close() }()

	line := entry + "\n"
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		line = "\n" + line
	}
	if _, err = f.WriteString(line); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}

// GetIdentity returns the git author identity from git config.
// Full implementation in step 02-01; this stub satisfies the port contract.
func (a *GitAdapter) GetIdentity() (ports.Identity, error) {
	return ports.Identity{}, ports.ErrGitIdentityNotConfigured
}

// runGitIn runs a git command with the given arguments in the specified directory.
func runGitIn(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Compile-time interface compliance check.
var _ ports.GitPort = (*GitAdapter)(nil)
