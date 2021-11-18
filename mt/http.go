package mt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/jefflinse/melatonin/golden"
)

const (
	modeNone = iota
	modeBaseURL
	modeHandler
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
			color.HiYellow("invalid MELATONIN_DEFAULT_TEST_TIMEOUT value %q in environment, using default of %s",
				envTimeoutStr, defaultRequestTimeoutStr)
		}
	}
}

// Object is a type alias for map[string]interface{}.
type Object map[string]interface{}

// Array is a type alias for []interface{}.
type Array []interface{}

type HTTPTestContext struct {
	BaseURL string
	Client  *http.Client
	Handler http.Handler
	mode    int
}

// NewURLContext creates a new HTTPTestContext for creating tests that target
// the specified base URL.
func NewURLContext(baseURL string) *HTTPTestContext {
	return &HTTPTestContext{
		BaseURL: baseURL,
		mode:    modeBaseURL,
	}
}

// NewHandlerContext creates a new HTTPTestContext for creating tests that target
// the specified HTTP handler.
func NewHandlerContext(handler http.Handler) *HTTPTestContext {
	return &HTTPTestContext{
		Handler: handler,
		mode:    modeHandler,
	}
}

// WithHTTPClient sets the HTTP client used for HTTP requests and returns the
// context.
func (c *HTTPTestContext) WithHTTPClient(client *http.Client) *HTTPTestContext {
	c.Client = client
	return c
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

	// GoldenFilePath is a path to a golden file defining expectations for the test case.
	//
	// If set, any WantStatus, WantHeaders, or WantBody values are overriden with
	// values from the golden file.
	GoldenFilePath string

	// Method is the HTTP method to use for the request. Default is "GET".
	Method string

	// Path is the relative Path to use for the request. Must begin with "/".
	Path string

	// QueryParams is a map of query string parameters.
	QueryParams url.Values

	// RequestHeaders is a map of HTTP headers to use for the request.
	RequestHeaders http.Header

	// RequestBody is the content to send in the body of the HTTP request.
	RequestBody interface{}

	// Timeout is the maximum amount of time to wait for the request to complete.
	//
	// Default is 5 seconds.
	Timeout time.Duration

	// WantBody is the expected HTTP response body content.
	WantBody interface{}

	// WantExactHeaders indicates whether or not any unexpected response headers
	// should be treated as a test failure.
	WantExactHeaders bool

	// WantExactJSONBody indicates whether or not the expected JSON should be matched
	// exactly (true) or treated as a subset of the response JSON (false).
	WantExactJSONBody bool

	// WantHeaders is a map of HTTP headers that are expected to be present in
	// the HTTP response.
	WantHeaders http.Header

	// WantStatus is the expected HTTP status code of the response. Default is 200.
	WantStatus int

	// Configuration for the test
	context *HTTPTestContext

	// Underlying HTTP request for the test case.
	request *http.Request
}

var _ TestCase = &HTTPTestCase{}

func newHTTPTestCase(context *HTTPTestContext, method, path string, description ...string) *HTTPTestCase {
	return &HTTPTestCase{
		Desc:    strings.Join(description, " "),
		Method:  method,
		Path:    path,
		context: context,
	}
}

func (tc *HTTPTestCase) Action() string {
	return tc.Method
}

func (tc *HTTPTestCase) Target() string {
	return tc.Path
}

func (tc *HTTPTestCase) Description() string {
	return tc.Desc
}

func (tc *HTTPTestCase) Execute(t *testing.T) (TestResult, error) {
	if err := tc.Validate(); err != nil {
		return nil, err
	}

	result := &HTTPTestCaseResult{
		testCase: tc,
	}

	if tc.BeforeFunc != nil {
		debug("%s: running before()", tc.DisplayName())
		if err := tc.BeforeFunc(); err != nil {
			result.addErrors(fmt.Errorf("before(): %w", err))
			return result, nil
		}
	}

	timeout := defaultRequestTimeout
	if tc.Timeout > 0 {
		timeout = tc.Timeout
	}

	var body []byte
	var err error
	if tc.RequestBody != nil {
		switch v := tc.RequestBody.(type) {
		case []byte:
			body = v
		case string:
			body = []byte(v)
		case func() []byte:
			body = v()
		case func() ([]byte, error):
			body, err = v()
		default:
			body, err = json.Marshal(tc.RequestBody)
		}

		if err != nil {
			result.addErrors(fmt.Errorf("request body: %w", err))
			return result, nil
		}
	}

	if tc.request == nil {
		req, cancel, err := createRequest(
			tc.Method,
			tc.context.BaseURL+tc.Path,
			tc.QueryParams,
			tc.RequestHeaders,
			body,
			timeout)
		defer cancel()
		if err != nil {
			result.addErrors(fmt.Errorf("failed to create HTTP request: %w", err))
			return result, err
		}

		tc.request = req
	}

	if tc.context.mode == modeBaseURL {
		result.Status, result.Headers, result.Body, err = doRequest(tc.context.Client, tc.request)
		if err != nil {
			debug("%s: failed to execute HTTP request: %s", tc.DisplayName(), err)
			result.addErrors(fmt.Errorf("failed to execute HTTP request: %w", err))
			return nil, err
		}
	} else if tc.context.mode == modeHandler {
		result.Status, result.Headers, result.Body, err = handleRequest(tc.context.Handler, tc.request)
		if err != nil {
			debug("%s: failed to handle HTTP request: %s", tc.DisplayName(), err)
			result.addErrors(fmt.Errorf("failed to handle HTTP request: %w", err))
			return nil, err
		}
	}

	result.validateExpectations()

	if tc.AfterFunc != nil {
		debug("%s: running after()", tc.DisplayName())
		if err := tc.AfterFunc(); err != nil {
			result.addErrors(fmt.Errorf("after(): %w", err))
		}
	}

	return result, err
}

