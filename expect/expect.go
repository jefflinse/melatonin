package expect

import (
	"errors"
	"fmt"
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
func Bool(expected ...bool) Predicate {
	return func(actual interface{}) error {
		n, ok := actual.(bool)
		if !ok {
			return wrongTypeError(true, actual)
		}

		if len(expected) > 0 {
			for _, e := range expected {
				if err := compareBoolValues(e, n); err == nil {
					return nil
				}
			}

			return fmt.Errorf("expected one of %+v, got %t", expected, n)
		}

		return nil
	}
}

// Float64 creates a predicate requiring a value to be a float64, optionally matching
// against a set of values.
func Float64(expected ...float64) Predicate {
	return func(actual interface{}) error {
		n, ok := actual.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", actual)
		}

		if len(expected) > 0 {
			for _, value := range expected {
				if errs := CompareValues(value, n, true); len(errs) == 0 {
					return nil
				}
			}

			return fmt.Errorf("expected one of %+v, got %f", expected, n)
		}

		return nil
	}
}

// Int64 creates a predicate requiring a value to be an int64, optionally matching
// against a set of values.
func Int64(expected ...int64) Predicate {
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

		if len(expected) > 0 {
			for _, value := range expected {
				if n == value {
					return nil
				}
			}

			return fmt.Errorf("expected one of %+v, got %d", expected, n)
		}

		return nil
	}
}

// Map creates a predicate requiring a value to be a map, optionally matching
// against a set of values.
func Map(expected ...map[string]interface{}) Predicate {
	return func(actual interface{}) error {
		m, ok := actual.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map, got %T", actual)
		}

		if len(expected) > 0 {
			for _, value := range expected {
				if errs := CompareValues(value, m, true); len(errs) == 0 {
					return nil
				}
			}

			return fmt.Errorf("expected one of %+v, got %+v", expected, m)
		}

		return nil
	}
}

// Slice creates a predicate requiring a value to be a slice, optionally matching
// against a set of values.
func Slice(expected ...[]interface{}) Predicate {
	return func(actual interface{}) error {
		s, ok := actual.([]interface{})
		if !ok {
			return fmt.Errorf("expected slice, got %T", actual)
		}

		if len(expected) > 0 {
			for _, value := range expected {
				if errs := CompareValues(value, s, true); len(errs) == 0 {
					return nil
				}
			}

			return fmt.Errorf("expected one of %+v, got %+v", expected, s)
		}

		return nil
	}
}

// String creates a predicate requiring a value to be a string, optionally matching
// against a set of values.
func String(expected ...string) Predicate {
	return func(actual interface{}) error {
		s, ok := actual.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", actual)
		}

		if len(expected) > 0 {
			for _, value := range expected {
				if s == value {
					return nil
				}
			}

			return fmt.Errorf("expected one of %+v, got %q", expected, s)
		}

		return nil
	}
}

// CompareValues compares an expected value to an actual value.
func CompareValues(expected, actual interface{}, exactJSON bool) []error {
	switch expectedValue := expected.(type) {

	case bool:
		err := compareBoolValues(expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case *bool:
		err := compareBoolValues(*expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case float64:
		err := compareFloat64Values(expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case *float64:
		err := compareFloat64Values(*expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case int64:
		err := compareInt64Values(expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case *int64:
		err := compareInt64Values(*expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case string:
		err := compareStringValues(expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case *string:
		err := compareStringValues(*expectedValue, actual)
		if err != nil {
			return []error{err}
		}

	case mtjson.Object, map[string]interface{}:
		ev, ok := expectedValue.(map[string]interface{})
		if !ok {
			ev = map[string]interface{}(expectedValue.(mtjson.Object))
		}
		return compareMapValues(ev, actual, exactJSON)

	case mtjson.Array, []interface{}:
		ev, ok := expectedValue.([]interface{})
		if !ok {
			ev = []interface{}(expectedValue.(mtjson.Array))
		}
		return compareSliceValues(ev, actual, exactJSON)

	case Predicate, func(interface{}) error:
		f, ok := expectedValue.(Predicate)
		if !ok {
			f = Predicate(expectedValue.(func(interface{}) error))
		}
		if err := f(actual); err != nil {
			return []error{err}
		}

	default:
		return []error{fmt.Errorf("unexpected value type: %T", actual)}
	}

	return nil
}

// compareBoolValues compares an expected bool to an actual bool.
func compareBoolValues(expected bool, actual interface{}) error {
	b, ok := actual.(bool)
	if !ok {
		return wrongTypeError(expected, actual)
	}

	if b != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

// compareFloat64Values compares an expected float64 to an actual float64.
func compareFloat64Values(expected float64, actual interface{}) error {
	n, ok := actual.(float64)
	if !ok {
		return wrongTypeError(expected, actual)
	}

	if n != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

// compareInt64Values compares an expected int64 to an actual int64.
func compareInt64Values(expected int64, actual interface{}) error {
	n, ok := actual.(int64)
	if !ok {
		f, ok := actual.(float64)
		if !ok {
			return wrongTypeError(expected, actual)
		}

		n, ok = floatToInt(f)
		if !ok {
			return wrongTypeError(expected, actual)
		}
	}

	if n != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

// compareMapValues compares an expected JSON object to an actual JSON object.
func compareMapValues(expected map[string]interface{}, actual interface{}, exact bool) []error {
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
		if elemErrs := CompareValues(v, m[k], exact); len(elemErrs) > 0 {
			errs = append(errs, elemErrs...)
		}
	}

	return errs
}

// compareSliceValues compares an expected slice to an actual slice.
func compareSliceValues(expected []interface{}, actual interface{}, exact bool) []error {
	a, ok := actual.([]interface{})
	if !ok {
		return []error{wrongTypeError(expected, actual)}
	}

	if exact && len(a) != len(expected) {
		return []error{fmt.Errorf("expected %d elements, got %d", len(expected), len(a))}
	}

	errs := []error{}
	for i, v := range expected {
		if elemErrs := CompareValues(v, a[i], exact); len(elemErrs) > 0 {
			errs = append(errs, elemErrs...)
		}
	}

	return errs
}

// compareStringValues compares an expected string to an actual string.
func compareStringValues(expected string, actual interface{}) error {
	s, ok := actual.(string)
	if !ok {
		return wrongTypeError(expected, actual)
	}

	if s != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

func wrongTypeError(expected, actual interface{}) error {
	var msg string
	if expected != nil && actual == nil {
		msg = fmt.Sprintf("expected %T, got nothing", expected)
	} else {
		msg = fmt.Sprintf("expected type %T, got %T: %+v", expected, actual, actual)
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
