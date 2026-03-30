package web

import (
	"fmt"
	"net/http"
)

// BoardHandler serves the kanban board page at GET /board.
type BoardHandler struct{}

// NewBoardHandler constructs a BoardHandler.
func NewBoardHandler() *BoardHandler {
	return &BoardHandler{}
}

// ServeHTTP handles GET /board requests, returning a minimal HTML board page.
func (h *BoardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `<!DOCTYPE html>
<html>
<head><title>Kanban Board</title></head>
<body><h1>Kanban Board</h1></body>
</html>`)
}
