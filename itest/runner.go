package itest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
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
			color.HiYellow("invalid ITEST_DEFAULT_TEST_TIMEOUT value %q in environment, using default of %s",
				envTimeoutStr, defaultRequestTimeoutStr)
		}
	}
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

	// Internal
	outputWriter *ColumnWriter
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
func (r *TestRunner) RunTestsT(t *testing.T, tests []*TestCase) ([]*TestCaseResult, error) {
	results := []*TestCaseResult{}
	r.outputWriter = NewColumnWriter(os.Stdout, 5, 2)

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

	fmt.Printf("running %d tests for %s\n", len(tests), underline(r.BaseURL))
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

	fmt.Printf("\n%d passed, %d failed, %d skipped %s\n", passed, failed, skipped,
		faintFG(fmt.Sprintf("in %s", totalExecutionTime.String())))
	return results, nil
}

// RunTest runs a single test.
func (r *TestRunner) RunTest(test *TestCase) (*TestCaseResult, error) {
	return r.RunTestT(nil, test)
}

// RunTest runs a single test within a Go testing context.
func (r *TestRunner) RunTestT(t *testing.T, test *TestCase) (*TestCaseResult, error) {
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
	} else if r.RequestTimeout > 0 {
		timeout = r.RequestTimeout
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
			r.BaseURL+test.Path,
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

	result.executionTime = time.Since(startTime)

	queryParamsStr := ""
	if l := len(test.request.URL.Query()); l > 0 {
		queryParamsStr = fmt.Sprintf("? (%d params)", l)
	}

	if result.Failed() {
		r.outputWriter.PrintColumns(
			redFG("✘"),
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
			greenFG("✔"),
			whiteFG(test.Description),
			blueBG(fmt.Sprintf("%7s ", test.request.Method)),
			test.request.URL.Path+queryParamsStr,
			faintFG(result.executionTime.String()))
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
func RunTestsT(t *testing.T, baseURL string, tests []*TestCase) {
	NewTestRunner(baseURL).RunTestsT(t, tests)
}
