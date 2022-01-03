package mt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/jefflinse/melatonin/golden"
	mtjson "github.com/jefflinse/melatonin/json"
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

	// Desc is a description of the test case.
	Desc string

	// Expectations is a set of values to compare the response against.
	Expectations expectatons `json:"expectations"`

	// GoldenFilePath is a path to a golden file defining expectations for the test case.
	//
	// If set, any WantStatus, WantHeaders, or WantBody values are overridden with
	// values from the golden file.
	GoldenFilePath string

	// Path parameters to be mapped into the request path.
	pathParams valueMap

	// Body for the HTTP request. May contain deferred values.
	requestBody interface{}

	// Configuration for the test
	tctx *HTTPTestContext

	// Underlying HTTP request for the test case.
	request *http.Request

	// Cancel function for the underlying HTTP request.
	cancel context.CancelFunc
}

// expectatons represents the expected values for single HTTP response.
type expectatons struct {
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

var _ TestCase = &HTTPTestCase{}

// Action returns a short, uppercase verb describing the action performed by the
// test case.
func (tc *HTTPTestCase) Action() string {
	return strings.ToUpper(tc.request.Method)
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

// Description returns a string describing the test case.
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

// Execute runs the test case.
func (tc *HTTPTestCase) Execute() TestResult {
	if tc.cancel != nil {
		defer tc.cancel()
	}

	result := &HTTPTestCaseResult{
		testCase: tc,
	}

	if tc.BeforeFunc != nil {
		if err := tc.BeforeFunc(); err != nil {
			return result.addFailures(err)
		}
	}

	// apply path parameters
	expandedPath, err := tc.pathParams.apply(tc.request.URL.Path)
	if err != nil {
		return result.addFailures(err)
	}

	tc.request.URL.Path = expandedPath

	// resolve deferred values
	resolvedBody, err := mtjson.ResolveDeferred(tc.requestBody)
	if err != nil {
		return result.addFailures(err)
	}

	b, err := toBytes(resolvedBody)
	if err != nil {
		return result.addFailures(err)
	}

	tc.request.Body = io.NopCloser(bytes.NewReader(b))

	if tc.tctx.Handler != nil {
		result.Status, result.Headers, result.Body, err = handleRequest(tc.tctx.Handler, tc.request)
		if err != nil {
			return result.addFailures(fmt.Errorf("failed to handle HTTP request: %w", err))
		}
	} else {
		if tc.tctx.Client == nil {
			tc.tctx.Client = http.DefaultClient
		}

		result.Status, result.Headers, result.Body, err = doRequest(tc.tctx.Client, tc.request)
		if err != nil {
			return result.addFailures(fmt.Errorf("failed to execute HTTP request: %w", err))
		}
	}

	result.validateExpectations()

	if tc.AfterFunc != nil {
		if err := tc.AfterFunc(); err != nil {
			result.addFailures(err)
		}
	}

	return result
}

// Target returns a string representing the target of the action performed by the
// test case.
func (tc *HTTPTestCase) Target() string {
	return tc.request.URL.Path
}

//
// Chainable qualifier methods that can be used to configure the test case.
//

// WithBody sets the request body for the test case.
func (tc *HTTPTestCase) WithBody(body interface{}) *HTTPTestCase {
	tc.requestBody = body
	return tc
}

// WithHeader adds a request header to the test case.
func (tc *HTTPTestCase) WithHeader(key, value string) *HTTPTestCase {
	tc.request.Header.Set(key, value)
	return tc
}

// WithHeaders sets the request headers for the test case.
func (tc *HTTPTestCase) WithHeaders(headers http.Header) *HTTPTestCase {
	tc.request.Header = headers
	return tc
}

// WithPathParam adds a request path parameter to the test case.
func (tc *HTTPTestCase) WithPathParam(key string, value interface{}) *HTTPTestCase {
	tc.pathParams[key] = value
	return tc
}

// WithPathParams sets the request path parameters for the test case.
func (tc *HTTPTestCase) WithPathParams(params map[string]interface{}) *HTTPTestCase {
	tc.pathParams = params
	return tc
}

// WithQueryParam adds a request query parameter to the test case.
func (tc *HTTPTestCase) WithQueryParam(key, value string) *HTTPTestCase {
	q := tc.request.URL.Query()
	q.Add(key, value)
	tc.request.URL.RawQuery = q.Encode()
	return tc
}

// WithQueryParams sets the request query parameters for the test case.
func (tc *HTTPTestCase) WithQueryParams(params url.Values) *HTTPTestCase {
	tc.request.URL.RawQuery = params.Encode()
	return tc
}

// WithTimeout sets a timeout for the test case.
func (tc *HTTPTestCase) WithTimeout(timeout time.Duration) *HTTPTestCase {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	tc.request = tc.request.WithContext(ctx)
	tc.cancel = cancel
	return tc
}

//
// Chainable expectation methods that can be used to configure the test case.
//

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

// ExpectExactHeaders sets the expected HTTP response headers for the test case.
//
// Unlike ExpectHeaders, ExpectExactHeaders willl cause the test case to fail
// if any unexpected headers are present in the response.
func (tc *HTTPTestCase) ExpectExactHeaders(headers http.Header) *HTTPTestCase {
	tc.Expectations.WantExactHeaders = true
	return tc.ExpectHeaders(headers)
}

// ExpectHeader adds an expected HTTP response header for the test case.
func (tc *HTTPTestCase) ExpectHeader(key, value string) *HTTPTestCase {
	if tc.Expectations.Headers == nil {
		tc.Expectations.Headers = http.Header{}
	}

	tc.Expectations.Headers.Set(key, value)
	return tc
}

// ExpectHeaders sets the expected HTTP response headers for the test case.
//
// Unlike ExpectExactHeaders, ExpectHeaders only verifies that the expected
// headers are present in the response, and ignores any additional headers.
func (tc *HTTPTestCase) ExpectHeaders(headers http.Header) *HTTPTestCase {
	tc.Expectations.Headers = headers
	return tc
}

// ExpectGolden causes the test case to load its HTTP response expectations
// from a golden file.
func (tc *HTTPTestCase) ExpectGolden(path string) *HTTPTestCase {
	tc.GoldenFilePath = path
	return tc
}

// ExpectStatus sets the expected HTTP status code for the test case.
func (tc *HTTPTestCase) ExpectStatus(status int) *HTTPTestCase {
	tc.Expectations.Status = status
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

type jsonTestCase struct {
	Headers      http.Header              `json:"headers,omitempty"`
	Body         interface{}              `json:"body,omitempty"`
	Expectations jsonTestCaseExpectations `json:"expectations,omitempty"`
}

type jsonTestCaseExpectations struct {
	Status            int         `json:"status,omitempty"`
	Headers           http.Header `json:"headers,omitempty"`
	Body              interface{} `json:"body,omitempty"`
	WantExactHeaders  bool        `json:"want_exact_headers"`
	WantExactJSONBody bool        `json:"want_exact_json_body"`
}

// MarshalJSON customizes the JSON representaton of the test case.
func (tc HTTPTestCase) MarshalJSON() ([]byte, error) {
	o := jsonTestCase{
		Headers: tc.request.Header,
		Body:    tc.request.Body,
		Expectations: jsonTestCaseExpectations{
			Status:            tc.Expectations.Status,
			Headers:           tc.Expectations.Headers,
			Body:              tc.Expectations.Body,
			WantExactHeaders:  tc.Expectations.WantExactHeaders,
			WantExactJSONBody: tc.Expectations.WantExactJSONBody,
		},
	}

	return json.Marshal(o)
}
