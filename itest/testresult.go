package itest

type TestCaseResult struct {
	TestCase *TestCase
	Errors   []error
}

func (r *TestCaseResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
}
