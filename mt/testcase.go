package mt

import (
	"testing"
)

type TestCase interface {
	Action() string
	Description() string
	Execute(*testing.T) (TestResult, error)
}

type TestResult interface {
	Errors() []error
}
