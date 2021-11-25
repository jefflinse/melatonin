package mt

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fatih/color"
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
		result, duration, err := runTestT(t, test)
		if err != nil {
			return nil, err
		}

		totalDuration += duration
		executed++

		if len(result.Errors()) > 0 {
			failed++
			r.printTestFailure(test, result, duration)

			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					for _, err := range result.Errors() {
						t.Log(err)
					}

					t.FailNow()
				})
			}

			if !r.ContinueOnFailure {
				color.HiYellow("skipping remaininig tests")
				skipped = len(tests) - executed
				break
			}

		} else {
			passed++
			if t == nil {
				r.printTestSuccess(test, result, duration)
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

func runTestT(t *testing.T, test TestCase) (TestResult, time.Duration, error) {
	start := time.Now()
	result, err := test.Execute(t)
	duration := time.Since(start)
	if err != nil {
		return nil, duration, err
	}

	return result, duration, nil
}

func RunTests(tests []TestCase) ([]TestResult, error) {
	return NewTestRunner().RunTests(tests)
}

func RunTestsT(t *testing.T, tests []TestCase) ([]TestResult, error) {
	return NewTestRunner().RunTestsT(t, tests)
}
