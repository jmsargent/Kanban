package driver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// ServerDriver compiles and starts the kanban-web binary as a subprocess.
// It manages the full lifecycle: build, start, health check, stop.
type ServerDriver struct {
	t            *testing.T
	cmd          *exec.Cmd
	url          string
	binPath      string
	repoDir      string
	syncInterval time.Duration
	cookieKey    string
	githubAPIURL string
}

// NewServerDriver constructs a ServerDriver. The binary is not started until
// Start is called.
func NewServerDriver(t *testing.T) *ServerDriver {
	t.Helper()
	return &ServerDriver{
		t:            t,
		syncInterval: 30 * time.Second,
		cookieKey:    "test-cookie-key-32bytes-padded!!", // exactly 32 bytes for AES-256
	}
}

// SetRepoDir configures the git repository directory the server will serve.
func (d *ServerDriver) SetRepoDir(dir string) {
	d.repoDir = dir
}

// SetGitHubAPIURL configures the GitHub API URL for token validation.
func (d *ServerDriver) SetGitHubAPIURL(url string) {
	d.githubAPIURL = url
}

// URL returns the base URL of the running server.
func (d *ServerDriver) URL() string {
	return d.url
}

// projectRoot walks up from the driver source file to find the directory
// containing go.mod (the project root).
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

// Build compiles the kanban-web binary. If KANBAN_WEB_BIN is set, uses that
// path instead.
func (d *ServerDriver) Build() error {
	if bin := os.Getenv("KANBAN_WEB_BIN"); bin != "" {
		d.binPath = bin
		return nil
	}

	binDir := d.t.TempDir()
	d.binPath = filepath.Join(binDir, "kanban-web")

	buildCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(buildCtx, "go", "build", "-o", d.binPath, "./cmd/kanban-web")
	cmd.Dir = projectRoot()
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build kanban-web: %w", err)
	}
	return nil
}

// Start launches the kanban-web binary as a subprocess and waits for it to
// become healthy.
func (d *ServerDriver) Start(port int) error {
	if d.binPath == "" {
		if err := d.Build(); err != nil {
			return err
		}
	}

	d.url = fmt.Sprintf("http://localhost:%d", port)

	args := []string{
		"--port", fmt.Sprintf("%d", port),
		"--cookie-key", d.cookieKey,
	}
	if d.repoDir != "" {
		args = append(args, "--repo", d.repoDir)
	}

	d.cmd = exec.Command(d.binPath, args...)
	d.cmd.Stdout = os.Stdout
	d.cmd.Stderr = os.Stderr

	env := os.Environ()
	if d.githubAPIURL != "" {
		env = append(env, "KANBAN_WEB_GITHUB_API_URL="+d.githubAPIURL)
	}
	d.cmd.Env = env

	if err := d.cmd.Start(); err != nil {
		return fmt.Errorf("start kanban-web: %w", err)
	}

	d.t.Cleanup(func() {
		d.Stop()
	})

	return d.waitForHealthy(10 * time.Second)
}

// Stop terminates the server subprocess.
func (d *ServerDriver) Stop() {
	if d.cmd != nil && d.cmd.Process != nil {
		_ = d.cmd.Process.Kill()
		_ = d.cmd.Wait()
	}
}

// waitForHealthy polls /healthz until it returns 200 or the timeout expires.
func (d *ServerDriver) waitForHealthy(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 1 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(d.url + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server did not become healthy within %s", timeout)
}
