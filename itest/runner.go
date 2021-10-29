package itest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
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
	envTimeoutStr := os.Getenv("ITEST_DEFAULT_TEST_TIMEOUT")
	if envTimeoutStr != "" {
		if timeout, err := time.ParseDuration(envTimeoutStr); err == nil {
			defaultRequestTimeout = timeout
		} else {
			color.HiYellow("invalid ITEST_DEFAULT_TEST_TIMEOUT value %q in environment, using default of %s",
				envTimeoutStr, defaultRequestTimeoutStr)
		}
	}
}

// A TestRunner runs a set of tests against an HTTP endpoint or handler.
//
// Use NewEndpointTester() to create a test runner that makes real HTTP
// requests to an actual server (local or remote). This is typically used
// to defice and run E2E tests against a running web service.
//
// Use NewHandlerTester() to create a test runner to perform in-memory
// functional tests against a Go HTTP handler (such as a router or mux).
// This is typically used for unit testing.
// .
type TestRunner struct {
	// ContinueOnFailure indicates whether the test runner should continue
	// executing further tests after a failure.
	//
	// Default is false.
	ContinueOnFailure bool

	// TestTimeout the the amount of time to wait for a test to complete.
	//
	// Default is 10 seconds.
	TestTimeout time.Duration

	// Internal
	baseURL      string
	client       *http.Client
	handler      http.Handler
	outputWriter *ColumnWriter
}

// NewEndpointTester creates a new TestRunner for testing an HTTP endpoint.
// targeting a base URL.
func NewEndpointTester(baseURL string) *TestRunner {
	return (&TestRunner{}).WithBaseURL(baseURL)
}

// NewHandlerTester creates a new TestRunner for testing an HTTP handler.
func NewHandlerTester(handler http.Handler) *TestRunner {
	return (&TestRunner{}).WithHTTPHandler(handler)
}

// WithBaseURL sets the base URL for the runner and returns the TestRunner.
func (r *TestRunner) WithBaseURL(baseURL string) *TestRunner {
	r.baseURL = baseURL
	return r
}

// WithHTTPHandler sets the HTTP handler for the runner and returns the
// TestRunner.
func (r *TestRunner) WithHTTPHandler(handler http.Handler) *TestRunner {
	r.handler = handler
	return r
}

// WithContinueOnFailure sets the ContinueOnFailure field of the TestRunner and
// returns the TestRunner.
func (r *TestRunner) WithContinueOnFailure(continueOnFailure bool) *TestRunner {
	r.ContinueOnFailure = continueOnFailure
	return r
}

// WithHTTPClient sets the HTTP client used for HTTP requests and  returns the
// TestRunner.
func (r *TestRunner) WithHTTPClient(client *http.Client) *TestRunner {
	r.client = client
	return r
}

// WithRequestTimeout sets the RequestTimeout field of the TestRunner and returns
// the TestRunner.
func (r *TestRunner) WithRequestTimeout(timeout time.Duration) *TestRunner {
	r.TestTimeout = timeout
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
func (r *TestRunner) RunTestsT(t *testing.T, tests []*TestCase) ([]*TestCaseResult, error) {
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("invalid test runner: %w", err)
	}

	if !validateTests(tests) {
		return nil, fmt.Errorf("one or more test cases failed validation")
	}

	outputTarget := io.Discard
	if t == nil {
		outputTarget = os.Stdout
	}

	r.outputWriter = NewColumnWriter(outputTarget, 5, 2)

	if r.mode() == modeBaseURL {
		fmt.Printf("running %d tests for %s\n", len(tests), underline(r.baseURL))
	} else {
		fmt.Printf("running %d tests for %v\n", len(tests), underline("HTTP handler"))
	}

	var executed, passed, failed, skipped int
	results := []*TestCaseResult{}
	for _, test := range tests {
		result, err := r.runTest(t, test)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
		executed++

		if result.Failed() {
			failed++

			if !r.ContinueOnFailure {
				r.outputWriter.Flush()
				color.HiYellow("skipping remaininig tests")
				skipped = len(tests) - executed
				break
			}

		} else {
			passed++
		}
	}

	r.outputWriter.Flush()

	totalExecutionTime := time.Duration(0)
	for _, result := range results {
		totalExecutionTime += result.executionTime
	}

	fmt.Printf("%d passed, %d failed, %d skipped %s\n", passed, failed, skipped,
		faintFG(fmt.Sprintf("in %s", totalExecutionTime.String())))

	return results, nil
}

