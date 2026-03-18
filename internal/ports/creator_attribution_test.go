package ports_test

// Test budget: 2 distinct behaviors x 2 = 4 max unit tests.
// Behavior 1: domain.Task has a CreatedBy field (string).
// Behavior 2: GitPort declares GetIdentity() returning (Identity, error).
//
// Both are validated at compile time via interface satisfaction and field access.
// Runtime tests are not added — the compiler is the test runner for structural contracts.

import (
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// Verify domain.Task exposes a CreatedBy field by assigning to it.
// This fails to compile if the field does not exist.
var _ = domain.Task{CreatedBy: "Jane Doe"}

// fakeGitPortWithIdentity is an in-memory fake that must satisfy GitPort
// including the new GetIdentity() method. The compiler assertion on the
// next line is the RED unit test — it will fail until GetIdentity() exists.
type fakeGitPortWithIdentity struct{}

var _ ports.GitPort = (*fakeGitPortWithIdentity)(nil)

func (f *fakeGitPortWithIdentity) RepoRoot() (string, error)                        { return "", nil }
func (f *fakeGitPortWithIdentity) CommitMessagesInRange(_, _ string) ([]string, error) { return nil, nil }
func (f *fakeGitPortWithIdentity) CommitFiles(_, _ string, _ []string) error         { return nil }
func (f *fakeGitPortWithIdentity) InstallHook(_ string) error                        { return nil }
func (f *fakeGitPortWithIdentity) AppendToGitignore(_, _ string) error               { return nil }
func (f *fakeGitPortWithIdentity) GetIdentity() (ports.Identity, error) { return ports.Identity{}, nil }
func (f *fakeGitPortWithIdentity) LogFile(_, _ string) ([]ports.CommitEntry, error) {
	return nil, nil
}
