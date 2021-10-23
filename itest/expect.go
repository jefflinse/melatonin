package itest

import (
	"errors"
	"fmt"
)

func WrongTypeError(key string, expected, actual interface{}) error {
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

func WrongValueError(key string, expected, actual interface{}) error {
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

	case JSONMap, map[string]interface{}:
		expectedMap, ok := expectedValue.(JSONMap)
		if !ok {
			expectedMap = JSONMap(expectedValue.(map[string]interface{}))
		}
		return expectJSONMap(key, expectedMap, actual)

	case JSONArray, []interface{}:
		expectedArray, ok := expectedValue.(JSONArray)
		if !ok {
			expectedArray = JSONArray(expectedValue.([]interface{}))
		}
		return expectJSONArray(key, expectedArray, actual)

	case String, string:
		expectedString, ok := expectedValue.(String)
		if !ok {
			expectedString = String(expectedValue.(string))
		}
		return expectString(key, expectedString, actual)

	case Int, int:
		expectedInt, ok := expectedValue.(Int)
		if !ok {
			expectedInt = Int(expectedValue.(int))
		}
		return expectInt(key, expectedInt, actual)

	case Float, float64:
		expectedFloat, ok := expectedValue.(Float)
		if !ok {
			expectedFloat = Float(expectedValue.(float64))
		}
		return expectFloat(key, expectedFloat, actual)

	case Bool, bool:
		expectedBool, ok := expectedValue.(Bool)
		if !ok {
			expectedBool = Bool(expectedValue.(bool))
		}
		return expectBool(key, expectedBool, actual)

	case func(interface{}) bool:
		if !expectedValue(actual) {
			fatal("field %q did not satisfy predicate, got %q\n", key, actual)
		}

	default:
		fatal("unexpected value type for field %q: %T\n", key, actual)
	}

	return nil
}

func expectBool(key string, expected Bool, actual interface{}) error {
	b, ok := actual.(Bool)
	if !ok {
		nb, ok := actual.(bool)
		if !ok {
			return WrongTypeError(key, expected, actual)
		}

		b = Bool(nb)
	}

	if b != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func expectInt(key string, expected Int, actual interface{}) error {
	i, ok := actual.(Int)
	if !ok {
		ni, ok := actual.(int)
		if !ok {
			nf := actual.(float64)
			if !ok || nf != float64(int(nf)) {
				return WrongTypeError(key, expected, actual)
			}

			ni = int(nf)
		}

		i = Int(ni)
	}

	if i != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func expectFloat(key string, expected Float, actual interface{}) error {
	f, ok := actual.(Float)
	if !ok {
		nf, ok := actual.(float64)
		if !ok {
			return WrongTypeError(key, expected, actual)
		}

		f = Float(nf)
	}

	if f != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func expectString(key string, expected String, actual interface{}) error {
	s, ok := actual.(String)
	if !ok {
		ns, ok := actual.(string)
		if !ok {
			return WrongTypeError(key, expected, actual)
		}

		s = String(ns)
	}

	if s != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func expectJSONMap(key string, expected JSONMap, actual interface{}) error {
	m, ok := actual.(JSONMap)
	if !ok {
		m, ok = actual.(map[string]interface{})
		if !ok {
			return WrongTypeError(key, expected, actual)
		}

		m = JSONMap(m)
	}

	for k, v := range expected {
		if err := expect(fmt.Sprintf("%s.%s", key, k), v, m[k]); err != nil {
			return err
		}
	}

	return nil
}

func expectJSONArray(key string, expected JSONArray, actual interface{}) error {
	a, ok := actual.(JSONArray)
	if !ok {
		a, ok = actual.([]interface{})
		if !ok {
			return WrongTypeError(key, expected, actual)
		}

		a = JSONArray(a)
	}

	for i, v := range expected {
		if err := expect(fmt.Sprintf("%s[%d]", key, i), v, a[i]); err != nil {
			return err
		}
	}

	return nil
}
