package usecases

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// TaskDiff holds the before and after snapshots of a task edit plus the list
// of field names that changed.
type TaskDiff struct {
	Before        domain.Task
	After         domain.Task
	ChangedFields []string
	NoChanges     bool
}

// editFields is the editable subset of task fields written to the temp file.
type editFields struct {
	Title       string `yaml:"title"`
	Priority    string `yaml:"priority"`
	Due         string `yaml:"due"`
	Assignee    string `yaml:"assignee"`
	Description string `yaml:"description"`
}

// EditTask implements the edit-task use case.
// Driving port entrypoint for "kanban edit".
type EditTask struct {
	tasks ports.TaskRepository
}

// NewEditTask constructs an EditTask use case.
func NewEditTask(tasks ports.TaskRepository) *EditTask {
	return &EditTask{tasks: tasks}
}

// Execute finds the task, opens $EDITOR with its editable fields, parses the
// result, and persists changed fields via TaskRepository.Update.
// Returns ErrTaskNotFound when the task does not exist.
// Returns TaskDiff.NoChanges == true when the editor made no changes.
func (u *EditTask) Execute(repoRoot, taskID string) (TaskDiff, error) {
	task, err := u.tasks.FindByID(repoRoot, taskID)
	if err != nil {
		return TaskDiff{}, err
	}

	before := task

	// Write editable fields to a temp file.
	tmpFile, err := writeTempEditFile(task)
	if err != nil {
		return TaskDiff{}, fmt.Errorf("write temp edit file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Open editor.
	if err := openEditor(tmpFile); err != nil {
		return TaskDiff{}, fmt.Errorf("open editor: %w", err)
	}

	// Parse updated fields from temp file.
	updated, err := readTempEditFile(tmpFile)
	if err != nil {
		return TaskDiff{}, fmt.Errorf("read temp edit file: %w", err)
	}

	// Apply parsed fields to the task copy.
	after := applyEditFields(task, updated)

	// Detect changes.
	changed := detectChangedFields(before, after)
	if len(changed) == 0 {
		return TaskDiff{Before: before, After: after, NoChanges: true}, nil
	}

	// Validate changed title.
	if after.Title == "" {
		return TaskDiff{}, fmt.Errorf("%w: title cannot be empty", ports.ErrInvalidInput)
	}

	if err := u.tasks.Update(repoRoot, after); err != nil {
		return TaskDiff{}, fmt.Errorf("update task: %w", err)
	}

	return TaskDiff{Before: before, After: after, ChangedFields: changed}, nil
}

func writeTempEditFile(task domain.Task) (string, error) {
	f, err := os.CreateTemp("", "kanban-edit-*.yaml")
	if err != nil {
		return "", err
	}
	defer f.Close()

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

func openEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func readTempEditFile(filePath string) (editFields, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return editFields{}, err
	}
	var ef editFields
	if err := yaml.Unmarshal(data, &ef); err != nil {
		return editFields{}, err
	}
	return ef, nil
}

func applyEditFields(task domain.Task, ef editFields) domain.Task {
	task.Title = strings.TrimSpace(ef.Title)
	task.Priority = ef.Priority
	task.Assignee = ef.Assignee
	task.Description = ef.Description
	return task
}

func detectChangedFields(before, after domain.Task) []string {
	var changed []string
	if before.Title != after.Title {
		changed = append(changed, "title")
	}
	if before.Priority != after.Priority {
		changed = append(changed, "priority")
	}
	if before.Assignee != after.Assignee {
		changed = append(changed, "assignee")
	}
	if before.Description != after.Description {
		changed = append(changed, "description")
	}
	return changed
}
