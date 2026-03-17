package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// TaskRepository implements ports.TaskRepository using Markdown files with YAML
// front matter stored under {repoRoot}/.kanban/tasks/.
type TaskRepository struct{}

// NewTaskRepository constructs a TaskRepository.
func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
}

// taskFrontMatter is the YAML serialisation shape for a Task header.
type taskFrontMatter struct {
	ID        string `yaml:"id"`
	Title     string `yaml:"title"`
	Status    string `yaml:"status"`
	Priority  string `yaml:"priority"`
	Due       string `yaml:"due"`
	Assignee  string `yaml:"assignee"`
	CreatedBy string `yaml:"created_by"`
}

const dueDateLayout = "2006-01-02"

func tasksDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".kanban", "tasks")
}

func taskFilePath(repoRoot, taskID string) string {
	return filepath.Join(tasksDir(repoRoot), taskID+".md")
}

// Save writes the task to disk using an atomic .tmp → rename pattern.
// Returns an error wrapping os.ErrExist if the task file already exists.
func (r *TaskRepository) Save(repoRoot string, task domain.Task) error {
	dir := tasksDir(repoRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create tasks dir: %w", err)
	}

	finalPath := taskFilePath(repoRoot, task.ID)

	// Detect duplicate: if file already exists this is a create-collision.
	if _, err := os.Stat(finalPath); err == nil {
		return fmt.Errorf("task %s already exists: %w", task.ID, os.ErrExist)
	}

	content, err := marshalTask(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	return atomicOverwrite(finalPath, content)
}

// Update overwrites an existing task file atomically.
// Returns ErrTaskNotFound if no file exists for the task ID.
func (r *TaskRepository) Update(repoRoot string, task domain.Task) error {
	finalPath := taskFilePath(repoRoot, task.ID)
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		return ports.ErrTaskNotFound
	}

	content, err := marshalTask(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	return atomicOverwrite(finalPath, content)
}

// FindByID reads the task file for the given ID.
func (r *TaskRepository) FindByID(repoRoot, taskID string) (domain.Task, error) {
	path := taskFilePath(repoRoot, taskID)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("read task file: %w", err)
	}
	return unmarshalTask(data)
}

// ListAll reads every TASK-NNN.md file in the tasks directory.
func (r *TaskRepository) ListAll(repoRoot string) ([]domain.Task, error) {
	dir := tasksDir(repoRoot)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []domain.Task{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read tasks dir: %w", err)
	}

	tasks := make([]domain.Task, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		task, err := unmarshalTask(data)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// Delete removes the task file for the given ID.
func (r *TaskRepository) Delete(repoRoot, taskID string) error {
	path := taskFilePath(repoRoot, taskID)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return ports.ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("delete task file: %w", err)
	}
	return nil
}

// NextID scans existing TASK-NNN.md files and returns the next ID.
func (r *TaskRepository) NextID(repoRoot string) (string, error) {
	dir := tasksDir(repoRoot)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return "TASK-001", nil
	}
	if err != nil {
		return "", fmt.Errorf("read tasks dir: %w", err)
	}

	taskNumPattern := regexp.MustCompile(`^TASK-(\d+)\.md$`)
	maxNum := 0
	for _, e := range entries {
		matches := taskNumPattern.FindStringSubmatch(e.Name())
		if matches == nil {
			continue
		}
		n, _ := strconv.Atoi(matches[1])
		if n > maxNum {
			maxNum = n
		}
	}
	return fmt.Sprintf("TASK-%03d", maxNum+1), nil
}

// marshalTask serialises a Task to Markdown with YAML front matter.
func marshalTask(task domain.Task) ([]byte, error) {
	fm := taskFrontMatter{
		ID:        task.ID,
		Title:     task.Title,
		Status:    string(task.Status),
		Priority:  task.Priority,
		Assignee:  task.Assignee,
		CreatedBy: task.CreatedBy,
	}
	if task.Due != nil {
		fm.Due = task.Due.Format(dueDateLayout)
	}

	header, err := yaml.Marshal(fm)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(header)
	sb.WriteString("---\n")
	if task.Description != "" {
		sb.WriteString("\n")
		sb.WriteString(task.Description)
		sb.WriteString("\n")
	}
	return []byte(sb.String()), nil
}

