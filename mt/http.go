package mt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jefflinse/melatonin/expect"
	"github.com/jefflinse/melatonin/golden"
)

const (
	defaultRequestTimeoutStr = "10s"
)

var (
	defaultRequestTimeout time.Duration
)

func init() {
	defaultRequestTimeout, _ = time.ParseDuration(defaultRequestTimeoutStr)
	envTimeoutStr := os.Getenv("MELATONIN_DEFAULT_TEST_TIMEOUT")
	if envTimeoutStr != "" {
		if timeout, err := time.ParseDuration(envTimeoutStr); err == nil {
			defaultRequestTimeout = timeout
		} else {
			fmt.Printf("invalid MELATONIN_DEFAULT_TEST_TIMEOUT value %q in environment, using default of %s\n",
				envTimeoutStr, defaultRequestTimeoutStr)
		}
	}
}

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

// An HTTPTestCase tests a single call to an HTTP endpoint.
//
// An optional setup function can be provided to perform any necessary
// setup before the test is run, such as adding or removing objects in
// a database.
//
// All fields in the WantBody map are expected to be present in the
// response body.
type HTTPTestCase struct {
	// After is an optional function that is run after the test is run.
	// It can be used to perform any cleanup actions after the test,
	// such as adding or removing objects in a database. Any error
	// returned by After is treated as a test failure.
	AfterFunc func() error

	// Before is an optional function that is run before the test is run.
	// It can be used to perform any prerequisites actions for the test,
	// such as adding or removing objects in a database. Any error
	// returned by Before is treated as a test failure.
	BeforeFunc func() error

	// Desc is a description of the test case.
	Desc string

	// Expectations is a set of values to compare the response against.
	Expectations HTTPResponseExpectations

	// GoldenFilePath is a path to a golden file defining expectations for the test case.
	//
	// If set, any WantStatus, WantHeaders, or WantBody values are overriden with
	// values from the golden file.
	GoldenFilePath string

	// Configuration for the test
	tctx *HTTPTestContext

	// Underlying HTTP request for the test case.
	request *http.Request

	// Cancel function for the underlying HTTP request.
	cancel context.CancelFunc
}

var _ TestCase = &HTTPTestCase{}

func (tc *HTTPTestCase) Action() string {
	return strings.ToUpper(tc.request.Method)
}

func (tc *HTTPTestCase) Target() string {
	return tc.request.URL.Path
}

func (tc *HTTPTestCase) Description() string {
	if tc.Desc != "" {
		return tc.Desc
	}

	return fmt.Sprintf("%s %s (%d q, %d h)",
		tc.Action(), tc.Target(),
		len(tc.request.URL.Query()),
		len(tc.request.Header),
	)
}

