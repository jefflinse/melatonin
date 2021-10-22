package itest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
func (r *TestRunner) RunTests(tests []*TestCase) {
	if err := r.Validate(); err != nil {
		fatal("invalid test runner:", err)
		return
	}

	if !ValidateTests(tests) {
		fatal("one or more test cases failed validation")
	}

	for _, test := range tests {
		err := r.RunTest(test)
		r.excuted++
		if err != nil {
			r.failed++
			info("%s  %s %s", redText("FAIL"), test.Method, test.Path)
			problem("     %s", err)
			if !r.ContinueOnFailure {
				warn("skipping remaininig tests")
				r.skipped = len(tests) - r.excuted
				break
			}

		} else {
			r.passed++
			info("%s  %s %s", greenText("OK"), test.Method, test.Path)
		}
	}

	info("\n%d passed, %d failed, %d skipped", r.passed, r.failed, r.skipped)
}

// RunTest runs a single test.
func (r *TestRunner) RunTest(test *TestCase) error {
	if test.Setup != nil {
		debug("%s: running setup\n", test.DisplayName())
		if err := test.Setup(); err != nil {
			return fmt.Errorf("test %q failed setup: %s", test.DisplayName(), err)
		}
	}

	if test.request == nil {
		req, err := r.createRequest(test.Method, r.BaseURL+test.Path, test.RequestHeaders, test.RequestBody)
		if err != nil {
			return fmt.Errorf("test %q failed to create HTTP request: %s", test.DisplayName(), err)
		}

		test.request = req
	}

	status, body, err := r.doRequest(test.request)
	if err != nil {
		return fmt.Errorf("test %q failed to perform HTTP request: %s", test.DisplayName(), err)
	}

	if status != test.WantStatus {
		return fmt.Errorf("expected status %d, got %d", test.WantStatus, status)
	}

	return assertTypeAndValue("", test.WantBody, body)
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

func (r *TestRunner) doRequest(req *http.Request) (int, JSONMap, error) {
	debug("%s %s\n", req.Method, req.URL.RawPath)
	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return -1, nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, err
	}

	var bodyMap JSONMap
	if len(b) > 0 {
		if err := json.Unmarshal(b, &bodyMap); err == nil {
			debug("%s\n", string(b))
			return resp.StatusCode, bodyMap, nil
		}
	}

	return resp.StatusCode, nil, nil
}

// RunTests runs a set of tests using the provided base URL and the default TestRunner.
func RunTests(baseURL string, tests []*TestCase) {
	DefaultTestRunner().WithBaseURL(baseURL).RunTests(tests)
}
