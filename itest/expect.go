package itest

import "fmt"

func WrongTypeError(key string, expected, actual interface{}) error {
	return fmt.Errorf(`%s: expected "%T", got '%T"`, key, expected, actual)
}

func WrongValueError(key string, expected, actual interface{}) error {
	return fmt.Errorf(`%s: expected "%v", got "%v"`, key, expected, actual)
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
		return requireJSONMapValue(key, expectedMap, actual)

	case JSONArray, []interface{}:
		expectedArray, ok := expectedValue.(JSONArray)
		if !ok {
			expectedArray = JSONArray(expectedValue.([]interface{}))
		}
		return requireJSONArray(key, expectedArray, actual)

	case string:
		return requireStringValue(key, expectedValue, actual)

	case int:
		return requireIntValue(key, expectedValue, actual)

	case float64:
		return requireFloatValue(key, expectedValue, actual)

	case bool:
		return requireBoolValue(key, expectedValue, actual)

	case func(interface{}) bool:
		if !expectedValue(actual) {
			fatal("field %q did not satisfy predicate, got %q\n", key, actual)
		}

	default:
		fatal("unexpected value type for field %q: %T\n", key, actual)
	}

	return nil
}

func requireBoolValue(key string, expected bool, actual interface{}) error {
	b, ok := actual.(bool)
	if !ok {
		return WrongTypeError(key, expected, actual)
	} else if b != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func requireIntValue(key string, expected int, actual interface{}) error {
	i, ok := actual.(int)
	if !ok {
		f, ok := actual.(float64)
		if !ok || f != float64(int(f)) {
			return WrongTypeError(key, expected, actual)
		}

		i = int(f)
	}

	if i != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func requireFloatValue(key string, expected float64, actual interface{}) error {
	b, ok := actual.(float64)
	if !ok {
		return WrongTypeError(key, expected, actual)
	} else if b != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func requireStringValue(key string, expected string, actual interface{}) error {
	s, ok := actual.(string)
	if !ok {
		return WrongTypeError(key, expected, actual)
	} else if s != expected {
		return WrongValueError(key, expected, actual)
	}

	return nil
}

func requireJSONMapValue(key string, expected JSONMap, actual interface{}) error {
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

func requireJSONArray(key string, expected JSONArray, actual interface{}) error {
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
