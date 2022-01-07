package mt

import (
	"fmt"
	"net/http"
	"sort"

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
		if err := compareStatus(tc.Expectations.Status, r.Status); err != nil {
			r.addFailures(err)
		}
	}

	if tc.Expectations.Headers != nil {
		if errs := compareHeaders(tc.Expectations.Headers, r.Headers); len(errs) > 0 {
			r.addFailures(errs...)
		}
	}

	if tc.Expectations.Body != nil {
		body := toInterface(r.Body)
		for _, err := range expect.CompareValues(tc.Expectations.Body, body, tc.Expectations.WantExactJSONBody) {
			err.PushField("body")
			r.addFailures(err)
		}
	}
}

// Compares a set of expected headers against a set of actual headers,
func compareHeaders(expected http.Header, actual http.Header) []error {
	var errs []error
	for key, expectedValues := range expected {
		actualValues, ok := actual[key]
		if !ok {
			errs = append(errs, fmt.Errorf("expected header %q, got nothing", key))
			continue
		}

		sort.Strings(expectedValues)
		sort.Strings(actualValues)

		for _, expectedValue := range expectedValues {
			found := false
			for _, actualValue := range actualValues {
				if actualValue == expectedValue {
					found = true
					break
				}
			}

			if !found {
				errs = append(errs, fmt.Errorf("expected header %q to contain %q, got %q", key, expectedValue, actualValues))
			}
		}
	}

	return errs
}

// Compares an expected status code to an actual status code.
func compareStatus(expected, actual int) error {
	if expected != actual {
		return fmt.Errorf(`expected status %d, got %d`, expected, actual)
	}
	return nil
}
