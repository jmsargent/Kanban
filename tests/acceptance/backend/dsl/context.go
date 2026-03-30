package dsl

import (
	"net/http"
	"testing"
	"time"
)

// WebContext holds the mutable state for a single web acceptance test scenario.
// It mirrors the CLI Context but targets the web binary instead.
type WebContext struct {
	T            *testing.T
	ServerURL    string
	HTTPClient   *http.Client
	Cookies      []*http.Cookie
	LastResponse *http.Response
	LastBody     string
	LastDuration time.Duration
	RepoDir      string
	RemoteDir    string // bare remote path; set by ARepoWithRemote
}

// NewWebContext constructs a WebContext for a test. The server is not started
// automatically -- use setup steps to configure the repo and start the server.
func NewWebContext(t *testing.T) *WebContext {
	t.Helper()
	return &WebContext{
		T:          t,
		HTTPClient: &http.Client{},
	}
}
