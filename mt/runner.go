package mt

import (
	"fmt"
	"testing"
	"time"
)

// A TestRunner runs a set of tests.
type TestRunner struct {
	// ContinueOnFailure indicates whether the test runner should continue
	// executing further tests after a test encounters a failure.
	//
	// Default is false.
	ContinueOnFailure bool

	// TestTimeout the the amount of time to wait for any single test to complete.
	//
	// Default is 10 seconds.
	TestTimeout time.Duration
}

// A RunResult contains information about a set of test cases run by a test runner.
type RunResult struct {
	// Group is a reference to the test group that was run.
	Group *TestGroup

	// TestResults is a list of test results corresponding to tests in the test group.
	TestResults []TestResult

	// TestDurations is a list of durations corresponding to results in the test results.
	TestDurations []time.Duration

	// Passed is the number of tests that passed.
	Passed int

	// Failed is the number of tests that failed.
	Failed int

	// Skipped is the number of tests that were skipped.
	Skipped int

	// Total is the total number of tests in the test group.
	Total int

	// Duration is the total duration of all tests in the test group.
	Duration time.Duration
}

// NewTestRunner creates a new TestRunner with default configuration.
func NewTestRunner() *TestRunner {
	return &TestRunner{
		ContinueOnFailure: cfg.ContinueOnFailure,
		TestTimeout:       10 * time.Second,
	}
}

// WithContinueOnFailure sets the ContinueOnFailure field of the TestRunner and
// returns the TestRunner.
func (r *TestRunner) WithContinueOnFailure(continueOnFailure bool) *TestRunner {
	r.ContinueOnFailure = continueOnFailure
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
// To run tests within a Go test context, use RunTestsT().
func (r *TestRunner) RunTests(tests []TestCase) RunResult {
	return r.RunTestsT(nil, tests)
}

// RunTestsT runs a set of tests within a Go test context.
//
// To run tests standalone to print or examine results, use RunTests().
func (r *TestRunner) RunTestsT(t *testing.T, tests []TestCase) RunResult {
	group := NewTestGroup(fmt.Sprintf("%d tests in sequence", len(tests))).Add(tests...)
	return r.RunTestGroupT(t, group)
}

// RunTestGroup runs a test group.
//
// To run a test group within a Go test context, use RunTestGroupT().
func (r *TestRunner) RunTestGroup(group *TestGroup) RunResult {
	return r.RunTestGroupT(nil, group)
}

// RunTestGroupT runs a test group within the context of a Go test.
//
// To run tests as a standalone binary without a testing context, use RunTests().
func (r *TestRunner) RunTestGroupT(t *testing.T, group *TestGroup) RunResult {
	runResult := RunResult{
		Group: group,
	}

	for _, test := range group.TestCases {
		start := time.Now()
		testResult := test.Execute()
		duration := time.Since(start)
		runResult.TestResults = append(runResult.TestResults, testResult)
		runResult.TestDurations = append(runResult.TestDurations, duration)
		runResult.Total++
		runResult.Duration += duration

		if len(testResult.Failures()) > 0 {
			runResult.Failed++
			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					for _, err := range testResult.Failures() {
						t.Log(err)
					}

					t.FailNow()
				})
			}

			if !r.ContinueOnFailure {
				runResult.Skipped = len(group.TestCases) - runResult.Total
				break
			}

		} else {
			runResult.Passed++
			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					t.Log(testResult.TestCase().Description())
				})
			}
		}
	}

	return runResult
}

// RunTests runs a set of tests using the default test runner.
func RunTests(tests []TestCase) RunResult {
	return NewTestRunner().RunTests(tests)
}

// RunTestsT runs a set of tests within a Go test context
// using the default test runner.
func RunTestsT(t *testing.T, tests []TestCase) RunResult {
	return NewTestRunner().RunTestsT(t, tests)
}

// RunTestGroup runs a test group using the default test runner.
func RunTestGroup(group *TestGroup) RunResult {
	return NewTestRunner().RunTestGroup(group)
}

// RunTestGroupT runs a test group within the context of a Go test
// using the default test runner.
func RunTestGroupT(t *testing.T, group *TestGroup) RunResult {
	return NewTestRunner().RunTestGroupT(t, group)
}