func (r *TestRunner) runTest(t *testing.T, test *TestCase) (*TestCaseResult, error) {
	result := &TestCaseResult{
		TestCase: test,
		Errors:   []error{},
	}

	startTime := time.Now()

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
	} else if r.TestTimeout > 0 {
		timeout = r.TestTimeout
	}

	var reqBody []byte
	if test.RequestBody != nil {
		var err error
		reqBody, err = json.Marshal(test.RequestBody)
		if err != nil {
			result.addErrors(fmt.Errorf("error marshaling request body: %w", err))
			return result, nil
		}
	}

	if test.request == nil {
		req, cancel, err := createRequest(
			test.Method,
			r.baseURL+test.Path,
			test.QueryParams,
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
	if r.mode() == modeBaseURL {
		result.Status, result.Headers, result.Body, err = doRequest(r.client, test.request)
		if err != nil {
			debug("%s: failed to execute HTTP request: %s", test.DisplayName(), err)
			result.addErrors(fmt.Errorf("failed to execute HTTP request: %w", err))
			return nil, err
		}
	} else if r.mode() == modeHandler {
		result.Status, result.Headers, result.Body, err = handleRequest(r.handler, test.request)
		if err != nil {
			debug("%s: failed to handle HTTP request: %s", test.DisplayName(), err)
			result.addErrors(fmt.Errorf("failed to handle HTTP request: %w", err))
			return nil, err
		}
	}

	result.validateExpectations()

	if test.AfterFunc != nil {
		debug("%s: running after()", test.DisplayName())
		if err := test.AfterFunc(); err != nil {
			result.addErrors(fmt.Errorf("after(): %w", err))
		}
	}

	result.executionTime = time.Since(startTime)

	queryParamsStr := ""
	if l := len(test.request.URL.Query()); l > 0 {
		queryParamsStr = fmt.Sprintf("? (%d params)", l)
	}

	if result.Failed() {
		r.outputWriter.PrintColumns(
			redFG(" ✘"),
			whiteFG(test.Description),
			blueBG(fmt.Sprintf("%7s ", test.request.Method)),
			test.request.URL.Path+queryParamsStr,
			faintFG(result.executionTime.String()))

		for _, err := range result.Errors {
			r.outputWriter.PrintColumns(redFG(""), redFG(fmt.Sprintf("  %s", err)), blueBG(""), "", faintFG(""))
		}

		if t != nil {
			t.Run(test.DisplayName(), func(t *testing.T) {
				for _, err := range result.Errors {
					t.Log(err)
				}

				t.FailNow()
			})
		}
	} else {
		r.outputWriter.PrintColumns(
			greenFG(" ✔"),
			whiteFG(test.Description),
			blueBG(fmt.Sprintf("%7s ", test.request.Method)),
			test.request.URL.Path+queryParamsStr,
			faintFG(result.executionTime.String()))
	}

	return result, err
}

func (r *TestRunner) Validate() error {
	if r.mode() == modeNone {
		return fmt.Errorf("runner must be configured with a base URL or HTTP handler")
	} else if r.mode() == modeBaseURL {
		if r.baseURL[len(r.baseURL)-1] == '/' {
			return fmt.Errorf("base URL must not end with a slash")
		}
	}

	if r.client == nil {
		r.client = http.DefaultClient
	}

	return nil
}

func (r *TestRunner) mode() int {
	if r.handler != nil {
		return modeHandler
	} else if r.baseURL != "" {
		return modeBaseURL
	}

	return modeNone
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

// validateTests validates a set of tests.
func validateTests(tests []*TestCase) bool {
	valid := true
	for _, test := range tests {
		if err := test.Validate(); err != nil {
			fmt.Printf("test case %q is invalid: %s\n", test.DisplayName(), err)
			valid = false
		}
	}

	return valid
}

// TestEndpoint runs a set of tests using the provided base URL and the default TestRunner.
func TestEndpoint(baseURL string, tests []*TestCase) {
	TestEndpointT(nil, baseURL, tests)
}

// RunTests runs a set of tests within a Go testing context using the provided
// base URL and the default TestRunner.
func TestEndpointT(t *testing.T, baseURL string, tests []*TestCase) {
	NewEndpointTester(baseURL).RunTestsT(t, tests)
}

// TestEndpoint runs a set of tests using the provided base URL and the default TestRunner.
func TestHandler(handler http.Handler, tests []*TestCase) {
	TestHandlerT(nil, handler, tests)
}

// RunTests runs a set of tests within a Go testing context using the provided
// base URL and the default TestRunner.
func TestHandlerT(t *testing.T, handler http.Handler, tests []*TestCase) {
	NewHandlerTester(handler).RunTestsT(t, tests)
}
