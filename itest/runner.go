package itest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// TestRunner contains configuration for running tests.
type TestRunner struct {
	// BaseURL is the base URL for the API, including the port.
	//
	// Examples:
	//   http://localhost:8080
	//   https://api.example.com
	BaseURL string

	// ContinueOnFailure indicates whether the test runner should continue
	// executing further tests after a failure. Defaults to false.
	ContinueOnFailure bool

	// HTTPClient is the HTTP client to use for requests.
	// If left unset, http.DefaultClient will be used.
	HTTPClient *http.Client

	excuted int
	passed  int
	failed  int
	skipped int
}

// DefaultTestRunner creates a new TestRunner with the default settings.
func DefaultTestRunner() *TestRunner {
	return &TestRunner{
		ContinueOnFailure: false,
		HTTPClient:        http.DefaultClient,
	}
}

// WithBaseURL sets the BaseURL field of the TestRunner and returns the TestRunner.
func (r *TestRunner) WithBaseURL(baseURL string) *TestRunner {
	r.BaseURL = baseURL
	return r
}

// WithContinueOnFailure sets the ContinueOnFailure field of the TestRunner and
// returns the TestRunner.
func (r TestRunner) WithContinueOnFailure(continueOnFailure bool) *TestRunner {
	r.ContinueOnFailure = continueOnFailure
	return &r
}

// WithHTTPClient sets the HTTPClient field of the TestRunner and returns the
// TestRunner.
func (r *TestRunner) WithHTTPClient(client *http.Client) *TestRunner {
	r.HTTPClient = client
	return r
}

// RunTests runs a set of tests.
func (r *TestRunner) RunTests(tests []*TestCase) (results []*TestCaseResult) {
	results = []*TestCaseResult{}

	if err := r.Validate(); err != nil {
		fatal("invalid test runner:", err)
		return
	}

	if !ValidateTests(tests) {
		fatal("one or more test cases failed validation")
	}

	info("running %d tests for %s", len(tests), r.BaseURL)
	for _, test := range tests {
		result := r.RunTest(test)
		r.excuted++
		if len(result.Errors) > 0 {
			r.failed++
			info("%s  %s %s", redText("✘"), test.Method, test.Path)
			for _, err := range result.Errors {
				problem("   %s", err)
			}

			if !r.ContinueOnFailure {
				warn("skipping remaininig tests")
				r.skipped = len(tests) - r.excuted
				break
			}

		} else {
			r.passed++
			info("%s  %s %s", greenText("✔"), test.Method, test.Path)
		}
	}

	info("\n%d passed, %d failed, %d skipped", r.passed, r.failed, r.skipped)
	return
}

// RunTest runs a single test.
func (r *TestRunner) RunTest(test *TestCase) (result *TestCaseResult) {
	result = &TestCaseResult{
		TestCase: test,
		Errors:   []error{},
	}

	if test.Before != nil {
		debug("%s: running before()", test.DisplayName())
		if err := test.Before(); err != nil {
			result.AddError(fmt.Errorf("before(): %w", err))
			return
		}
	}

	if test.request == nil {
		req, err := r.createRequest(test.Method, r.BaseURL+test.Path, test.RequestHeaders, test.RequestBody)
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

	if test.After != nil {
		debug("%s: running after()", test.DisplayName())
		if err := test.After(); err != nil {
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
	} else if r.HTTPClient == nil {
		return fmt.Errorf("HTTPClient is required")
	}

	return nil
}

// ValidateTests validates a set of tests
func ValidateTests(tests []*TestCase) bool {
	valid := true
	for _, test := range tests {
		if err := test.Validate(); err != nil {
			problem("test case %q is invalid: %s", test.DisplayName(), err)
			valid = false
		}
	}

	return valid
}

func (r *TestRunner) createRequest(method, uri string, headers http.Header, body Stringable) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader([]byte(body.String()))
	}

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		req.Header = headers
	} else {
		req.Header = http.Header{}
	}

	return req, nil
}

func (r *TestRunner) doRequest(req *http.Request) (int, Stringable, error) {
	debug("%s %s\n", req.Method, req.URL.String())
	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return -1, nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, err
	}

	responseString := strings.TrimSuffix(string(b), "\n")
	debug("%d\n%s", resp.StatusCode, responseString)

	if len(b) > 0 {
		var bodyMap JSONMap
		if err := json.Unmarshal(b, &bodyMap); err == nil {
			return resp.StatusCode, bodyMap, nil
		}

		var bodyArray JSONArray
		if err := json.Unmarshal(b, &bodyArray); err == nil {
			return resp.StatusCode, bodyArray, nil
		}

		return resp.StatusCode, String(responseString), err
	}

	return resp.StatusCode, nil, nil
}

// RunTests runs a set of tests using the provided base URL and the default TestRunner.
func RunTests(baseURL string, tests []*TestCase) {
	DefaultTestRunner().WithBaseURL(baseURL).RunTests(tests)
}