// unmarshalTask parses Markdown+YAML front matter back into a Task.
func unmarshalTask(data []byte) (domain.Task, error) {
	content := string(data)

	// Expect file to start with "---\n"
	if !strings.HasPrefix(content, "---\n") {
		return domain.Task{}, fmt.Errorf("missing YAML front matter delimiter")
	}

	// Find the closing "---"
	rest := content[4:] // skip opening "---\n"
	parts := strings.SplitN(rest, "\n---\n", 2)
	if len(parts) < 2 {
		// Try "---" at end-of-file without trailing newline
		endIdx := strings.Index(rest, "\n---")
		if endIdx < 0 {
			return domain.Task{}, fmt.Errorf("unclosed YAML front matter")
		}
		parts = []string{rest[:endIdx], rest[endIdx+4:]}
	}

	yamlBlock := parts[0]
	body := strings.TrimPrefix(parts[1], "\n")
	body = strings.TrimSuffix(body, "\n")

	var fm taskFrontMatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return domain.Task{}, fmt.Errorf("parse front matter: %w", err)
	}

	task := domain.Task{
		ID:          fm.ID,
		Title:       fm.Title,
		Status:      domain.TaskStatus(fm.Status),
		Priority:    fm.Priority,
		Assignee:    fm.Assignee,
		CreatedBy:   fm.CreatedBy,
		Description: body,
	}

	if fm.Due != "" {
		t, err := time.Parse(dueDateLayout, fm.Due)
		if err != nil {
			return domain.Task{}, fmt.Errorf("parse due date %q: %w", fm.Due, err)
		}
		task.Due = &t
	}

	return task, nil
}

// atomicOverwrite writes content to a .tmp file then renames it to finalPath.
// Overwrites any existing file at finalPath atomically.
func atomicOverwrite(finalPath string, content []byte) error {
	tmpPath := finalPath + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0o644); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename tmp to final: %w", err)
	}
	return nil
}

// editFields is the YAML shape used for the temp edit file.
type editFields struct {
	Title       string `yaml:"title"`
	Priority    string `yaml:"priority"`
	Due         string `yaml:"due"`
	Assignee    string `yaml:"assignee"`
	Description string `yaml:"description"`
}

// WriteTemp writes the editable fields of task to a temporary YAML file.
// Returns the temp file path. The caller is responsible for removing the file.
func (r *TaskRepository) WriteTemp(task domain.Task) (string, error) {
	f, err := os.CreateTemp("", "kanban-edit-*.yaml")
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	due := ""
	if task.Due != nil {
		due = task.Due.Format("2006-01-02")
	}
	ef := editFields{
		Title:       task.Title,
		Priority:    task.Priority,
		Due:         due,
		Assignee:    task.Assignee,
		Description: task.Description,
	}
	data, err := yaml.Marshal(ef)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// ReadTemp reads the YAML temp file at path and returns an EditSnapshot.
func (r *TaskRepository) ReadTemp(path string) (ports.EditSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ports.EditSnapshot{}, err
	}
	var ef editFields
	if err := yaml.Unmarshal(data, &ef); err != nil {
		return ports.EditSnapshot{}, err
	}
	return ports.EditSnapshot{
		Title:       ef.Title,
		Priority:    ef.Priority,
		Due:         ef.Due,
		Assignee:    ef.Assignee,
		Description: ef.Description,
	}, nil
}

// ensure compile-time interface compliance
var _ ports.TaskRepository = (*TaskRepository)(nil)
var _ ports.EditFilePort = (*TaskRepository)(nil)
