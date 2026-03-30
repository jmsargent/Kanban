package driver

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// HTTPDriver wraps net/http.Client with a cookie jar for stateful HTTP
// interactions with the kanban-web server.
type HTTPDriver struct {
	client  *http.Client
	baseURL string
}

// NewHTTPDriver constructs an HTTPDriver targeting the given base URL.
// It configures a cookie jar for automatic cookie handling.
func NewHTTPDriver(baseURL string) *HTTPDriver {
	jar, _ := cookiejar.New(nil)
	return &HTTPDriver{
		client: &http.Client{
			Jar: jar,
		},
		baseURL: baseURL,
	}
}

// Response holds the result of an HTTP request.
type Response struct {
	StatusCode int
	Body       string
	Headers    http.Header
}

// GET performs an HTTP GET to the given path and returns the response.
func (d *HTTPDriver) GET(path string) (*Response, error) {
	resp, err := d.client.Get(d.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Headers:    resp.Header,
	}, nil
}

// POST performs an HTTP POST with form data to the given path.
func (d *HTTPDriver) POST(path string, formData url.Values) (*Response, error) {
	resp, err := d.client.Post(
		d.baseURL+path,
		"application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Headers:    resp.Header,
	}, nil
}

// Cookies returns the cookies stored for the server URL.
func (d *HTTPDriver) Cookies() []*http.Cookie {
	u, _ := url.Parse(d.baseURL)
	return d.client.Jar.Cookies(u)
}
