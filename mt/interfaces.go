package mt

// A TestCase is anything that can be Execute()'d to produce a TestResult.
// Additionally, it must provide an Action, Target, and Description for
// reporting purposes.
type TestCase interface {
	Action() string
	Target() string
	Description() string
	Execute() TestResult
}

// A TestResult is anything that produces a set of failures.
// Additionally, it must reference the TestCase that produced it.
type TestResult interface {
	TestCase() TestCase
	Failures() []error
}
