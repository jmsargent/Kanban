package web

import (
	"fmt"
	"net/http"
)

// Server wires HTTP routes and manages the server lifecycle.
type Server struct {
	addr             string
	mux              *http.ServeMux
	sessionKey       []byte
	githubAPIBaseURL string
}

// NewServer constructs a Server listening on addr with the given providers.
// sessionKey must be 32 bytes for AES-256-GCM cookie encryption. If nil, a
// zero key is used (suitable for dev/test only; not secure).
// githubAPIBaseURL is the GitHub API base URL for token validation (e.g.
// "https://api.github.com"). Pass an empty string to use the default.
func NewServer(addr string, getBoard BoardProvider, getTask TaskProvider, sessionKey []byte, githubAPIBaseURL string) *Server {
	if sessionKey == nil {
		sessionKey = make([]byte, 32) // zero key — insecure dev default
	}
	if githubAPIBaseURL == "" {
		githubAPIBaseURL = "https://api.github.com"
	}
	mux := http.NewServeMux()
	s := &Server{addr: addr, mux: mux, sessionKey: sessionKey, githubAPIBaseURL: githubAPIBaseURL}
	s.registerRoutes(getBoard, getTask)
	return s
}

// registerRoutes registers all HTTP routes on the mux.
func (s *Server) registerRoutes(getBoard BoardProvider, getTask TaskProvider) {
	// Public read routes — no auth required.
	s.mux.Handle("/board", NewBoardHandler(getBoard))
	s.mux.Handle("/card/{id}", NewCardDetailHandler(getTask))

	// Auth routes — GET renders form, POST processes submission.
	s.mux.Handle("GET /auth/token", NewTokenEntryHandler())
	s.mux.Handle("POST /auth/token", NewTokenSubmitHandler(s.sessionKey, s.githubAPIBaseURL))

	// Write routes — auth required.
	s.mux.Handle("/task/new", RequireAuth(s.sessionKey, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Placeholder: task creation form (implemented in a later step).
		http.Error(w, "not implemented", http.StatusNotImplemented)
	})))

	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}
