package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmsargent/kanban/internal/adapters/web"
	"github.com/jmsargent/kanban/internal/domain"
)

// Test Budget: 3 behaviors × 2 = 6 max unit tests. Using 3.
// Behavior 1: GET /board returns 200 with non-empty HTML body.
// Behavior 2: BoardHandler renders tasks in correct columns (Todo/Doing/Done).
// Behavior 3: BoardHandler renders three empty columns when no tasks exist.

func TestBoardHandler_Returns200WithHTMLBody(t *testing.T) {
	handler := web.NewBoardHandler(web.StaticBoardProvider(domain.Board{}))

	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.TrimSpace(body) == "" {
		t.Fatal("expected non-empty response body")
	}
}

func TestBoardHandler_RendersTasksInCorrectColumns(t *testing.T) {
	board := domain.Board{
		Columns: []domain.Column{
			{Name: "todo", Label: "Todo"},
			{Name: "in-progress", Label: "Doing"},
			{Name: "done", Label: "Done"},
		},
		Tasks: map[domain.TaskStatus][]domain.Task{
			domain.StatusTodo:       {{ID: "TASK-002", Title: "Write docs", Status: domain.StatusTodo}},
			domain.StatusInProgress: {{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusInProgress}},
			domain.StatusDone:       {{ID: "TASK-003", Title: "Deploy v1", Status: domain.StatusDone}},
		},
	}

	handler := web.NewBoardHandler(web.StaticBoardProvider(board))
	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assertColumnContains(t, body, "Todo", "Write docs")
	assertColumnContains(t, body, "Doing", "Fix login bug")
	assertColumnContains(t, body, "Done", "Deploy v1")
}

func TestBoardHandler_RendersEmptyColumns(t *testing.T) {
	board := domain.Board{
		Columns: []domain.Column{
			{Name: "todo", Label: "Todo"},
			{Name: "in-progress", Label: "Doing"},
			{Name: "done", Label: "Done"},
		},
		Tasks: map[domain.TaskStatus][]domain.Task{},
	}

	handler := web.NewBoardHandler(web.StaticBoardProvider(board))
	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, col := range []string{"Todo", "Doing", "Done"} {
		marker := fmt.Sprintf(`data-column="%s"`, col)
		if !strings.Contains(body, marker) {
			t.Errorf("expected column %q in HTML", col)
		}
	}
}

// assertColumnContains checks that the named column in body contains a card
// with the given title.
func assertColumnContains(t *testing.T, body, column, title string) {
	t.Helper()
	colMarker := fmt.Sprintf(`data-column="%s"`, column)
	colIdx := strings.Index(body, colMarker)
	if colIdx < 0 {
		t.Errorf("column %q not found in HTML", column)
		return
	}
	section := extractSection(body, colIdx)
	cardMarker := fmt.Sprintf(`data-title="%s"`, title)
	if !strings.Contains(section, cardMarker) {
		t.Errorf("card %q not found in column %q\nSection:\n%s", title, column, section)
	}
}

func extractSection(html string, startIdx int) string {
	openTag := startIdx
	for openTag > 0 && html[openTag] != '<' {
		openTag--
	}
	depth := 0
	i := openTag
	n := len(html)
	for i < n {
		if strings.HasPrefix(html[i:], "<div") {
			depth++
			i += 4
			continue
		}
		if strings.HasPrefix(html[i:], "</div>") {
			depth--
			i += 6
			if depth == 0 {
				return html[openTag:i]
			}
			continue
		}
		i++
	}
	return html[openTag:]
}
