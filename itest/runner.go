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
}

// DefaultTestRunner creates a new TestRunner with the default settings.
func DefaultTestRunner() *TestRunner {
	return &TestRunner{
		ContinueOnFailure: false,
		HTTPClient:        http.DefaultClient,
	}
}

// WithContinueOnFailure sets the ContinueOnFailure field of the TestRunner and
// returns the TestRunner.
func (r TestRunner) WithContinueOnFailure(continueOnFailure bool) *TestRunner {
	r.ContinueOnFailure = continueOnFailure
	return &r
}

// WithBaseURL sets the BaseURL field of the TestRunner and returns the TestRunner.
func (r *TestRunner) WithBaseURL(baseURL string) *TestRunner {
	r.BaseURL = baseURL
	return r
}

// WithHTTPClient sets the HTTPClient field of the TestRunner and returns the
// TestRunner.
func (r *TestRunner) WithHTTPClient(client *http.Client) *TestRunner {
	r.HTTPClient = client
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
			fmt.Printf("%s: %s\n", test.Name, err)
			if !r.ContinueOnFailure {
				break
			}
		}
	}
}

// RunTest runs a single test.
func (r *TestRunner) RunTest(test TestCase) error {
	if test.Setup != nil {
		test.Setup()
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
	} else if r.HTTPClient.Transport == nil {
		return fmt.Errorf("HTTPClient.Transport is required")
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
		if err := json.Unmarshal(b, &bodyMap); err != nil {
			return -1, nil, err
		}
	}

	return resp.StatusCode, bodyMap, nil
}

// RunTests runs a set of tests using the provided base URL and the default TestRunner.
func RunTests(baseURL string, tests []TestCase) {
	DefaultTestRunner().WithBaseURL(baseURL).RunTests(tests)
}
