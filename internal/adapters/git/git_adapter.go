package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmsargent/kanban/internal/ports"
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
// Returns ErrGitIdentityNotConfigured when user.name is not set.
func (a *GitAdapter) GetIdentity() (ports.Identity, error) {
	nameOut, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return ports.Identity{}, ports.ErrGitIdentityNotConfigured
	}
	name := strings.TrimSpace(string(nameOut))
	if name == "" {
		return ports.Identity{}, ports.ErrGitIdentityNotConfigured
	}

	var email string
	emailOut, err := exec.Command("git", "config", "user.email").Output()
	if err == nil {
		email = strings.TrimSpace(string(emailOut))
	}

	return ports.Identity{Name: name, Email: email}, nil
}

// LogFile returns the commit history for a specific file within the repository
// at repoRoot, following renames. Results are oldest-first.
// Format: "%H|%aI|%ae|%s" — SHA, author ISO 8601 date, author email, subject.
func (a *GitAdapter) LogFile(repoRoot, filePath string) ([]ports.CommitEntry, error) {
	cmd := exec.Command("git", "log", "--follow", "--format=%H|%aI|%ae|%s", "--", filePath)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log --follow: %w", err)
	}

	var entries []ports.CommitEntry
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}
		ts, err := time.Parse(time.RFC3339, parts[1])
		if err != nil {
			ts = time.Time{}
		}
		entries = append(entries, ports.CommitEntry{
			SHA:       parts[0],
			Timestamp: ts,
			Author:    parts[2],
			Message:   parts[3],
		})
	}

	// Reverse to return oldest-first (git log returns newest-first by default).
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	if entries == nil {
		entries = []ports.CommitEntry{}
	}
	return entries, nil
}

// Compile-time interface compliance check.
var _ ports.GitPort = (*GitAdapter)(nil)
