package expect

import (
	"errors"
	"fmt"
	"strings"
)

// A FailedPredicateError indicates that a predicate failed.
type FailedPredicateError struct {
	// The underlying error.
	Cause error
	// Stack of JSON field names that lead to the current expectation.
	FieldStack []string
}

func (e *FailedPredicateError) Error() string {
	return fmt.Sprintf("%s: %s", e.FieldString(), e.Cause.Error())
}

func (e *FailedPredicateError) Unwrap() error {
	return e.Cause
}

// PushField pushes a field name onto the field stack.
func (e *FailedPredicateError) PushField(field string) {
	e.FieldStack = append([]string{field}, e.FieldStack...)
}

// FieldString returns a dot-delimited string representation of the field stack.
func (e *FailedPredicateError) FieldString() string {
	return strings.ReplaceAll(strings.Join(e.FieldStack, "."), ".[", "[")
}

func failedPredicate(cause error) *FailedPredicateError {
	return &FailedPredicateError{
		Cause:      cause,
		FieldStack: []string{},
	}
}

func wrongTypeError(expected, actual any) *FailedPredicateError {
	var msg string
	if expected != nil && actual == nil {
		msg = fmt.Sprintf("expected %T, got nothing", expected)
	} else {
		msg = fmt.Sprintf("expected type %T, got %T: %+v", expected, actual, actual)
	}

	return failedPredicate(errors.New(msg))
}

func wrongValueError(expected []any, actual any) *FailedPredicateError {
	var msg string
	if len(expected) > 0 && actual == nil {
		msg = fmt.Sprintf("expected %+v, got nothing", expected)
	} else {
		if len(expected) > 1 {
			msg = fmt.Sprintf("expected one of %+v, got %+v", expected, actual)
		} else {
			msg = fmt.Sprintf("expected %+v, got %+v", expected[0], actual)
		}
	}

	return failedPredicate(errors.New(msg))
}
