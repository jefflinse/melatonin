package mt

import (
	"fmt"
	"io"
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

	outputWriter *columnWriter
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
func (r *TestRunner) RunTests(tests []TestCase) ([]TestResult, error) {
	return r.RunTestsT(nil, tests)
}

// RunTests runs a set of tests within the context of a Go test.
//
// To run tests as a standalone binary without a testing context, use RunTests().
func (r *TestRunner) RunTestsT(t *testing.T, tests []TestCase) ([]TestResult, error) {
	outputTarget := io.Discard
	if t == nil {
		outputTarget = cfg.Stdout
	}

	r.outputWriter = newColumnWriter(outputTarget, 5, 2)

	var executed, passed, failed, skipped int
	results := []TestResult{}
	totalDuration := time.Duration(0)
	for _, test := range tests {
		result, duration := runTest(test)
		totalDuration += duration
		executed++

		if len(result.Errors()) > 0 {
			failed++
			if t == nil {
				r.printTestFailure(test, result, duration)
			} else {
				t.Run(test.Description(), func(t *testing.T) {
					for _, err := range result.Errors() {
						t.Log(err)
					}

					t.FailNow()
				})
			}

			if !r.ContinueOnFailure {
				skipped = len(tests) - executed
				break
			}

		} else {
			passed++
			if t == nil {
				r.printTestSuccess(test, result, duration)
			} else {
				t.Run(test.Description(), func(t *testing.T) {
					t.Log(result.TestCase().Description())
				})
			}
		}

		results = append(results, result)
	}

	r.outputWriter.Flush()
	r.outputWriter.PrintLine("%d passed, %d failed, %d skipped %s\n", passed, failed, skipped,
		faintFG(fmt.Sprintf("in %s", totalDuration.String())))

	return results, nil
}

func (r *TestRunner) printTestFailure(test TestCase, result TestResult, duration time.Duration) {
	r.outputWriter.PrintColumns(
		redFG(" ✘"),
		whiteFG(test.Description()),
		blueBG(fmt.Sprintf("%7s ", test.Action())),
		test.Target(),
		faintFG(duration.String()))

	for _, err := range result.Errors() {
		r.outputWriter.PrintColumns(redFG(""), redFG(fmt.Sprintf("  %s", err)), blueBG(""), "", faintFG(""))
	}
}

func (r *TestRunner) printTestSuccess(test TestCase, result TestResult, duration time.Duration) {
	r.outputWriter.PrintColumns(
		greenFG(" ✔"),
		whiteFG(test.Description()),
		blueBG(fmt.Sprintf("%7s ", test.Action())),
		test.Target(),
		faintFG(duration.String()))
}

func runTest(test TestCase) (TestResult, time.Duration) {
	start := time.Now()
	result := test.Execute()
	duration := time.Since(start)
	return result, duration
}

func RunTests(tests []TestCase) ([]TestResult, error) {
	return NewTestRunner().RunTests(tests)
}

func RunTestsT(t *testing.T, tests []TestCase) ([]TestResult, error) {
	return NewTestRunner().RunTestsT(t, tests)
}
