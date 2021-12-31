package mt

import (
	"net/http"

	"github.com/jefflinse/melatonin/expect"
)

// HTTPTestCaseResult represents the result of running a single test case.
type HTTPTestCaseResult struct {
	// Status is the HTTP status code returned in the response.
	Status int `json:"status"`

	// Headers is the HTTP response headers.
	Headers http.Header `json:"headers"`

	// Body is the HTTP response body.
	Body []byte `json:"body"`

	testCase *HTTPTestCase
	failures []error
}

// Failures returns a list of test case failures.
func (r *HTTPTestCaseResult) Failures() []error {
	return r.failures
}

// TestCase returns a reference to the test case that generated the result.
func (r *HTTPTestCaseResult) TestCase() TestCase {
	return r.testCase
}

func (r *HTTPTestCaseResult) addFailures(errs ...error) *HTTPTestCaseResult {
	if len(errs) == 0 {
		return r
	}

	r.failures = append(r.failures, errs...)
	return r
}

func (r *HTTPTestCaseResult) validateExpectations() {
	tc := r.TestCase().(*HTTPTestCase)
	if tc.Expectations.Status != 0 {
		if err := expect.Status(tc.Expectations.Status, r.Status); err != nil {
			r.addFailures(err)
		}
	}

	if tc.Expectations.Headers != nil {
		if errs := expect.Headers(tc.Expectations.Headers, r.Headers); len(errs) > 0 {
			r.addFailures(errs...)
		}
	}

	if tc.Expectations.Body != nil {
		body := toInterface(r.Body)
		if errs := expect.Value(tc.Expectations.Body, body, tc.Expectations.WantExactJSONBody); len(errs) > 0 {
			r.addFailures(errs...)
		}
	}
}