// DisplayName returns the name of the test case.
func (tc *HTTPTestCase) DisplayName() string {
	return fmt.Sprintf("%s %s", tc.Method, tc.Path)
}

// DELETE is a shortcut for NewTestCase(http.MethodDelete, path).
func (c *HTTPTestContext) DELETE(path string, description ...string) *HTTPTestCase {
	return newHTTPTestCase(c, http.MethodDelete, path, description...)
}

// HEAD is a shortcut for NewTestCase(http.MethodHead, path, description...).
func (c *HTTPTestContext) HEAD(path string, description ...string) *HTTPTestCase {
	return newHTTPTestCase(c, http.MethodHead, path, description...)
}

// GET is a shortcut for NewTestCase(http.MethodGet, path, description...).
func (c *HTTPTestContext) GET(path string, description ...string) *HTTPTestCase {
	return newHTTPTestCase(c, http.MethodGet, path, description...)
}

// OPTIONS is a shortcut for NewTestCase(http.MethodOptions, path, description...).
func (c *HTTPTestContext) OPTIONS(path string, description ...string) *HTTPTestCase {
	return newHTTPTestCase(c, http.MethodOptions, path, description...)
}

// PATCH is a shortcut for NewTestCase(http.MethodPatch, path, description...).
func (c *HTTPTestContext) PATCH(path string, description ...string) *HTTPTestCase {
	return newHTTPTestCase(c, http.MethodPatch, path, description...)
}

// POST is a shortcut for NewTestCase(http.MethodPost, path, description...).
func (c *HTTPTestContext) POST(path string, description ...string) *HTTPTestCase {
	return newHTTPTestCase(c, http.MethodPost, path, description...)
}

// PUT is a shortcut for NewTestCase(http.MethodPut, path, description...).
func (c *HTTPTestContext) PUT(path string, description ...string) *HTTPTestCase {
	return newHTTPTestCase(c, http.MethodPut, path, description...)
}

// DO creates a test case from a custom HTTP request.
func (c *HTTPTestContext) DO(request *http.Request, description ...string) *HTTPTestCase {
	return &HTTPTestCase{
		Method:  request.Method,
		Path:    request.URL.Path,
		context: c,
		request: request,
	}
}

// After registers a function to be run after the test case.
func (tc *HTTPTestCase) After(after func() error) *HTTPTestCase {
	if tc.AfterFunc != nil {
		color.HiYellow("overriding previously defined AfterFunc")
	}

	tc.AfterFunc = after
	return tc
}

// Before registers a function to be run before the test case.
func (tc *HTTPTestCase) Before(before func() error) *HTTPTestCase {
	if tc.BeforeFunc != nil {
		color.HiYellow("overriding previously defined BeforeFunc")
	}

	tc.BeforeFunc = before
	return tc
}

// Describe sets a description for the test case.
func (tc *HTTPTestCase) Describe(description string) *HTTPTestCase {
	if tc.Desc != "" {
		color.HiYellow("overriding previous description")
	}

	tc.Desc = description
	return tc
}

// WithBody sets the request body for the test case.
func (tc *HTTPTestCase) WithBody(body interface{}) *HTTPTestCase {
	if tc.RequestBody != nil {
		color.HiYellow("overriding previously defined request body")
	}

	tc.RequestBody = body
	return tc
}

// WithHeaders sets the request headers for the test case.
func (tc *HTTPTestCase) WithHeaders(headers http.Header) *HTTPTestCase {
	if tc.RequestHeaders != nil {
		color.HiYellow("overriding previously defined request headers")
	}

	tc.RequestHeaders = headers
	return tc
}

// WithHeader adds a request header to the test case.
func (tc *HTTPTestCase) WithHeader(key, value string) *HTTPTestCase {
	if tc.RequestHeaders == nil {
		tc.RequestHeaders = http.Header{}
	}

	tc.RequestHeaders.Set(key, value)
	return tc
}

func (tc *HTTPTestCase) WithQueryParams(params url.Values) *HTTPTestCase {
	if tc.QueryParams != nil {
		color.HiYellow("overriding previously defined query params")
	}

	tc.QueryParams = params
	return tc
}

