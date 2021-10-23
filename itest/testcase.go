package itest

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// A TestCase tests a single call to an HTTP endpoint.
//
// An optional setup function can be provided to perform any necessary
// setup before the test is run, such as adding or removing objects in
// a database.
//
// All fields in the WantBody map are expected to be present in the
// response body.
type TestCase struct {
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

	// ContinueOnFailure indicates whether the test should continue to the next
	// test case if the current test fails. Default is false.
	ContinueOnFailure bool

	// Method is the HTTP method to use for the request. Default is "GET".
	Method string

	// Path is the relative Path to use for the request.
	Path string

	// RequestHeaders is a map of HTTP headers to use for the request.
	RequestHeaders http.Header

	// RequestBody is the content to send in the body of the HTTP request.
	RequestBody Stringable

	// Timeout is the maximum amount of time to wait for the request to complete.
	//
	// Default is 5 seconds.
	Timeout time.Duration

	// WantStatus is the expected HTTP status code of the response. Default is 200.
	WantStatus int

	// WantHeaders is a map of HTTP headers that are expected to be present in
	// the HTTP response.
	WantHeaders http.Header

	// WantBody is the expected HTTP response body content.
	WantBody Stringable

	// Underlying HTTP request for the test case.
	request *http.Request
}

// NewTestCase creates a new TestCase with the given method and path.
func NewTestCase(method, path string) *TestCase {
	return &TestCase{
		Method: method,
		Path:   path,
	}
}

// DisplayName returns the name of the test case.
func (tc *TestCase) DisplayName() string {
	name := fmt.Sprintf("%s %s", tc.Method, tc.Path)
	if tc.RequestBody != nil {
		name += fmt.Sprintf(" (%d)", len(tc.RequestBody.String()))
	}

	return name
}

// DELETE is a shortcut for NewTestCase("DELETE", path).
func DELETE(path string) *TestCase {
	return NewTestCase("DELETE", path)
}

// HEAD is a shortcut for NewTestCase("HEAD", path).
func HEAD(path string) *TestCase {
	return NewTestCase("HEAD", path)
}

// GET is a shortcut for NewTestCase("GET", path).
func GET(path string) *TestCase {
	return NewTestCase("GET", path)
}

// OPTIONS is a shortcut for NewTestCase("OPTIONS", path).
func OPTIONS(path string) *TestCase {
	return NewTestCase("OPTIONS", path)
}

// PATCH is a shortcut for NewTestCase("PATCH", path).
func PATCH(path string) *TestCase {
	return NewTestCase("PATCH", path)
}

// POST is a shortcut for NewTestCase("POST", path).
func POST(path string) *TestCase {
	return NewTestCase("POST", path)
}

// PUT is a shortcut for NewTestCase("PUT", path).
func PUT(path string) *TestCase {
	return NewTestCase("PUT", path)
}

// DO creates a test case from a custom HTTP request.
func DO(request *http.Request) *TestCase {
	return &TestCase{
		Method:  request.Method,
		Path:    request.URL.Path,
		request: request,
	}
}

// After registers a function to be run after the test case.
func (tc *TestCase) After(after func() error) *TestCase {
	if tc.AfterFunc != nil {
		warn("overriding previously defined AfterFunc")
	}

	tc.AfterFunc = after
	return tc
}

// Before registers a function to be run before the test case.
func (tc *TestCase) Before(before func() error) *TestCase {
	if tc.BeforeFunc != nil {
		warn("overriding previously defined BeforeFunc")
	}

	tc.BeforeFunc = before
	return tc
}

// WithBody sets the request body for the test case.
func (tc *TestCase) WithBody(body Stringable) *TestCase {
	if tc.RequestBody != nil {
		warn("overriding previously defined request body")
	}

	tc.RequestBody = body
	return tc
}

// WithHeaders sets the request headers for the test case.
func (tc *TestCase) WithHeaders(headers http.Header) *TestCase {
	if tc.RequestHeaders != nil {
		warn("overriding previously defined request headers")
	}

	tc.RequestHeaders = headers
	return tc
}

// WithHeader adds a request header to the test case.
func (tc *TestCase) WithHeader(key, value string) *TestCase {
	if tc.RequestHeaders == nil {
		tc.RequestHeaders = http.Header{}
	}

	tc.RequestHeaders.Set(key, value)
	return tc
}

// WithTimeout sets a timeout for the test case.
func (tc *TestCase) WithTimeout(timeout time.Duration) *TestCase {
	if tc.Timeout != 0 {
		warn("overriding previously defined timeout")
	}

	tc.Timeout = timeout
	return tc
}

// ExpectStatus sets the expected HTTP status code for the test case.
func (tc *TestCase) ExpectStatus(status int) *TestCase {
	if tc.WantStatus > 0 {
		warn("overriding previously expected status")
	}

	tc.WantStatus = status
	return tc
}

// ExpectHeaders sets the expected HTTP response headers for the test case.
func (tc *TestCase) ExpectHeaders(headers http.Header) *TestCase {
	if tc.WantHeaders != nil && len(tc.WantHeaders) > 0 {
		warn("overriding previously expected headers")
	}

	tc.WantHeaders = headers
	return tc
}

// ExpectHeader adds an expected HTTP response header for the test case.
func (tc *TestCase) ExpectHeader(key, value string) *TestCase {
	if tc.WantHeaders == nil {
		tc.WantHeaders = http.Header{}
	}

	tc.WantHeaders.Set(key, value)
	return tc
}

// ExpectBody sets the expected HTTP response body for the test case.
func (tc *TestCase) ExpectBody(body Stringable) *TestCase {
	if tc.WantBody != nil {
		warn("overriding previously expected body")
	}

	tc.WantBody = body
	return tc
}

// Validate ensures that the test case is valid can can be run.
func (tc *TestCase) Validate() error {
	if tc.Method == "" {
		return errors.New("missing Method")
	} else if tc.Path == "" {
		return errors.New("missing Path")
	} else if tc.Path[0] != '/' {
		return errors.New("path must begin with '/'")
	}

	return nil
}

// TestCaseResult represents the result of running a single test case.
type TestCaseResult struct {
	TestCase *TestCase
	Errors   []error
}

// AddError adds an error to the test result.
func (r *TestCaseResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
}
