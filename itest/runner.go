package itest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultRequestTimeoutStr = "5s"
)

var (
	defaultRequestTimeout time.Duration
)

func init() {
	defaultRequestTimeout, _ = time.ParseDuration(defaultRequestTimeoutStr)
	envTimeoutStr := os.Getenv("ITEST_DEFAULT_TEST_TIMEOUT")
	if envTimeoutStr != "" {
		if timeout, err := time.ParseDuration(envTimeoutStr); err == nil {
			defaultRequestTimeout = timeout
		} else {
			warn("invalid ITEST_DEFAULT_TEST_TIMEOUT value %q in environment, using default of %s",
				envTimeoutStr, defaultRequestTimeoutStr)
		}
	}
}

// GoTestContext is a minimal interface for testing.T.
type GoTestContext interface {
	Run(string, func(*testing.T)) bool
}

// TestRunner contains configuration for running tests.
type TestRunner struct {
	// BaseURL is the base URL for the API, including the port.
	//
	// Examples:
	//   http://localhost:8080
	//   https://api.example.com
	//
	// Required.
	BaseURL string

	// ContinueOnFailure indicates whether the test runner should continue
	// executing further tests after a failure.
	//
	// Defaults is false.
	ContinueOnFailure bool

	// HTTPClient is the HTTP client to use for requests.
	//
	// If left unset, http.DefaultClient will be used.
	HTTPClient *http.Client

	// RequestTimeout is the default timeout for HTTP requests.
	//
	// Default is 5 seconds.
	RequestTimeout time.Duration
}

// NewTestRunner creates a new TestRunner with the default settings.
func NewTestRunner(baseURL string) *TestRunner {
	return &TestRunner{
		BaseURL:           baseURL,
		ContinueOnFailure: false,
		HTTPClient:        http.DefaultClient,
	}
}

// WithContinueOnFailure sets the ContinueOnFailure field of the TestRunner and
// returns the TestRunner.
func (r *TestRunner) WithContinueOnFailure(continueOnFailure bool) *TestRunner {
	r.ContinueOnFailure = continueOnFailure
	return r
}

// WithHTTPClient sets the HTTPClient field of the TestRunner and returns the
// TestRunner.
func (r *TestRunner) WithHTTPClient(client *http.Client) *TestRunner {
	r.HTTPClient = client
	return r
}

// WithRequestTimeout sets the RequestTimeout field of the TestRunner and returns
// the TestRunner.
func (r *TestRunner) WithRequestTimeout(timeout time.Duration) *TestRunner {
	r.RequestTimeout = timeout
	return r
}

// RunTests runs a set of tests.
//
// To run tests within the context of a Go test, use RunTestsT().
func (r *TestRunner) RunTests(tests []*TestCase) ([]*TestCaseResult, error) {
	return r.RunTestsT(nil, tests)
}

// RunTests runs a set of tests within the context of a Go test.
//
// To run tests as a standalone binary without a testing context, use RunTests().
func (r *TestRunner) RunTestsT(t GoTestContext, tests []*TestCase) ([]*TestCaseResult, error) {
	results := []*TestCaseResult{}

	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("invalid test runner: %w", err)
	}

	if r.HTTPClient == nil {
		debug("using default HTTP client")
		r.HTTPClient = http.DefaultClient
	}

	if !validateTests(tests) {
		return nil, fmt.Errorf("one or more test cases failed validation")
	}

	info("running %d tests for %s", len(tests), r.BaseURL)
	var executed, passed, failed, skipped int
	for _, test := range tests {
		result, err := r.RunTestT(t, test)
		if err != nil {
			return results, err
		}

		results = append(results, result)
		executed++

		if result.Failed() {
			failed++

			if !r.ContinueOnFailure {
				warn("skipping remaininig tests")
				skipped = len(tests) - executed
				break
			}

		} else {
			passed++
			info("%s  %s %s", greenText("✔"), test.request.Method, test.request.URL.Path)
		}
	}

	info("\n%d passed, %d failed, %d skipped", passed, failed, skipped)
	return results, nil
}

