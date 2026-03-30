package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmsargent/kanban/internal/adapters/web"
)

// Test Budget: 1 behavior × 2 = 2 max unit tests. Using 1.
// Behavior: GET /board returns 200 with non-empty HTML body.

func TestBoardHandler_Returns200WithHTMLBody(t *testing.T) {
	handler := web.NewBoardHandler()

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
