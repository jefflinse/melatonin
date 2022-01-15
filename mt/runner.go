package mt

import (
	"testing"
	"time"
)

const (
	// ExecuteTestsFirst causes the test runner to execute tests before subgroups.
	ExecuteTestsFirst = iota

	// ExecuteSubgroupsFirst causes the test runner to execute subgroups before tests.
	ExecuteSubgroupsFirst
)

// A TestRunner runs a set of tests.
type TestRunner struct {
	// ContinueOnFailure indicates whether the test runner should continue
	// executing further tests after a test encounters a failure.
	//
	// Default is false.
	ContinueOnFailure bool

	// GroupExecutionPriority indicates whether the test runner should execute
	// tests before or after subgroups.
	GroupExecutionPriority int

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

	// Results is a list of test run results for each test in the group that was run.
	TestResults []TestRunResult `json:"test_results,omitempty"`

	// GroupRunResults is a list of group run results for each subgroup that was run.
	SubgroupResults []*GroupRunResult `json:"group_results,omitempty"`

	// Passed is the number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the number of tests that failed.
	Failed int `json:"failed"`

	// Skipped is the number of tests that were skipped.
	Skipped int `json:"skipped"`

	// Total is the total number of tests in the test group.
	Total int `json:"total"`

	// Duration is the total duration of all tests in the test group.
	Duration time.Duration `json:"duration"`
}

// NewTestRunner creates a new TestRunner with default configuration.
func NewTestRunner() *TestRunner {
	return &TestRunner{
		ContinueOnFailure:      cfg.ContinueOnFailure,
		GroupExecutionPriority: ExecuteTestsFirst,
		TestTimeout:            10 * time.Second,
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
func (r *TestRunner) RunTests(tests ...TestCase) *GroupRunResult {
	return r.RunTestsT(nil, tests...)
}

// RunTestsT runs a set of tests within a Go test context.
//
// To run tests standalone to print or examine results, use RunTests().
func (r *TestRunner) RunTestsT(t *testing.T, tests ...TestCase) *GroupRunResult {
	group := NewTestGroup("").AddTests(tests...)
	return r.RunTestGroupT(t, group)
}

// RunTestGroup runs a test group.
//
// To run a test group within a Go test context, use RunTestGroupT().
func (r *TestRunner) RunTestGroup(group *TestGroup) *GroupRunResult {
	return r.RunTestGroupT(nil, group)
}

// RunTestGroupT runs a test group within the context of a Go test.
//
// To run tests as a standalone binary without a testing context, use RunTests().
func (r *TestRunner) RunTestGroupT(t *testing.T, group *TestGroup) *GroupRunResult {
	groupResult := &GroupRunResult{
		Group: group,
	}

	if group.BeforeFunc != nil {
		group.BeforeFunc()
	}

	if r.GroupExecutionPriority == ExecuteSubgroupsFirst {
		r.runSubgroups(t, groupResult)
	}

	for _, test := range group.Tests {
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

		groupResult.TestResults = append(groupResult.TestResults, runResult)
		groupResult.Total++
		groupResult.Duration += runResult.Duration

		if len(testResult.Failures()) > 0 {
			groupResult.Failed++
			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					for _, err := range testResult.Failures() {
						t.Log(err)
					}

					t.FailNow()
				})
			}

			if !r.ContinueOnFailure {
				groupResult.Skipped = len(group.Tests) - groupResult.Total
				break
			}

		} else {
			groupResult.Passed++
			if t != nil {
				t.Run(test.Description(), func(t *testing.T) {
					t.Log(testResult.TestCase().Description())
				})
			}
		}
	}

	if r.GroupExecutionPriority == ExecuteTestsFirst {
		r.runSubgroups(t, groupResult)
	}

	if group.AfterFunc != nil {
		group.AfterFunc()
	}

	return groupResult
}

func (r *TestRunner) runSubgroups(t *testing.T, groupResult *GroupRunResult) {
	for _, subgroup := range groupResult.Group.Subgroups {
		result := r.RunTestGroupT(t, subgroup)
		groupResult.SubgroupResults = append(groupResult.SubgroupResults, result)
		groupResult.Passed += result.Passed
		groupResult.Failed += result.Failed
		groupResult.Total += result.Total
		groupResult.Duration += result.Duration
	}
}

// RunTestGroups runs a set of test groups using the default test runner.
func (r *TestRunner) RunTestGroups(groups ...*TestGroup) *GroupRunResult {
	group := NewTestGroup("").AddGroups(groups...)
	return r.RunTestGroup(group)
}

// RunTestGroupsT runs a set of test groups within the context of a Go test
// using the default test runner.
func (r *TestRunner) RunTestGroupsT(t *testing.T, groups ...*TestGroup) *GroupRunResult {
	group := NewTestGroup("").AddGroups(groups...)
	return r.RunTestGroupT(t, group)
}

// RunTests runs a set of tests using the default test runner.
func RunTests(tests ...TestCase) *GroupRunResult {
	return NewTestRunner().RunTests(tests...)
}

// RunTestsT runs a set of tests within a Go test context
// using the default test runner.
func RunTestsT(t *testing.T, tests ...TestCase) *GroupRunResult {
	return NewTestRunner().RunTestsT(t, tests...)
}

// RunTestGroup runs a test group using the default test runner.
func RunTestGroup(group *TestGroup) *GroupRunResult {
	return NewTestRunner().RunTestGroup(group)
}

// RunTestGroupT runs a test group within the context of a Go test
// using the default test runner.
func RunTestGroupT(t *testing.T, group *TestGroup) *GroupRunResult {
	return NewTestRunner().RunTestGroupT(t, group)
}

// RunTestGroups runs a set of test groups using the default test runner.
func RunTestGroups(groups ...*TestGroup) *GroupRunResult {
	return NewTestRunner().RunTestGroups(groups...)
}

// RunTestGroupsT runs a set of test groups within the context of a Go test
// using the default test runner.
func RunTestGroupsT(t *testing.T, groups ...*TestGroup) *GroupRunResult {
	return NewTestRunner().RunTestGroupsT(t, groups...)
}
