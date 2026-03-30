package dsl

import (
	"net/http"
	"testing"
	"time"

	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

// WebContext holds the mutable state for a single web acceptance test scenario.
// It mirrors the CLI Context but targets the web binary instead.
type WebContext struct {
	T             *testing.T
	ServerURL     string
	HTTPClient    *http.Client
	HTTPDriver    *driver.HTTPDriver // stateful driver with cookie jar; set on first authenticated request
	Cookies       []*http.Cookie
	LastResponse  *http.Response
	LastBody      string
	LastDuration  time.Duration
	RepoDir       string
	RemoteDir     string // bare remote path; set by ARepoWithRemote
	GitHubStubURL string // base URL of the GitHub API stub; set by WithGitHubStub
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
