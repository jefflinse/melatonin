package expect

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	mtjson "github.com/jefflinse/melatonin/json"
)

// A Predicate is a function that takes a test result value and possibly returns an error.
type Predicate func(interface{}) error

// Then creates a new predicate by chaining the given predicates.
func (p Predicate) Then(next Predicate) Predicate {
	if next == nil {
		return p
	}

	return func(actual interface{}) error {
		if err := p(actual); err != nil {
			return err
		}

		return next(actual)
	}
}

// Bool creates a predicate requiring a value to be a bool, optionally matching
// against a set of values.
func Bool(values ...bool) Predicate {
	return func(actual interface{}) error {
		n, ok := actual.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", actual)
		}

		if len(values) > 0 {
			for _, value := range values {
				if errs := Value(value, n, true); len(errs) == 0 {
					return nil
				}
			}

			return fmt.Errorf("expected one of %v, got %v", values, n)
		}

		return nil
	}
}

// Float64 creates a predicate requiring a value to be a float64, optionally matching
// against a set of values.
func Float64(values ...float64) Predicate {
	return func(actual interface{}) error {
		n, ok := actual.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", actual)
		}

		if len(values) > 0 {
			for _, value := range values {
				if errs := Value(value, n, true); len(errs) == 0 {
					return nil
				}
			}

			return fmt.Errorf("expected one of %v, got %v", values, n)
		}

		return nil
	}
}

// Int64 creates a predicate requiring a value to be an int64, optionally matching
// against a set of values.
func Int64(values ...int64) Predicate {
	return func(actual interface{}) error {
		n, ok := actual.(int64)
		if !ok {
			f, ok := actual.(float64)
			if !ok {
				return fmt.Errorf("expected int64, got %T", actual)
			}

			if n, ok = floatToInt(f); !ok {
				return fmt.Errorf("expected int64, got %T", actual)
			}
		}

		if len(values) > 0 {
			for _, value := range values {
				if n == value {
					return nil
				}
			}

			return fmt.Errorf("expected one of %v, got %d", values, n)
		}

		return nil
	}
}

// Map creates a predicate requiring a value to be a map, optionally matching
// against a set of values.
func Map(values ...map[string]interface{}) Predicate {
	return func(actual interface{}) error {
		m, ok := actual.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map, got %T", actual)
		}

		if len(values) > 0 {
			for _, value := range values {
				if errs := Value(value, m, true); len(errs) == 0 {
					return nil
				}
			}

			return fmt.Errorf("expected one of %v, got %v", values, m)
		}

		return nil
	}
}

// Slice creates a predicate requiring a value to be a slice, optionally matching
// against a set of values.
func Slice(values ...[]interface{}) Predicate {
	return func(actual interface{}) error {
		s, ok := actual.([]interface{})
		if !ok {
			return fmt.Errorf("expected slice, got %T", actual)
		}

		if len(values) > 0 {
			for _, value := range values {
				if errs := Value(value, s, true); len(errs) == 0 {
					return nil
				}
			}

			return fmt.Errorf("expected one of %v, got %v", values, s)
		}

		return nil
	}
}

// String creates a predicate requiring a value to be a string, optionally matching
// against a set of values.
func String(values ...string) Predicate {
	return func(actual interface{}) error {
		s, ok := actual.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", actual)
		}

		if len(values) > 0 {
			for _, value := range values {
				if s == value {
					return nil
				}
			}

			return fmt.Errorf("expected one of %v, got %q", values, s)
		}

		return nil
	}
}

// Headers compares a set of expected headers against a set of actual headers,
func Headers(expected http.Header, actual http.Header) []error {
	var errs []error
	for key, expectedValues := range expected {
		actualValues, ok := actual[key]
		if !ok {
			errs = append(errs, fmt.Errorf("expected header %q, got nothing", key))
			continue
		}

		sort.Strings(expectedValues)
		sort.Strings(actualValues)

		for _, expectedValue := range expectedValues {
			found := false
			for _, actualValue := range actualValues {
				if actualValue == expectedValue {
					found = true
					break
				}
			}

			if !found {
				errs = append(errs, fmt.Errorf("expected header %q to contain %q, got %q", key, expectedValue, actualValues))
			}
		}
	}

	return errs
}

// Status compares an expected status code to an actual status code.
func Status(expected, actual int) error {
	if expected != actual {
		return fmt.Errorf(`expected status %d, got %d`, expected, actual)
	}
	return nil
}

