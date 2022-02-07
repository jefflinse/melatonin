package expect

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	mtjson "github.com/jefflinse/melatonin/json"
)

// A Predicate is a function that takes a test result value and possibly returns an error.
type Predicate func(interface{}) error

// Then chains a new Predicate to run after the current Predicate.
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

// And chains a new Predicate to run after the current Predicate if the current Predicate succeeds.
func (p Predicate) And(next Predicate) Predicate {
	return p.Then(next)
}

// Or chains a new Predicate to run after the current Predicate if the current Predicate fails.
func (p Predicate) Or(next Predicate) Predicate {
	if next == nil {
		return p
	}

	return func(actual interface{}) error {
		if err := p(actual); err != nil {
			return next(actual)
		}

		return nil
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

// Float creates a predicate requiring a value to be an floating point number,
// optionally matching against a set of values.
func Float(expected ...float64) Predicate {
	return func(actual interface{}) error {
		n, ok := toFloat(actual)
		if !ok {
			return wrongTypeError(float64(0), actual)
		}

		if len(expected) > 0 {
			for _, value := range expected {
				if n == value {
					return nil
				}
			}

			return fmt.Errorf("expected one of %+v, got %g", expected, n)
		}

		return nil
	}
}

// Int creates a predicate requiring a value to be an integer, optionally matching
// against a set of values.
func Int(expected ...int64) Predicate {
	return func(actual interface{}) error {
		n, ok := toInt(actual)
		if !ok {
			return wrongTypeError(expected, actual)
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
			return fmt.Errorf("expected map, got %T: %+v", actual, actual)
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

// Pattern creates a predicate requiring a value to be a string that matches a
// regular expression, optionally matching against a set of values.
func Pattern(regex string) Predicate {
	r, err := regexp.Compile(regex)
	if err != nil {
		return func(interface{}) error {
			return fmt.Errorf("invalid regex: %q", regex)
		}
	}

	return Regex(r)
}

// Regex creates a predicate requiring a value to be a string that matches a
// regular expression, optionally matching against a set of values.
func Regex(regex *regexp.Regexp) Predicate {
	return String().Then(func(actual interface{}) error {
		s, _ := actual.(string)
		if !regex.MatchString(s) {
			return fmt.Errorf("expected to match pattern %q, got %q", regex.String(), s)
		}

		return nil
	})
}

// Slice creates a predicate requiring a value to be a slice, optionally matching
// against a set of values.
func Slice(expected ...[]interface{}) Predicate {
	return func(actual interface{}) error {
		s, ok := actual.([]interface{})
		if !ok {
			return fmt.Errorf("expected slice, got %T: %+v", actual, actual)
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
			return fmt.Errorf("expected string, got %T: %+v", actual, actual)
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
func CompareValues(expected, actual interface{}, exactJSON bool) []*FailedPredicateError {
	errs := []*FailedPredicateError{}

	if expected == nil && actual != nil {
		errs = append(errs, failedPredicate(fmt.Errorf("expected nil, got %T: %+v", actual, actual)))
	}

	switch expectedValue := expected.(type) {

	case bool:
		if err := compareBoolValues(expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
		}

	case *bool:
		if err := compareBoolValues(*expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
		}

	case float64:
		if err := compareFloat64Values(expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
		}

	case *float64:
		if err := compareFloat64Values(*expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
		}

	case int64:
		if err := compareInt64Values(expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
		}

	case *int64:
		if err := compareInt64Values(*expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
		}

	case string:
		if err := compareStringValues(expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
		}

	case *string:
		if err := compareStringValues(*expectedValue, actual); err != nil {
			errs = append(errs, err)
			return errs
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
			errs = append(errs, failedPredicate(err))
			return errs
		}

	default:
		errs = append(errs, failedPredicate(fmt.Errorf("unexpected value type: %T", actual)))
	}

	return nil
}

// compareBoolValues compares an expected bool to an actual bool.
func compareBoolValues(expected bool, actual interface{}) *FailedPredicateError {
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
func compareFloat64Values(expected float64, actual interface{}) *FailedPredicateError {
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
func compareInt64Values(expected int64, actual interface{}) *FailedPredicateError {
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
func compareMapValues(expected map[string]interface{}, actual interface{}, exact bool) []*FailedPredicateError {
	errs := []*FailedPredicateError{}

	m, ok := actual.(map[string]interface{})
	if !ok {
		errs = append(errs, wrongTypeError(expected, actual))
		return errs
	}

	if exact {
		if len(m) != len(expected) {
			j, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				errs = append(errs, failedPredicate(err))
			}

			errs = append(errs, failedPredicate(fmt.Errorf("expected %d fields, got %d:\n%+v", len(expected), len(m), string(j))))
			return errs
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
				errs = append(errs, failedPredicate(fmt.Errorf("expected key %q, got %q: %+v", expectedKeys[i], actualKeys[i], m[actualKeys[i]])))
			}
		}
	}

	for k, v := range expected {
		for _, err := range CompareValues(v, m[k], exact) {
			err.PushField(k)
			errs = append(errs, err)
		}
	}

	return errs
}

// compareSliceValues compares an expected slice to an actual slice.
func compareSliceValues(expected []interface{}, actual interface{}, exact bool) []*FailedPredicateError {
	errs := []*FailedPredicateError{}

	a, ok := actual.([]interface{})
	if !ok {
		errs = append(errs, wrongTypeError(expected, actual))
		return errs
	}

	if len(a) < len(expected) {
		j, err := json.MarshalIndent(a, "", "  ")
		if err != nil {
			errs = append(errs, failedPredicate(err))
		}
		errs = append(errs, failedPredicate(fmt.Errorf("expected at least %d elements, got %d: %+v", len(expected), len(a), string(j))))
		return errs
	} else if exact && len(a) > len(expected) {
		j, err := json.MarshalIndent(a, "", "  ")
		if err != nil {
			errs = append(errs, failedPredicate(err))
		}
		errs = append(errs, failedPredicate(fmt.Errorf("expected %d elements, got %d: %+v", len(expected), len(a), string(j))))
		return errs
	}

	for i, v := range expected {
		for _, err := range CompareValues(v, a[i], exact) {
			err.PushField(fmt.Sprintf("[%d]", i))
			errs = append(errs, err)
		}
	}

	return errs
}

// compareStringValues compares an expected string to an actual string.
func compareStringValues(expected string, actual interface{}) *FailedPredicateError {
	s, ok := actual.(string)
	if !ok {
		return wrongTypeError(expected, actual)
	}

	if s != expected {
		return wrongValueError([]interface{}{expected}, actual)
	}

	return nil
}

func floatToInt(f float64) (int64, bool) {
	n := int64(f)
	return n, float64(n) == f
}

func toInt(v interface{}) (int64, bool) {
	switch v := v.(type) {
	case int64:
		return v, true
	case float64:
		return floatToInt(v)
	case float32:
		return floatToInt(float64(v))
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	default:
		return 0, false
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch v := v.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	}

	if i, ok := toInt(v); ok {
		return float64(i), true
	}

	return 0, false
}