// RunTest runs a single test.
func (r *TestRunner) RunTest(test *TestCase) (*TestCaseResult, error) {
	return r.RunTestT(nil, test)
}

// RunTest runs a single test within a Go testing context.
func (r *TestRunner) RunTestT(t GoTestContext, test *TestCase) (*TestCaseResult, error) {
	result := &TestCaseResult{
		TestCase: test,
		Errors:   []error{},
	}

	if test.BeforeFunc != nil {
		debug("%s: running before()", test.DisplayName())
		if err := test.BeforeFunc(); err != nil {
			result.addErrors(fmt.Errorf("before(): %w", err))
			return result, nil
		}
	}

	timeout := defaultRequestTimeout
	if test.Timeout > 0 {
		timeout = test.Timeout
	} else if r.RequestTimeout > 0 {
		timeout = r.RequestTimeout
	}

	var reqBody []byte
	if test.RequestBody != nil {
		reqBody = []byte(test.RequestBody.String())
	}

	if test.request == nil {
		req, cancel, err := r.createRequest(
			test.Method,
			r.BaseURL+test.Path,
			test.RequestHeaders,
			reqBody,
			timeout)
		defer cancel()
		if err != nil {
			result.addErrors(fmt.Errorf("failed to create HTTP request: %w", err))
			return result, err
		}

		test.request = req
	}

	var err error
	result.Status, result.Headers, result.Body, err = doRequest(r.HTTPClient, test.request)
	if err != nil {
		debug("%s: failed to execute HTTP request: %s", test.DisplayName(), err)
		result.addErrors(fmt.Errorf("failed to perform HTTP request: %w", err))
		return result, err
	}

	result.validateExpectations()

	if test.AfterFunc != nil {
		debug("%s: running after()", test.DisplayName())
		if err := test.AfterFunc(); err != nil {
			result.addErrors(fmt.Errorf("after(): %w", err))
		}
	}

	if result.Failed() {
		info("%s  %s %s", redText("✘"), test.request.Method, test.request.URL.Path)
		for _, err := range result.Errors {
			problem("   %s", err)
		}

		if t != nil {
			t.Run(test.DisplayName(), func(t *testing.T) {
				for _, err := range result.Errors {
					t.Log(err)
				}

				t.FailNow()
			})
		}
	}

	return result, err
}

func (r *TestRunner) Validate() error {
	if r.BaseURL == "" {
		return fmt.Errorf("BaseURL is required")
	} else if r.BaseURL[len(r.BaseURL)-1] == '/' {
		return fmt.Errorf("BaseURL must not end with a slash")
	}

	return nil
}

func parseResponseBody(body []byte) Stringable {
	if len(body) > 0 {
		var bodyMap JSONObject
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			return bodyMap
		}

		var bodyArray JSONArray
		if err := json.Unmarshal(body, &bodyArray); err == nil {
			return bodyArray
		}

		return String(body)
	}

	return nil
}

// validateTests validates a set of tests.
func validateTests(tests []*TestCase) bool {
	valid := true
	for _, test := range tests {
		if err := test.Validate(); err != nil {
			problem("test case %q is invalid: %s", test.DisplayName(), err)
			valid = false
		}
	}

	return valid
}

// RunTest runs a single test using the provided base URL and the default TestRunner.
func RunTest(baseURL string, test *TestCase) {
	RunTestsT(nil, baseURL, []*TestCase{test})
}

// RunTestT runs a single test within a Go testing context using the provided
// base URL and the default TestRunner.
func RunTestT(t *testing.T, baseURL string, test *TestCase) {
	RunTestsT(t, baseURL, []*TestCase{test})
}

// RunTests runs a set of tests using the provided base URL and the default TestRunner.
func RunTests(baseURL string, tests []*TestCase) {
	RunTestsT(nil, baseURL, tests)
}

// RunTests runs a set of tests within a Go testing context using the provided
// base URL and the default TestRunner.
func RunTestsT(t GoTestContext, baseURL string, tests []*TestCase) {
	NewTestRunner(baseURL).RunTestsT(t, tests)
}
