package cli

import (
	"errors"
	"os"
	"strings"

	"github.com/kanban-tasks/kanban/internal/domain"
)

// ErrNoKanbanBlock is returned by writeMermaidToFile when the target file exists
// but contains no fenced Mermaid kanban block to replace.
var ErrNoKanbanBlock = errors.New("no mermaid kanban block found")

// writeMermaidToFile writes content to filename atomically (write to .tmp then rename).
// If filename does not exist, the file is created with the given content.
// If filename exists, it is scanned for a fenced Mermaid kanban block:
//   - No block found → ErrNoKanbanBlock is returned; the file is not modified.
//   - Block found    → the block is replaced in-place and nil is returned.
func writeMermaidToFile(filename, content string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		tmp := filename + ".tmp"
		if writeErr := os.WriteFile(tmp, []byte(content), 0644); writeErr != nil {
			return writeErr
		}
		return os.Rename(tmp, filename)
	}
	// File exists — scan for a fenced Mermaid kanban block.
	data, readErr := os.ReadFile(filename)
	if readErr != nil {
		return readErr
	}
	lines := strings.Split(string(data), "\n")
	inFence := false
	hasKanban := false
	fenceStart := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inFence && trimmed == "```mermaid" {
			inFence = true
			fenceStart = i
			continue
		}
		if inFence && trimmed == "kanban" {
			hasKanban = true
			continue
		}
		if inFence && trimmed == "```" && !hasKanban {
			// Non-kanban mermaid fence closed — reset and keep scanning.
			inFence = false
			fenceStart = -1
			continue
		}
		if inFence && trimmed == "```" && hasKanban {
			// Reconstruct: lines before fence + new content + lines after closing fence.
			var parts []string
			if before := strings.Join(lines[:fenceStart], "\n"); before != "" {
				parts = append(parts, before)
			}
			parts = append(parts, content)
			after := strings.Join(lines[i+1:], "\n")
			parts = append(parts, after)
			reconstructed := strings.Join(parts, "\n")
			// Ensure a single trailing newline (after is "" when file ends with \n).
			if !strings.HasSuffix(reconstructed, "\n") {
				reconstructed += "\n"
			}
			tmp := filename + ".tmp"
			if writeErr := os.WriteFile(tmp, []byte(reconstructed), 0o644); writeErr != nil {
				return writeErr
			}
			return os.Rename(tmp, filename)
		}
	}
	return ErrNoKanbanBlock
}

// mermaidTitleReplacer replaces characters that are unsafe in Mermaid node labels.
var mermaidTitleReplacer = strings.NewReplacer(
	`"`, `'`,
	`[`, `(`,
	`]`, `)`,
	"\n", " ",
	"\r", " ",
)

// sanitiseMermaidTitle replaces Mermaid-unsafe characters in a task title with safe substitutes.
func sanitiseMermaidTitle(s string) string {
	return mermaidTitleReplacer.Replace(s)
}

// mermaidLabelReplacer replaces characters that are unsafe in Mermaid section labels.
var mermaidLabelReplacer = strings.NewReplacer(
	`:`, ` `,
	`"`, `'`,
	`[`, `(`,
	`]`, `)`,
	"\n", " ",
	"\r", " ",
)

// sanitiseMermaidLabel replaces Mermaid-unsafe characters in a column label with safe substitutes.
func sanitiseMermaidLabel(s string) string {
	return mermaidLabelReplacer.Replace(s)
}

// renderBoardMermaid returns a fenced Mermaid kanban block representing the board.
func renderBoardMermaid(board domain.Board) string {
	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("kanban\n")
	for _, col := range board.Columns {
		sb.WriteString("  section ")
		sb.WriteString(sanitiseMermaidLabel(col.Label))
		sb.WriteString("\n")
		status := domain.TaskStatus(col.Name)
		for _, task := range board.Tasks[status] {
			sb.WriteString("    ")
			sb.WriteString(task.ID)
			sb.WriteString("@{ label: \"")
			sb.WriteString(sanitiseMermaidTitle(task.Title))
			sb.WriteString("\" }\n")
		}
	}
	sb.WriteString("```\n")
	return sb.String()
}
