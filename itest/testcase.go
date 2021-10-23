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

func NewTestCase(method, path string) *TestCase {
	return &TestCase{
		Method: method,
		Path:   path,
	}
}

func (tc *TestCase) DisplayName() string {
	name := fmt.Sprintf("%s %s", tc.Method, tc.Path)
	if tc.RequestBody != nil {
		name += fmt.Sprintf(" (%d)", len(tc.RequestBody.String()))
	}

	return name
}

func DELETE(path string) *TestCase {
	return NewTestCase("DELETE", path)
}

func HEAD(path string) *TestCase {
	return NewTestCase("HEAD", path)
}

func GET(path string) *TestCase {
	return NewTestCase("GET", path)
}

func OPTIONS(path string) *TestCase {
	return NewTestCase("OPTIONS", path)
}

func PATCH(path string) *TestCase {
	return NewTestCase("PATCH", path)
}

func POST(path string) *TestCase {
	return NewTestCase("POST", path)
}

func PUT(path string) *TestCase {
	return NewTestCase("PUT", path)
}

func DO(request *http.Request) *TestCase {
	return &TestCase{
		Method:  request.Method,
		Path:    request.URL.Path,
		request: request,
	}
}

func (tc *TestCase) After(after func() error) *TestCase {
	tc.requireMethodAndPath("WithAfter")
	if tc.AfterFunc != nil {
		fatal("test case %q specifies more than one after-function", tc.DisplayName())
	}

	tc.AfterFunc = after
	return tc
}

func (tc *TestCase) Before(before func() error) *TestCase {
	tc.requireMethodAndPath("WithBefore")
	if tc.BeforeFunc != nil {
		fatal("test case %q specifies more than one before function", tc.DisplayName())
	}

	tc.BeforeFunc = before
	return tc
}

func (tc *TestCase) WithBody(body Stringable) *TestCase {
	tc.requireMethodAndPath("WithBody")
	if tc.RequestBody != nil {
		fatal("test case %q specifies more than one request body", tc.DisplayName())
	}

	tc.RequestBody = body
	return tc
}

func (tc *TestCase) WithHeaders(headers http.Header) *TestCase {
	tc.requireMethodAndPath("WithHeaders")
	if tc.RequestHeaders != nil {
		fatal("test case %q specifies request headers more than once", tc.DisplayName())
	}

	tc.RequestHeaders = headers
	return tc
}

func (tc *TestCase) WithHeader(key, value string) *TestCase {
	tc.requireMethodAndPath("WithHeader")
	if tc.RequestHeaders == nil {
		tc.RequestHeaders = http.Header{}
	}

	tc.RequestHeaders.Set(key, value)
	return tc
}

func (tc *TestCase) WithTimeout(timeout time.Duration) *TestCase {
	tc.requireMethodAndPath("WithTimeout")
	if tc.Timeout != 0 {
		fatal("test case %q specifies more than one timeout", tc.DisplayName())
	}

	tc.Timeout = timeout
	return tc
}

func (tc *TestCase) ExpectStatus(status int) *TestCase {
	tc.requireMethodAndPath("ExpectStatus")
	if tc.WantStatus > 0 {
		fatal("test case %q specifies more than one expected status", tc.DisplayName())
	}

	tc.WantStatus = status
	return tc
}

func (tc *TestCase) ExpectHeaders(headers http.Header) *TestCase {
	tc.requireMethodAndPath("ExpectHeaders")
	if tc.WantHeaders != nil && len(tc.WantHeaders) > 0 {
		fatal("test case %q overrides previously defined expected headers", tc.DisplayName())
	}

	tc.WantHeaders = headers
	return tc
}

func (tc *TestCase) ExpectHeader(key, value string) *TestCase {
	tc.requireMethodAndPath("ExpectHeader")
	if tc.WantHeaders == nil {
		tc.WantHeaders = http.Header{}
	}

	tc.WantHeaders.Set(key, value)
	return tc
}

func (tc *TestCase) ExpectBody(body Stringable) *TestCase {
	tc.requireMethodAndPath("ExpectBody")
	if tc.WantBody != nil {
		fatal("test case %q specifies more than one expected body", tc.DisplayName())
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

func (tc *TestCase) requireMethodAndPath(caller string) {
	if tc.Method == "" || tc.Path == "" {
		fatal("test case must define method and path before specifying %s()", caller)
	}
}
