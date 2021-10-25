package itest

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
)

func wrongTypeError(key string, expected, actual interface{}) error {
	var msg string
	if expected != nil && actual == nil {
		msg = fmt.Sprintf("expected %T, got nothing", expected)
	} else {
		msg = fmt.Sprintf(`expected type "%T", got '%T"`, expected, actual)
	}

	if key != "" {
		msg = fmt.Sprintf("%s: %s", key, msg)
	}
	return errors.New(msg)
}

func wrongValueError(key string, expected, actual interface{}) error {
	var msg string
	if expected != nil && actual == nil {
		msg = fmt.Sprintf("expected %v, got nothing", expected)
	} else {
		msg = fmt.Sprintf(`expected "%v", got "%v"`, expected, actual)
	}

	if key != "" {
		msg = fmt.Sprintf("%s: %s", key, msg)
	}

	return errors.New(msg)
}

func expectStatus(expected, actual int) error {
	if expected != actual {
		return fmt.Errorf(`expected status %d, got %d`, expected, actual)
	}
	return nil
}

func expect(key string, expected, actual interface{}) error {
	switch expectedValue := expected.(type) {

	case Object, map[string]interface{}:
		ev, _ := expectedValue.(map[string]interface{})
		return expectJSONObject(key, ev, actual)

	case Array, []interface{}:
		ev, _ := expectedValue.([]interface{})
		return expectJSONArray(key, ev, actual)

	case string:
		return expectString(key, expectedValue, actual)

	case float64:
		return expectNumber(key, expectedValue, actual)

	case bool:
		return expectBool(key, expectedValue, actual)

	case func(interface{}) bool:
		if !expectedValue(actual) {
			fatal("field %q did not satisfy predicate, got %q\n", key, actual)
		}

	default:
		fatal("unexpected value type for field %q: %T\n", key, actual)
	}

	return nil
}

func expectBool(key string, expected bool, actual interface{}) error {
	b, ok := actual.(bool)
	if !ok {
		return wrongTypeError(key, expected, actual)
	}

	if b != expected {
		return wrongValueError(key, expected, actual)
	}

	return nil
}

func expectNumber(key string, expected float64, actual interface{}) error {
	n, ok := actual.(float64)
	if !ok {
		return wrongTypeError(key, expected, actual)
	}

	if n != expected {
		return wrongValueError(key, expected, actual)
	}

	return nil
}

func expectString(key string, expected string, actual interface{}) error {
	s, ok := actual.(string)
	if !ok {
		return wrongTypeError(key, expected, actual)
	}

	if s != expected {
		return wrongValueError(key, expected, actual)
	}

	return nil
}

func expectJSONObject(key string, expected map[string]interface{}, actual interface{}) error {
	m, ok := actual.(map[string]interface{})
	if !ok {
		return wrongTypeError(key, expected, actual)
	}

	for k, v := range expected {
		if err := expect(fmt.Sprintf("%s.%s", key, k), v, m[k]); err != nil {
			return err
		}
	}

	return nil
}

func expectJSONArray(key string, expected []interface{}, actual interface{}) error {
	a, ok := actual.([]interface{})
	if !ok {
		return wrongTypeError(key, expected, actual)
	}

	for i, v := range expected {
		if err := expect(fmt.Sprintf("%s[%d]", key, i), v, a[i]); err != nil {
			return err
		}
	}

	return nil
}

func expectHeaders(expected http.Header, actual http.Header) []error {
	var errs []error
	for key, expectedValuesForKey := range expected {
		actualValuesForKey, ok := actual[key]
		if !ok {
			errs = append(errs, fmt.Errorf("expected header %q, got nothing", key))
			continue
		}

		sort.Strings(expectedValuesForKey)
		sort.Strings(actualValuesForKey)

		for _, expectedValue := range expectedValuesForKey {
			found := false
			for _, actualValue := range actualValuesForKey {
				if actualValue == expectedValue {
					found = true
					break
				}
			}

			if !found {
				errs = append(errs, fmt.Errorf("expected header %q to contain %q, got %q", key, expectedValue, actualValuesForKey))
			}
		}
	}

	return errs
}
