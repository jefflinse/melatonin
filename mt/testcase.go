package mt

import (
	"testing"
)

type TestCase interface {
	Action() string
	Target() string
	Description() string
	Execute(*testing.T) (TestResult, error)
}

type TestResult interface {
	Errors() []error
}
