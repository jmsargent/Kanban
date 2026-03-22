package usecases

import (
	"fmt"
	"os"
	"strings"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// NewEditorTask implements the interactive "kanban new" use case where no title
// argument is provided and the user fills in a blank template via $EDITOR.
type NewEditorTask struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
	editor ports.EditFilePort
}

// NewEditorTaskUseCase constructs a NewEditorTask use case.
func NewEditorTaskUseCase(config ports.ConfigRepository, tasks ports.TaskRepository, editor ports.EditFilePort) *NewEditorTask {
	return &NewEditorTask{config: config, tasks: tasks, editor: editor}
}

// Execute opens $EDITOR with a blank task template, parses the saved result,
// and persists the task.
// Returns ErrNotInitialised when kanban has not been initialised.
// Returns ErrInvalidInput when the title is empty after the editor closes.
func (u *NewEditorTask) Execute(repoRoot, createdBy string) (domain.Task, error) {
	if _, err := u.config.Read(repoRoot); err != nil {
		return domain.Task{}, fmt.Errorf("read config: %w", err)
	}

	tmpFile, err := u.editor.WriteTempNew()
	if err != nil {
		return domain.Task{}, fmt.Errorf("write temp new file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	if err := OpenEditor(resolveEditor(), tmpFile); err != nil {
		return domain.Task{}, fmt.Errorf("open editor: %w", err)
	}

	snapshot, err := u.editor.ReadTemp(tmpFile)
	if err != nil {
		return domain.Task{}, fmt.Errorf("read temp file: %w", err)
	}

	title := strings.TrimSpace(snapshot.Title)
	if title == "" {
		return domain.Task{}, fmt.Errorf("%w: title cannot be empty", ports.ErrInvalidInput)
	}

	id, err := u.tasks.NextID(repoRoot)
	if err != nil {
		return domain.Task{}, fmt.Errorf("next id: %w", err)
	}

	task := domain.Task{
		ID:        id,
		Title:     title,
		Status:    domain.StatusTodo,
		Priority:  snapshot.Priority,
		Assignee:  snapshot.Assignee,
		CreatedBy: createdBy,
	}

	if err := u.tasks.Save(repoRoot, task); err != nil {
		return domain.Task{}, fmt.Errorf("save task: %w", err)
	}

	return task, nil
}
