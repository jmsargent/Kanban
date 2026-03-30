package driver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// GitHubStubDriver runs a local httptest server that stubs GitHub API endpoints
// for token validation. Tests configure which tokens are valid and what user
// info they return.
type GitHubStubDriver struct {
	server      *httptest.Server
	mu          sync.RWMutex
	validTokens map[string]githubUser
}

type githubUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

// NewGitHubStubDriver creates and starts a stub GitHub API server.
func NewGitHubStubDriver() *GitHubStubDriver {
	d := &GitHubStubDriver{
		validTokens: make(map[string]githubUser),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /user", d.handleGetUser)
	d.server = httptest.NewServer(mux)

	return d
}

// URL returns the base URL of the stub server for injection via
// KANBAN_WEB_GITHUB_API_URL environment variable.
func (d *GitHubStubDriver) URL() string {
	return d.server.URL
}

// Close shuts down the stub server.
func (d *GitHubStubDriver) Close() {
	d.server.Close()
}

// AddValidToken registers a token that will be accepted by the stub.
// When a request arrives with "Bearer <token>", GET /user returns the
// configured user info.
func (d *GitHubStubDriver) AddValidToken(token, login, displayName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.validTokens[token] = githubUser{Login: login, Name: displayName}
}

// handleGetUser implements GET /user. It checks the Authorization header for
// a valid Bearer token and returns user info or 401.
func (d *GitHubStubDriver) handleGetUser(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, `{"message":"Requires authentication"}`, http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(auth, "Bearer ")

	d.mu.RLock()
	user, ok := d.validTokens[token]
	d.mu.RUnlock()

	if !ok {
		http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