func (tc *HTTPTestCase) WithQueryParam(key, value string) *HTTPTestCase {
	if tc.QueryParams == nil {
		tc.QueryParams = url.Values{}
	}

	tc.QueryParams.Add(key, value)
	return tc
}

// WithTimeout sets a timeout for the test case.
func (tc *HTTPTestCase) WithTimeout(timeout time.Duration) *HTTPTestCase {
	if tc.Timeout != 0 {
		color.HiYellow("overriding previously defined timeout")
	}

	tc.Timeout = timeout
	return tc
}

// ExpectStatus sets the expected HTTP status code for the test case.
func (tc *HTTPTestCase) ExpectStatus(status int) *HTTPTestCase {
	if tc.WantStatus > 0 {
		color.HiYellow("overriding previously expected status")
	}

	tc.WantStatus = status
	return tc
}

// ExpectExactHeaders sets the expected HTTP response headers for the test case.
//
// Unlike ExpectHeaders, ExpectExactHeaders willl cause the test case to fail
// if any unexpected headers are present in the response.
func (tc *HTTPTestCase) ExpectExactHeaders(headers http.Header) *HTTPTestCase {
	tc.WantExactHeaders = true
	return tc.ExpectHeaders(headers)
}

// ExpectHeaders sets the expected HTTP response headers for the test case.
//
// Unlike ExpectExactHeaders, ExpectHeaders only verifies that the expected
// headers are present in the response, and ignores any additional headers.
func (tc *HTTPTestCase) ExpectHeaders(headers http.Header) *HTTPTestCase {
	if tc.WantHeaders != nil && len(tc.WantHeaders) > 0 {
		color.HiYellow("overriding previously expected headers")
	}

	tc.WantHeaders = headers
	return tc
}

// ExpectHeader adds an expected HTTP response header for the test case.
func (tc *HTTPTestCase) ExpectHeader(key, value string) *HTTPTestCase {
	if tc.WantHeaders == nil {
		tc.WantHeaders = http.Header{}
	}

	tc.WantHeaders.Set(key, value)
	return tc
}

// ExpectBody sets the expected HTTP response body for the test case.
func (tc *HTTPTestCase) ExpectBody(body interface{}) *HTTPTestCase {
	if tc.WantBody != nil {
		color.HiYellow("overriding previously expected body")
	}

	tc.WantBody = body
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
	tc.WantExactJSONBody = true
	return tc.ExpectBody(body)
}

func (tc *HTTPTestCase) ExpectGolden(path string) *HTTPTestCase {
	if tc.GoldenFilePath != "" {
		color.HiYellow("overriding previously expected golden file")
	}

	tc.GoldenFilePath = path
	return tc
}

// Validate ensures that the test case is valid can can be run.
func (tc *HTTPTestCase) Validate() error {
	if tc.Method == "" {
		return errors.New("missing Method")
	} else if tc.Path == "" {
		return errors.New("missing Path")
	} else if tc.Path[0] != '/' {
		return errors.New("path must begin with '/'")
	}

	if tc.GoldenFilePath != "" {
		golden, err := golden.LoadFile(tc.GoldenFilePath)
		if err != nil {
			return err
		}

		if tc.WantStatus != 0 {
			color.HiYellow("overriding previously expected status with golden file value")
		}
		tc.WantStatus = golden.WantStatus

		if tc.WantHeaders != nil {
			color.HiYellow("overriding previously expected headers with golden file content")
		}
		tc.WantHeaders = golden.WantHeaders

		if tc.WantBody != nil {
			color.HiYellow("overriding previously expected body with golden file content")
		}
		tc.WantBody = golden.WantBody
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

func (r *HTTPTestCaseResult) addErrors(errs ...error) {
	if len(errs) == 0 {
		return
	}

	r.errors = append(r.errors, errs...)
}

func parseResponseBody(body []byte) interface{} {
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

func (r *HTTPTestCaseResult) validateExpectations() {
	tc := r.TestCase().(*HTTPTestCase)
	if tc.WantStatus != 0 {
		if err := expectStatus(tc.WantStatus, r.Status); err != nil {
			r.addErrors(err)
		}
	}

	if tc.WantHeaders != nil {
		if errs := expectHeaders(tc.WantHeaders, r.Headers); len(errs) > 0 {
			r.addErrors(errs...)
		}
	}

	if tc.WantBody != nil {
		body := parseResponseBody(r.Body)
		if errs := expect("body", tc.WantBody, body, tc.WantExactJSONBody); len(errs) > 0 {
			r.addErrors(errs...)
		}
	}
}

func createRequest(method, path string,
	query url.Values,
	headers http.Header,
	body []byte,
	timeout time.Duration) (*http.Request, context.CancelFunc, error) {

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	req, err := http.NewRequestWithContext(ctx, method, path, reader)
	if err != nil {
		return nil, cancel, err
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	if headers != nil {
		req.Header = headers
	} else {
		req.Header = http.Header{}
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
