package itest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	Log(args ...interface{})
	Fail()
	FailNow()
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

	// T is a testing.T instance to use for running the tests as standard Go tests.
	T GoTestContext
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

// WithT sets the T field of the TestRunner and returns the TestRunner.
func (r *TestRunner) WithT(t GoTestContext) *TestRunner {
	r.T = t
	return r
}

// RunTests runs a set of tests.
func (r *TestRunner) RunTests(tests []*TestCase) (results []*TestCaseResult) {
	results = []*TestCaseResult{}

	if err := r.Validate(); err != nil {
		fatal("invalid test runner:", err)
		return
	}

	if r.HTTPClient == nil {
		debug("using default HTTP client")
		r.HTTPClient = http.DefaultClient
	}

	if !validateTests(tests) {
		fatal("one or more test cases failed validation")
	}

	info("running %d tests for %s", len(tests), r.BaseURL)
	var executed, passed, failed, skipped int
	for _, test := range tests {
		result := r.RunTest(test)
		if r.T != nil {
			r.T.Run(test.DisplayName(), func(t *testing.T) {
				if len(result.Errors) > 0 {
					for _, err := range result.Errors {
						t.Log(err)
					}

					t.FailNow()
				}
			})
		}

		results = append(results, result)
		executed++

		if len(result.Errors) > 0 {
			failed++
			info("%s  %s %s", redText("✘"), test.request.Method, test.request.URL.Path)
			for _, err := range result.Errors {
				problem("   %s", err)
			}

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
	return
}

// RunTest runs a single test.
func (r *TestRunner) RunTest(test *TestCase) (result *TestCaseResult) {
	result = &TestCaseResult{
		TestCase: test,
		Errors:   []error{},
	}

	if test.BeforeFunc != nil {
		debug("%s: running before()", test.DisplayName())
		if err := test.BeforeFunc(); err != nil {
			result.AddError(fmt.Errorf("before(): %w", err))
			return
		}
	}

	timeout := defaultRequestTimeout
	if test.Timeout > 0 {
		timeout = test.Timeout
	} else if r.RequestTimeout > 0 {
		timeout = r.RequestTimeout
	}

	if test.request == nil {
		req, cancel, err := r.createRequest(
			test.Method,
			r.BaseURL+test.Path,
			test.RequestHeaders,
			test.RequestBody,
			timeout)
		defer cancel()
		if err != nil {
			result.AddError(fmt.Errorf("failed to create HTTP request: %w", err))
			return
		}

		test.request = req
	}

	status, body, err := r.doRequest(test.request)
	if err != nil {
		result.AddError(fmt.Errorf("failed to perform HTTP request: %w", err))
		return
	}

	if test.WantStatus != 0 {
		if err := expectStatus(test.WantStatus, status); err != nil {
			result.AddError(err)
		}
	}

	if test.WantBody != nil {
		if err := expect("body", test.WantBody, body); err != nil {
			result.AddError(err)
		}
	}

	if test.AfterFunc != nil {
		debug("%s: running after()", test.DisplayName())
		if err := test.AfterFunc(); err != nil {
			result.AddError(fmt.Errorf("after(): %w", err))
		}
	}

	return
}

func (r *TestRunner) Validate() error {
	if r.BaseURL == "" {
		return fmt.Errorf("BaseURL is required")
	} else if r.BaseURL[len(r.BaseURL)-1] == '/' {
		return fmt.Errorf("BaseURL must not end with a slash")
	}

	return nil
}

func (r *TestRunner) createRequest(method, uri string,
	headers http.Header,
	body Stringable,
	timeout time.Duration) (*http.Request, context.CancelFunc, error) {

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader([]byte(body.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	req, err := http.NewRequestWithContext(ctx, method, uri, reader)
	if err != nil {
		return nil, cancel, err
	}

	if headers != nil {
		req.Header = headers
	} else {
		req.Header = http.Header{}
	}

	return req, cancel, nil
}

func (r *TestRunner) doRequest(req *http.Request) (int, Stringable, error) {
	debug("%s %s", req.Method, req.URL.String())
	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return -1, nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, err
	}

	body := parseResponseBody(b)
	if body != nil {
		debug("%d\n%s", resp.StatusCode, body.String())
	} else {
		debug("%d", resp.StatusCode)
	}

	debug("\n")

	return resp.StatusCode, body, nil
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

// RunTests runs a set of tests using the provided base URL and the default TestRunner.
func RunTests(baseURL string, tests []*TestCase) {
	NewTestRunner(baseURL).RunTests(tests)
}

// RunTests runs a set of tests using the provided base URL and the default TestRunner.
func RunTestsT(t GoTestContext, baseURL string, tests []*TestCase) {
	NewTestRunner(baseURL).WithT(t).RunTests(tests)
}
