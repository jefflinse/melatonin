package mt

type TestCase interface {
	Action() string
	Target() string
	Description() string
	Execute() TestResult
}

type TestResult interface {
	TestCase() TestCase
	Errors() []error
}
