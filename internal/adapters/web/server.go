package web

import (
	"fmt"
	"net/http"
)

// Server wires HTTP routes and manages the server lifecycle.
type Server struct {
	addr string
	mux  *http.ServeMux
}

// NewServer constructs a Server listening on addr with the given providers.
func NewServer(addr string, getBoard BoardProvider, getTask TaskProvider) *Server {
	mux := http.NewServeMux()
	s := &Server{addr: addr, mux: mux}
	s.registerRoutes(getBoard, getTask)
	return s
}

// registerRoutes registers all HTTP routes on the mux.
func (s *Server) registerRoutes(getBoard BoardProvider, getTask TaskProvider) {
	s.mux.Handle("/board", NewBoardHandler(getBoard))
	s.mux.Handle("/card/{id}", NewCardDetailHandler(getTask))
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}
