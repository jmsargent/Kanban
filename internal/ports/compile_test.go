// Package ports_test contains a compilation test that verifies in-memory fakes
// can satisfy the port interfaces. This file is the RED_ACCEPTANCE checkpoint:
// it must fail to compile until the interfaces are declared.
package ports_test

import (
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// fakeTaskRepo is an in-memory fake that must satisfy TaskRepository.
// The compiler enforces the contract — no runtime assertions needed.
type fakeTaskRepo struct {
	tasks map[string]domain.Task
}

var _ ports.TaskRepository = (*fakeTaskRepo)(nil)

func (r *fakeTaskRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	t, ok := r.tasks[taskID]
	if !ok {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	return t, nil
}

func (r *fakeTaskRepo) Save(repoRoot string, task domain.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *fakeTaskRepo) ListAll(repoRoot string) ([]domain.Task, error) {
	out := make([]domain.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		out = append(out, t)
	}
	return out, nil
}

func (r *fakeTaskRepo) Update(repoRoot string, task domain.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *fakeTaskRepo) Delete(repoRoot, taskID string) error {
	delete(r.tasks, taskID)
	return nil
}

func (r *fakeTaskRepo) NextID(repoRoot string) (string, error) {
	return "TASK-001", nil
}

// fakeConfigRepo is an in-memory fake that must satisfy ConfigRepository.
type fakeConfigRepo struct {
	cfg ports.Config
}

var _ ports.ConfigRepository = (*fakeConfigRepo)(nil)

func (r *fakeConfigRepo) Read(repoRoot string) (ports.Config, error) {
	return r.cfg, nil
}

func (r *fakeConfigRepo) Write(repoRoot string, config ports.Config) error {
	r.cfg = config
	return nil
}

// fakeGitPort is an in-memory fake that must satisfy GitPort.
type fakeGitPort struct {
	root string
}

var _ ports.GitPort = (*fakeGitPort)(nil)

func (g *fakeGitPort) RepoRoot() (string, error) {
	return g.root, nil
}

func (g *fakeGitPort) CommitMessagesInRange(from, to string) ([]string, error) {
	return nil, nil
}

func (g *fakeGitPort) CommitFiles(repoRoot, message string, paths []string) error {
	return nil
}

func (g *fakeGitPort) InstallHook(repoRoot string) error {
	return nil
}

func (g *fakeGitPort) AppendToGitignore(repoRoot, entry string) error {
	return nil
}

func (g *fakeGitPort) GetIdentity() (ports.Identity, error) {
	return ports.Identity{}, nil
}
