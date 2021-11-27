package mt

import (
	"testing"
	"time"
)

type TestRunner struct {
	// ContinueOnFailure indicates whether the test runner should continue
	// executing further tests after a failure.
	//
	// Default is false.
	ContinueOnFailure bool

	// TestTimeout the the amount of time to wait for a test to complete.
	//
	// Default is 10 seconds.
	TestTimeout time.Duration
}

type RunResult struct {
	TestResults   []TestResult
	TestDurations []time.Duration
	Passed        int
	Failed        int
	Skipped       int
	Total         int
	Duration      time.Duration
}

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
// To run tests within the context of a Go test, use RunTestsT().
func (r *TestRunner) RunTests(tests []TestCase) RunResult {
	return r.RunTestsT(nil, tests)
}

// RunTests runs a set of tests within the context of a Go test.
//
// To run tests as a standalone binary without a testing context, use RunTests().
func (r *TestRunner) RunTestsT(t *testing.T, tests []TestCase) RunResult {
	runResult := RunResult{}
	for _, test := range tests {
		testResult, duration := runTest(test)
		runResult.TestResults = append(runResult.TestResults, testResult)
		runResult.TestDurations = append(runResult.TestDurations, duration)
		runResult.Total++
		runResult.Duration += duration

		if len(testResult.Errors()) > 0 {
			runResult.Failed++
			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					for _, err := range testResult.Errors() {
						t.Log(err)
					}

					t.FailNow()
				})
			}

			if !r.ContinueOnFailure {
				runResult.Skipped = len(tests) - runResult.Total
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

func runTest(test TestCase) (TestResult, time.Duration) {
	start := time.Now()
	result := test.Execute()
	duration := time.Since(start)
	return result, duration
}

func RunTests(tests []TestCase) RunResult {
	return NewTestRunner().RunTests(tests)
}

func RunTestsT(t *testing.T, tests []TestCase) RunResult {
	return NewTestRunner().RunTestsT(t, tests)
}
