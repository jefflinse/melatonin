package mt

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// An HTTPTestContext is used to create HTTP test cases that target either
// a specific base URL or a Go HTTP handler.
type HTTPTestContext struct {
	BaseURL string
	Client  *http.Client
	Handler http.Handler
}

// DefaultContext returns an HTTPTestContext using the default HTTP client.
func DefaultContext() *HTTPTestContext {
	return &HTTPTestContext{}
}

// NewURLContext creates a new HTTPTestContext for creating tests that target
// the specified base URL.
func NewURLContext(baseURL string) *HTTPTestContext {
	return &HTTPTestContext{
		BaseURL: baseURL,
	}
}

// NewHandlerContext creates a new HTTPTestContext for creating tests that target
// the specified HTTP handler.
func NewHandlerContext(handler http.Handler) *HTTPTestContext {
	return &HTTPTestContext{
		Handler: handler,
	}
}

// WithHTTPClient sets the HTTP client used for HTTP requests and returns the
// context.
func (c *HTTPTestContext) WithHTTPClient(client *http.Client) *HTTPTestContext {
	c.Client = client
	return c
}

// DELETE is a shortcut for NewTestCase(http.MethodDelete, path).
func (c *HTTPTestContext) DELETE(path string, description ...string) *HTTPTestCase {
	return c.newHTTPTestCase(http.MethodDelete, path, description...)
}

// HEAD is a shortcut for NewTestCase(http.MethodHead, path, description...).
func (c *HTTPTestContext) HEAD(path string, description ...string) *HTTPTestCase {
	return c.newHTTPTestCase(http.MethodHead, path, description...)
}

// GET is a shortcut for NewTestCase(http.MethodGet, path, description...).
func (c *HTTPTestContext) GET(path string, description ...string) *HTTPTestCase {
	return c.newHTTPTestCase(http.MethodGet, path, description...)
}

// OPTIONS is a shortcut for NewTestCase(http.MethodOptions, path, description...).
func (c *HTTPTestContext) OPTIONS(path string, description ...string) *HTTPTestCase {
	return c.newHTTPTestCase(http.MethodOptions, path, description...)
}

// PATCH is a shortcut for NewTestCase(http.MethodPatch, path, description...).
func (c *HTTPTestContext) PATCH(path string, description ...string) *HTTPTestCase {
	return c.newHTTPTestCase(http.MethodPatch, path, description...)
}

// POST is a shortcut for NewTestCase(http.MethodPost, path, description...).
func (c *HTTPTestContext) POST(path string, description ...string) *HTTPTestCase {
	return c.newHTTPTestCase(http.MethodPost, path, description...)
}

// PUT is a shortcut for NewTestCase(http.MethodPut, path, description...).
func (c *HTTPTestContext) PUT(path string, description ...string) *HTTPTestCase {
	return c.newHTTPTestCase(http.MethodPut, path, description...)
}

// DO creates a test case from a custom HTTP request.
func (c *HTTPTestContext) DO(request *http.Request, description ...string) *HTTPTestCase {
	tc := c.newHTTPTestCase(request.Method, request.URL.Path, description...)
	tc.request = request
	return tc
}

func (c *HTTPTestContext) createURL(path string) (*url.URL, error) {
	if path == "" {
		return nil, errors.New("not enough URL information")
	}

	// when using the default context, the path must be a complete URL.
	if c.BaseURL == "" {
		u, err := url.ParseRequestURI(path)
		if err != nil {
			return nil, fmt.Errorf("invalid URL %q: %s", path, err)
		}

		return u, nil
	}

	base, err := url.ParseRequestURI(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL %q: %s", c.BaseURL, err)
	}

	endpoint, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path %q: %s", path, err)
	}

	return base.ResolveReference(endpoint), nil
}

func (c *HTTPTestContext) newHTTPTestCase(method, path string, description ...string) *HTTPTestCase {
	u, err := c.createURL(path)
	if err != nil {
		log.Fatalf("failed to create URL for path %q: %v", path, err)
	}

	req, cancel, err := createRequest(method, u.String())
	if err != nil {
		log.Fatalf("failed to create request %v", err)
	}

	return &HTTPTestCase{
		Desc:    strings.Join(description, " "),
		tctx:    c,
		request: req,
		cancel:  cancel,
	}
}
