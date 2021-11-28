package mt

// A TestGroup is a set of TestCases with associated metadata.
type TestGroup struct {
	Name      string
	TestCases []TestCase
	Parallel  bool
}

// NewTestGroup creates a new TestGroup with the given name.
func NewTestGroup(name string) *TestGroup {
	return &TestGroup{
		Name:      name,
		TestCases: []TestCase{},
		Parallel:  false,
	}
}

// Add adds one or more TestCases to the TestGroup.
func (g *TestGroup) Add(tc ...TestCase) *TestGroup {
	g.TestCases = append(g.TestCases, tc...)
	return g
}

// InParallel indicates that the tests in the TestGroup should be run in parallel.
func (g *TestGroup) InParallel() *TestGroup {
	g.Parallel = true
	return g
}

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