// Value compares an expected value to an actual value.
func Value(expected, actual interface{}, exactJSON bool) []error {
	switch expectedValue := expected.(type) {

	case mtjson.Object, map[string]interface{}:
		ev, ok := expectedValue.(map[string]interface{})
		if !ok {
			ev = map[string]interface{}(expectedValue.(mtjson.Object))
		}
		return mapVal(ev, actual, exactJSON)

	case mtjson.Array, []interface{}:
		ev, ok := expectedValue.([]interface{})
		if !ok {
			ev = []interface{}(expectedValue.(mtjson.Array))
		}
		return arrayVal(ev, actual, exactJSON)

	case string:
		err := strVal(expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case *string:
		return []error{errors.New("bar")}

	case float64:
		err := numVal(expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case *float64:
		err := numVal(*expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case int, int64:
		ev, ok := expectedValue.(int64)
		if !ok {
			ev = int64(expectedValue.(int))
		}

		err := numVal(float64(ev), actual)
		if err != nil {
			return []error{err}
		}

	case *int, *int64:
		return []error{errors.New("foo")}

	case bool:
		err := boolVal(expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case Predicate:
		if err := expectedValue(actual); err != nil {
			return []error{err}
		}

	default:
		return []error{fmt.Errorf("unexpected value type: %T", actual)}
	}

	return nil
}

// boolVal compares an expected bool to an actual bool.
func boolVal(expected bool, actual interface{}) error {
	b, ok := actual.(bool)
	if !ok {
		return wrongTypeError(expected, actual)
	}

	if b != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

// numVal compares an expected float64 to an actual float64.
func numVal(expected float64, actual interface{}) error {
	n, ok := actual.(float64)
	if !ok {
		return wrongTypeError(expected, actual)
	}

	if n != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

// strVal compares an expected string to an actual string.
func strVal(expected string, actual interface{}) error {
	s, ok := actual.(string)
	if !ok {
		return wrongTypeError(expected, actual)
	}

	if s != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

// mapVal compares an expected JSON object to an actual JSON object.
func mapVal(expected map[string]interface{}, actual interface{}, exact bool) []error {
	m, ok := actual.(map[string]interface{})
	if !ok {
		return []error{wrongTypeError(expected, actual)}
	}

	if exact {
		if len(m) != len(expected) {
			return []error{fmt.Errorf("expected %d fields, got %d", len(expected), len(m))}
		}

		expectedKeys := make([]string, 0, len(expected))
		for k := range expected {
			expectedKeys = append(expectedKeys, k)
		}

		actualKeys := make([]string, 0, len(m))
		for k := range m {
			actualKeys = append(actualKeys, k)
		}

		sort.Strings(expectedKeys)
		sort.Strings(actualKeys)

		for i := range expectedKeys {
			if expectedKeys[i] != actualKeys[i] {
				return []error{fmt.Errorf("expected key %q, got %q", expectedKeys[i], actualKeys[i])}
			}
		}
	}

	errs := []error{}
	for k, v := range expected {
		if elemErrs := Value(v, m[k], exact); len(elemErrs) > 0 {
			errs = append(errs, elemErrs...)
		}
	}

	return errs
}

// arrayVal compares an expected JSON array to an actual JSON array.
func arrayVal(expected []interface{}, actual interface{}, exact bool) []error {
	a, ok := actual.([]interface{})
	if !ok {
		return []error{wrongTypeError(expected, actual)}
	}

	if exact && len(a) != len(expected) {
		return []error{fmt.Errorf("expected %d elements, got %d", len(expected), len(a))}
	}

	errs := []error{}
	for i, v := range expected {
		if elemErrs := Value(v, a[i], exact); len(elemErrs) > 0 {
			errs = append(errs, elemErrs...)
		}
	}

	return errs
}

func failedPredicateError(err error) error {
	return err
}

func wrongTypeError(expected, actual interface{}) error {
	var msg string
	if expected != nil && actual == nil {
		msg = fmt.Sprintf("expected %T, got nothing", expected)
	} else {
		msg = fmt.Sprintf("expected type %T, got %T", expected, actual)
	}

	return errors.New(msg)
}

func wrongValueError(expected []interface{}, actual interface{}) error {
	var msg string
	if len(expected) > 0 && actual == nil {
		msg = fmt.Sprintf("expected %v, got nothing", expected)
	} else {
		if len(expected) > 1 {
			msg = fmt.Sprintf("expected one of %v, got %v", expected, actual)
		} else {
			msg = fmt.Sprintf("expected %v, got %v", expected[0], actual)
		}
	}

	return errors.New(msg)
}

func floatToInt(f float64) (int64, bool) {
	n := int64(f)
	return n, float64(n) == f
}
