package mt

// A TestGroup is a set of Tests with associated metadata.
//
// Test groups are nestable, and can be used to create a hierarchy
// of tests.
type TestGroup struct {
	Name   string
	Groups []TestGroup
	Tests  []TestCase
}

// NewTestGroup creates a new TestGroup with the given name.
func NewTestGroup(name string) *TestGroup {
	return &TestGroup{
		Name:   name,
		Groups: []TestGroup{},
		Tests:  []TestCase{},
	}
}

// AddGroups adds one or more TestGroups to the TestGroup.
func (g *TestGroup) AddGroups(groups ...TestGroup) *TestGroup {
	g.Groups = append(g.Groups, groups...)
	return g
}

// AddTests adds one or more Tests to the TestGroup.
func (g *TestGroup) AddTests(tc ...TestCase) *TestGroup {
	g.Tests = append(g.Tests, tc...)
	return g
}