func (tc *HTTPTestCase) Execute() TestResult {
	defer tc.cancel()

	result := &HTTPTestCaseResult{
		testCase: tc,
	}

	if tc.BeforeFunc != nil {
		if err := tc.BeforeFunc(); err != nil {
			return result.addErrors(fmt.Errorf("before(): %w", err))
		}
	}

	var err error
	if tc.tctx.Handler != nil {
		result.Status, result.Headers, result.Body, err = handleRequest(tc.tctx.Handler, tc.request)
		if err != nil {
			return result.addErrors(fmt.Errorf("failed to handle HTTP request: %w", err))
		}
	} else {
		if tc.tctx.Client == nil {
			tc.tctx.Client = http.DefaultClient
		}

		result.Status, result.Headers, result.Body, err = doRequest(tc.tctx.Client, tc.request)
		if err != nil {
			return result.addErrors(fmt.Errorf("failed to execute HTTP request: %w", err))
		}
	}

	result.validateExpectations()

	if tc.AfterFunc != nil {
		if err := tc.AfterFunc(); err != nil {
			result.addErrors(fmt.Errorf("after(): %w", err))
		}
	}

	return result
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

// After registers a function to be run after the test case.
func (tc *HTTPTestCase) After(after func() error) *HTTPTestCase {
	tc.AfterFunc = after
	return tc
}

// Before registers a function to be run before the test case.
func (tc *HTTPTestCase) Before(before func() error) *HTTPTestCase {
	tc.BeforeFunc = before
	return tc
}

// Describe sets a description for the test case.
func (tc *HTTPTestCase) Describe(description string) *HTTPTestCase {
	tc.Desc = description
	return tc
}

// WithBody sets the request body for the test case.
func (tc *HTTPTestCase) WithBody(body interface{}) *HTTPTestCase {
	b, err := toBytes(body)
	if err != nil {
		log.Fatalf("failed to marshal request body: %s", err)
	}

	tc.request.Body = io.NopCloser(bytes.NewReader(b))
	return tc
}

// WithHeaders sets the request headers for the test case.
func (tc *HTTPTestCase) WithHeaders(headers http.Header) *HTTPTestCase {
	tc.request.Header = headers
	return tc
}

// WithHeader adds a request header to the test case.
func (tc *HTTPTestCase) WithHeader(key, value string) *HTTPTestCase {
	tc.request.Header.Set(key, value)
	return tc
}

func (tc *HTTPTestCase) WithQueryParams(params url.Values) *HTTPTestCase {
	tc.request.URL.RawQuery = params.Encode()
	return tc
}

func (tc *HTTPTestCase) WithQueryParam(key, value string) *HTTPTestCase {
	q := tc.request.URL.Query()
	q.Add(key, value)
	tc.request.URL.RawQuery = q.Encode()
	return tc
}

// WithTimeout sets a timeout for the test case.
func (tc *HTTPTestCase) WithTimeout(timeout time.Duration) *HTTPTestCase {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	tc.request = tc.request.WithContext(ctx)
	tc.cancel = cancel
	return tc
}

// ExpectStatus sets the expected HTTP status code for the test case.
func (tc *HTTPTestCase) ExpectStatus(status int) *HTTPTestCase {
	tc.Expectations.Status = status
	return tc
}

// ExpectExactHeaders sets the expected HTTP response headers for the test case.
//
// Unlike ExpectHeaders, ExpectExactHeaders willl cause the test case to fail
// if any unexpected headers are present in the response.
func (tc *HTTPTestCase) ExpectExactHeaders(headers http.Header) *HTTPTestCase {
	tc.Expectations.WantExactHeaders = true
	return tc.ExpectHeaders(headers)
}

// ExpectHeaders sets the expected HTTP response headers for the test case.
//
// Unlike ExpectExactHeaders, ExpectHeaders only verifies that the expected
// headers are present in the response, and ignores any additional headers.
func (tc *HTTPTestCase) ExpectHeaders(headers http.Header) *HTTPTestCase {
	tc.Expectations.Headers = headers
	return tc
}

// ExpectHeader adds an expected HTTP response header for the test case.
func (tc *HTTPTestCase) ExpectHeader(key, value string) *HTTPTestCase {
	if tc.Expectations.Headers == nil {
		tc.Expectations.Headers = http.Header{}
	}

	tc.Expectations.Headers.Set(key, value)
	return tc
}

// ExpectBody sets the expected HTTP response body for the test case.
func (tc *HTTPTestCase) ExpectBody(body interface{}) *HTTPTestCase {
	tc.Expectations.Body = body
	return tc
}

// ExpectExactBody sets the expected HTTP response body for the test case.
//
// Unlike ExpectBody, ExpectExactBody willl cause the test case to fail
// if the expected response body is a JSON object or array and contains any
// additional fields or values not present in the expected JSON content.
//
// For non-JSON values, ExpectExactBody behaves identically to ExpectBody.
func (tc *HTTPTestCase) ExpectExactBody(body interface{}) *HTTPTestCase {
	tc.Expectations.WantExactJSONBody = true
	return tc.ExpectBody(body)
}

func (tc *HTTPTestCase) ExpectGolden(path string) *HTTPTestCase {
	tc.GoldenFilePath = path
	return tc
}

// Validate ensures that the test case is valid can can be run.
func (tc *HTTPTestCase) Validate() error {
	if tc.tctx.BaseURL != "" && tc.tctx.Handler != nil {
		return fmt.Errorf("HTTP test context %q cannot specify both a base URL and handler", tc.tctx.BaseURL)
	}

	if tc.GoldenFilePath != "" {
		path := tc.GoldenFilePath
		if !filepath.IsAbs(path) {
			path = filepath.Join(cfg.WorkingDir, path)
		}

		golden, err := golden.LoadFile(path)
		if err != nil {
			return err
		}

		tc.Expectations.Status = golden.WantStatus
		tc.Expectations.Headers = golden.WantHeaders
		tc.Expectations.Body = golden.WantBody
		tc.Expectations.WantExactHeaders = golden.MatchHeadersExactly
		tc.Expectations.WantExactJSONBody = golden.MatchBodyJSONExactly
	}

	return nil
}

// HTTPTestCaseResult represents the result of running a single test case.
type HTTPTestCaseResult struct {
	Status  int
	Headers http.Header
	Body    []byte

	testCase *HTTPTestCase
	errors   []error
}

func (r *HTTPTestCaseResult) Errors() []error {
	return r.errors
}

func (r *HTTPTestCaseResult) TestCase() TestCase {
	return r.testCase
}

func (r *HTTPTestCaseResult) addErrors(errs ...error) *HTTPTestCaseResult {
	if len(errs) == 0 {
		return r
	}

	r.errors = append(r.errors, errs...)
	return r
}

func (r *HTTPTestCaseResult) validateExpectations() {
	tc := r.TestCase().(*HTTPTestCase)
	if tc.Expectations.Status != 0 {
		if err := expect.Status(tc.Expectations.Status, r.Status); err != nil {
			r.addErrors(err)
		}
	}

	if tc.Expectations.Headers != nil {
		if errs := expect.Headers(tc.Expectations.Headers, r.Headers); len(errs) > 0 {
			r.addErrors(errs...)
		}
	}

	if tc.Expectations.Body != nil {
		body := toInterface(r.Body)
		if errs := expect.Value("body", tc.Expectations.Body, body, tc.Expectations.WantExactJSONBody); len(errs) > 0 {
			r.addErrors(errs...)
		}
	}
}

// HTTPResponseExpectations represents the expected values for single HTTP response.
type HTTPResponseExpectations struct {
	// Body is the expected HTTP response body content.
	Body interface{}

	// ExactHeaders indicates whether or not any unexpected response headers
	// should be treated as a test failure.
	WantExactHeaders bool

	// ExactJSONBody indicates whether or not the expected JSON should be matched
	// exactly (true) or treated as a subset of the response JSON (false).
	WantExactJSONBody bool

	// Headers is a map of HTTP headers that are expected to be present in
	// the HTTP response.
	Headers http.Header

	// Status is the expected HTTP status code of the response. Default is 200.
	Status int
}

// DELETE is a shortcut for DefaultContext().NewTestCase(http.MethodDelete, path).
func DELETE(url string, description ...string) *HTTPTestCase {
	return DefaultContext().DELETE(url, description...)
}

// HEAD is a shortcut for NewTestCase(http.MethodHead, path, description...).
func HEAD(url string, description ...string) *HTTPTestCase {
	return DefaultContext().HEAD(url, description...)
}

// GET is a shortcut for NewTestCase(http.MethodGet, path, description...).
func GET(url string, description ...string) *HTTPTestCase {
	return DefaultContext().GET(url, description...)
}

// OPTIONS is a shortcut for NewTestCase(http.MethodOptions, path, description...).
func OPTIONS(url string, description ...string) *HTTPTestCase {
	return DefaultContext().OPTIONS(url, description...)
}

// PATCH is a shortcut for NewTestCase(http.MethodPatch, path, description...).
func PATCH(url string, description ...string) *HTTPTestCase {
	return DefaultContext().PATCH(url, description...)
}

// POST is a shortcut for NewTestCase(http.MethodPost, path, description...).
func POST(url string, description ...string) *HTTPTestCase {
	return DefaultContext().POST(url, description...)
}

// PUT is a shortcut for NewTestCase(http.MethodPut, path, description...).
func PUT(url string, description ...string) *HTTPTestCase {
	return DefaultContext().PUT(url, description...)
}

// DO creates a test case from a custom HTTP request.
func DO(request *http.Request, description ...string) *HTTPTestCase {
	tc := DefaultContext().newHTTPTestCase(request.Method, request.URL.Path, description...)
	tc.request = request
	return tc
}

func createRequest(method, path string) (*http.Request, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	req, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return nil, cancel, err
	}

	return req, cancel, nil
}

func doRequest(c *http.Client, req *http.Request) (int, http.Header, []byte, error) {
	debug("%s %s", req.Method, req.URL.String())
	resp, err := c.Do(req)
	if err != nil {
		return -1, nil, nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, nil, err
	}

	return resp.StatusCode, resp.Header, body, nil
}

func handleRequest(h http.Handler, req *http.Request) (int, http.Header, []byte, error) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	resp := w.Result()
	b, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, b, err
}

func toBytes(body interface{}) ([]byte, error) {
	var b []byte
	if body != nil {
		var err error
		switch v := body.(type) {
		case []byte:
			b = v
		case string:
			b = []byte(v)
		case func() []byte:
			b = v()
		case func() ([]byte, error):
			b, err = v()
		default:
			b, err = json.Marshal(body)
		}

		if err != nil {
			return nil, fmt.Errorf("request body: %w", err)
		}
	}

	return b, nil
}

func toInterface(body []byte) interface{} {
	if len(body) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			return bodyMap
		}

		var bodyArray []interface{}
		if err := json.Unmarshal(body, &bodyArray); err == nil {
			return bodyArray
		}

		return string(body)
	}

	return nil
}
