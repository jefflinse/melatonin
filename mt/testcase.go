package mt

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jefflinse/melatonin/golden"
)

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

	// Description is a description of the test case.
	Description string

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

	// Underlying HTTP request for the test case.
	request *http.Request
}

// NewTestCase creates a new TestCase with the given method and path.
func NewTestCase(method, path string, description ...string) *HTTPTestCase {
	return &HTTPTestCase{
		Description: strings.Join(description, " "),
		Method:      method,
		Path:        path,
	}
}

// DisplayName returns the name of the test case.
func (tc *HTTPTestCase) DisplayName() string {
	return fmt.Sprintf("%s %s", tc.Method, tc.Path)
}

// DELETE is a shortcut for NewTestCase(http.MethodDelete, path).
func DELETE(path string, description ...string) *HTTPTestCase {
	return NewTestCase(http.MethodDelete, path, description...)
}

// HEAD is a shortcut for NewTestCase(http.MethodHead, path, description...).
func HEAD(path string, description ...string) *HTTPTestCase {
	return NewTestCase(http.MethodHead, path, description...)
}

// GET is a shortcut for NewTestCase(http.MethodGet, path, description...).
func GET(path string, description ...string) *HTTPTestCase {
	return NewTestCase(http.MethodGet, path, description...)
}

// OPTIONS is a shortcut for NewTestCase(http.MethodOptions, path, description...).
func OPTIONS(path string, description ...string) *HTTPTestCase {
	return NewTestCase(http.MethodOptions, path, description...)
}

// PATCH is a shortcut for NewTestCase(http.MethodPatch, path, description...).
func PATCH(path string, description ...string) *HTTPTestCase {
	return NewTestCase(http.MethodPatch, path, description...)
}

// POST is a shortcut for NewTestCase(http.MethodPost, path, description...).
func POST(path string, description ...string) *HTTPTestCase {
	return NewTestCase(http.MethodPost, path, description...)
}

// PUT is a shortcut for NewTestCase(http.MethodPut, path, description...).
func PUT(path string, description ...string) *HTTPTestCase {
	return NewTestCase(http.MethodPut, path, description...)
}

// DO creates a test case from a custom HTTP request.
func DO(request *http.Request, description ...string) *HTTPTestCase {
	return &HTTPTestCase{
		Method:  request.Method,
		Path:    request.URL.Path,
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
	if tc.Description != "" {
		color.HiYellow("overriding previous description")
	}

	tc.Description = description
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

// TestCaseResult represents the result of running a single test case.
type TestCaseResult struct {
	TestCase *HTTPTestCase
	Status   int
	Headers  http.Header
	Body     []byte
	Errors   []error

	executionTime time.Duration
}

// Failed indicates that the test case failed.
func (r *TestCaseResult) Failed() bool {
	return len(r.Errors) > 0
}

func (r *TestCaseResult) addErrors(errs ...error) {
	if len(errs) == 0 {
		return
	}

	r.Errors = append(r.Errors, errs...)
}

func (r *TestCaseResult) validateExpectations() {
	if r.TestCase.WantStatus != 0 {
		if err := expectStatus(r.TestCase.WantStatus, r.Status); err != nil {
			r.addErrors(err)
		}
	}

	if r.TestCase.WantHeaders != nil {
		if errs := expectHeaders(r.TestCase.WantHeaders, r.Headers); len(errs) > 0 {
			r.addErrors(errs...)
		}
	}

	if r.TestCase.WantBody != nil {
		body := parseResponseBody(r.Body)
		if errs := expect("body", r.TestCase.WantBody, body, r.TestCase.WantExactJSONBody); len(errs) > 0 {
			r.addErrors(errs...)
		}
	}
}
