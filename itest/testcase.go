package itest

import (
	"errors"
	"fmt"
	"net/http"
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
	After func() error

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

	// Before is an optional function that is run before the test is run.
	// It can be used to perform any prerequisites actions for the test,
	// such as adding or removing objects in a database. Any error
	// returned by Before is treated as a test failure.
	Before func() error

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

func (tc *TestCase) DisplayName() string {
	name := fmt.Sprintf("%s %s", tc.Method, tc.Path)
	if tc.RequestBody != nil {
		name += fmt.Sprintf(" (%d)", len(tc.RequestBody.String()))
	}

	return name
}

func GET(path string) *TestCase {
	return &TestCase{Method: "GET", Path: path}
}

func POST(path string) *TestCase {
	return &TestCase{Method: "POST", Path: path}
}

func PUT(path string) *TestCase {
	return &TestCase{Method: "PUT", Path: path}
}

func PATCH(path string) *TestCase {
	return &TestCase{Method: "PATCH", Path: path}
}

func DELETE(path string) *TestCase {
	return &TestCase{Method: "DELETE", Path: path}
}

func DO(request *http.Request, err error) *TestCase {
	if err != nil {
		fatal("invalid custom request: ", err)
	}

	return &TestCase{
		Method:  request.Method,
		Path:    request.URL.Path,
		request: request,
	}
}

func (tc *TestCase) DoAfter(after func() error) *TestCase {
	tc.requireMethodAndPath("WithAfter")
	if tc.After != nil {
		fatal("test case %q specifies more than one after-function", tc.DisplayName())
	}

	tc.After = after
	return tc
}

func (tc *TestCase) DoBefore(before func() error) *TestCase {
	tc.requireMethodAndPath("WithBefore")
	if tc.Before != nil {
		fatal("test case %q specifies more than one before function", tc.DisplayName())
	}

	tc.Before = before
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

func (tc *TestCase) WithBody(body Stringable) *TestCase {
	tc.requireMethodAndPath("WithBody")
	if tc.RequestBody != nil {
		fatal("test case %q specifies more than one request body", tc.DisplayName())
	}

	tc.RequestBody = body
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
