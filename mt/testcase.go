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

// A TestResult is anything that references a TestCase and produces a set
// of failures. The success of any test result is determined by the number of
// failures in the result.
type TestResult interface {
	TestCase() TestCase
	Failures() []error
}
