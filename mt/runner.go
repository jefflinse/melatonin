package mt

import (
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

// A TestRunResult contains information about a completed test case run.
type TestRunResult struct {
	TestCase   TestCase      `json:"test"`
	TestResult TestResult    `json:"result"`
	StartedAt  time.Time     `json:"started_at"`
	EndedAt    time.Time     `json:"finished_at"`
	Duration   time.Duration `json:"duration"`
}

// A GroupRunResult contains information about a completed set of test cases run by a test runner.
type GroupRunResult struct {
	// Group is a reference to the test group that was run.
	Group *TestGroup `json:"-"`

	// Results is a list of test run results corresponding to tests in the test group.
	Results []TestRunResult `json:"results"`

	// Passed is the number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the number of tests that failed.
	Failed int `json:"failed"`

	// Skipped is the number of tests that were skipped.
	Skipped int `json:"skipped"`

	// Total is the total number of tests in the test group.
	Total int `json:"total"`

	// Duration is the total duration of all tests in the test group.
	Duration time.Duration `json:"duration`
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
func (r *TestRunner) RunTests(tests []TestCase) GroupRunResult {
	return r.RunTestsT(nil, tests)
}

// RunTestsT runs a set of tests within a Go test context.
//
// To run tests standalone to print or examine results, use RunTests().
func (r *TestRunner) RunTestsT(t *testing.T, tests []TestCase) GroupRunResult {
	group := NewTestGroup("").Add(tests...)
	return r.RunTestGroupT(t, group)
}

// RunTestGroup runs a test group.
//
// To run a test group within a Go test context, use RunTestGroupT().
func (r *TestRunner) RunTestGroup(group *TestGroup) GroupRunResult {
	return r.RunTestGroupT(nil, group)
}

// RunTestGroupT runs a test group within the context of a Go test.
//
// To run tests as a standalone binary without a testing context, use RunTests().
func (r *TestRunner) RunTestGroupT(t *testing.T, group *TestGroup) GroupRunResult {
	groupRunResult := GroupRunResult{
		Group: group,
	}

	for _, test := range group.TestCases {
		start := time.Now()
		testResult := test.Execute()
		end := time.Now()
		runResult := TestRunResult{
			TestCase:   test,
			TestResult: testResult,
			StartedAt:  start,
			EndedAt:    end,
			Duration:   end.Sub(start),
		}

		groupRunResult.Total++
		groupRunResult.Duration += runResult.Duration
		groupRunResult.Results = append(groupRunResult.Results, runResult)

		if len(testResult.Failures()) > 0 {
			groupRunResult.Failed++
			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					for _, err := range testResult.Failures() {
						t.Log(err)
					}

					t.FailNow()
				})
			}

			if !r.ContinueOnFailure {
				groupRunResult.Skipped = len(group.TestCases) - groupRunResult.Total
				break
			}

		} else {
			groupRunResult.Passed++
			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					t.Log(testResult.TestCase().Description())
				})
			}
		}
	}

	return groupRunResult
}

// RunTests runs a set of tests using the default test runner.
func RunTests(tests []TestCase) GroupRunResult {
	return NewTestRunner().RunTests(tests)
}

// RunTestsT runs a set of tests within a Go test context
// using the default test runner.
func RunTestsT(t *testing.T, tests []TestCase) GroupRunResult {
	return NewTestRunner().RunTestsT(t, tests)
}

// RunTestGroup runs a test group using the default test runner.
func RunTestGroup(group *TestGroup) GroupRunResult {
	return NewTestRunner().RunTestGroup(group)
}

// RunTestGroupT runs a test group within the context of a Go test
// using the default test runner.
func RunTestGroupT(t *testing.T, group *TestGroup) GroupRunResult {
	return NewTestRunner().RunTestGroupT(t, group)
}
