package mt

// A TestGroup is a set of TestCases with associated metadata.
type TestGroup struct {
	Name      string
	TestCases []TestCase
}

// NewTestGroup creates a new TestGroup with the given name.
func NewTestGroup(name string) *TestGroup {
	return &TestGroup{
		Name:      name,
		TestCases: []TestCase{},
	}
}

// Add adds one or more TestCases to the TestGroup.
func (g *TestGroup) Add(tc ...TestCase) *TestGroup {
	g.TestCases = append(g.TestCases, tc...)
	return g
}
