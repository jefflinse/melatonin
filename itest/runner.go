package itest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/fatih/color"
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

	// Verbose causes the running to print out additional information for each test.
	// Defaults to false.
	Verbose bool
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

// WithVerbose sets the Verbose field of the TestRunner and returns the TestRunner.
func (r *TestRunner) WithVerbose(verbose bool) *TestRunner {
	r.Verbose = verbose
	return r
}

// RunTests runs a set of tests.
func (r *TestRunner) RunTests(tests []TestCase) {
	if err := r.Validate(); err != nil {
		fmt.Println("invalid test runner:", err)
		return
	}

	for _, test := range tests {
		if err := r.RunTest(test); err != nil {
			color.Set(color.FgRed)
			fmt.Printf("%s: %s\n", test.Name, err)
			color.Unset()
			if !r.ContinueOnFailure {
				break
			}
		} else {
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("%s  %s %s\n", green("OK"), test.Method, test.URI)
		}
	}
}

// RunTest runs a single test.
func (r *TestRunner) RunTest(test TestCase) error {
	if test.Setup != nil {
		r.verbose("%s: running setup\n", test.Name)
		if err := test.Setup(); err != nil {
			return fmt.Errorf("test %q failed setup: %s", test.Name, err)
		}
	}

	status, body, err := r.doRequest(test.Method, r.BaseURL+test.URI, test.RequestBody)
	if err != nil {
		return fmt.Errorf("unexpeceted error while running test %q: %s", test.Name, err)
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
func ValidateTests(tests []TestCase) bool {
	valid := true
	for _, test := range tests {
		if err := test.Validate(); err != nil {
			fmt.Printf("test case %q is invalid: %s", test.Name, err)
			valid = false
		}
	}

	return valid
}

func (r *TestRunner) doRequest(method, uri string, body Stringable) (int, JSONMap, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader([]byte(body.String()))
	}

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return -1, nil, err
	}

	r.verbose("%s %s\n", method, uri)
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
			r.verbose("%s\n", string(b))
			return resp.StatusCode, bodyMap, nil
		}
	}

	return resp.StatusCode, nil, nil
}

func (r *TestRunner) verbose(format string, args ...interface{}) {
	if r.Verbose {
		color.Set(color.FgBlue)
		fmt.Printf(format, args...)
		color.Unset()
	}
}

// RunTests runs a set of tests using the provided base URL and the default TestRunner.
func RunTests(baseURL string, tests []TestCase) {
	DefaultTestRunner().WithBaseURL(baseURL).WithVerbose(true).RunTests(tests)
}
