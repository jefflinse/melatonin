package itest

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// A TestCase tests a single call to an HTTP endpoint.
//
// An optional setup function can be provided to perform any necessary
// setup before the test is run, such as adding or removing objects in
// a database.
//
// All fields in the WantBody map are expected to be present in the
// response body.
type TestCase struct {
	Name        string
	Setup       func(t *testing.T)
	Method      string
	URI         string
	RequestBody Stringable
	WantStatus  int
	WantBody    Stringable
}

// Validate ensures that the test case is valid can can be run.
func (tc *TestCase) validate() error {
	if tc.Name == "" {
		return errors.New("missing Name")
	} else if tc.Method == "" {
		return errors.New("missing Method")
	} else if tc.URI == "" {
		return errors.New("missing URI")
	} else if tc.WantStatus == 0 {
		return errors.New("missing WantStatus")
	} else if tc.WantBody == nil {
		return errors.New("missing WantBody")
	}

	return nil
}

func assertTypeAndValue(t *testing.T, key string, expected, actual interface{}) {
	t.Helper()

	switch expectedValue := expected.(type) {

	case JSONMap, map[string]interface{}:
		expectedMap, ok := expectedValue.(JSONMap)
		if !ok {
			expectedMap = JSONMap(expectedValue.(map[string]interface{}))
		}

		value := requireJSONMap(t, key, actual)
		for wantKey, wantVal := range expectedMap {
			assertTypeAndValue(t, fmt.Sprintf("%s.%s", key, wantKey), wantVal, value[wantKey])
		}

	case JSONArray, []interface{}:
		expectedArray, ok := expectedValue.(JSONArray)
		if !ok {
			expectedArray = JSONArray(expectedValue.([]interface{}))
		}

		value := requireJSONArray(t, key, actual)
		if len(value) != len(expectedArray) {
			t.Fatalf("expected %d values for field %q, got %d\n", len(expectedArray), key, len(value))
		}

		for i, wantItem := range expectedArray {
			assertTypeAndValue(t, fmt.Sprintf("%s[%d]", key, i), wantItem, value[i])
		}

	case string:
		value := requireString(t, key, actual)
		if expectedValue == "ANY_DATETIME" {
			if _, err := time.Parse(time.RFC3339, value); err != nil {
				t.Fatalf("expected valid datetime value for field %q, got %q\n", key, value)
			}
		} else if value != expectedValue {
			failedValueExpectation(t, key, expectedValue, value)
		}

	case int:
		value := requireInt(t, key, actual)
		if value != expectedValue {
			failedValueExpectation(t, key, expectedValue, value)
		}

	case float64:
		value := requireFloat(t, key, actual)
		if value != expectedValue {
			failedValueExpectation(t, key, expectedValue, value)
		}

	case bool:
		value := requireBool(t, key, actual)
		if value != expectedValue {
			failedValueExpectation(t, key, expectedValue, value)
		}

	default:
		t.Fatalf("unexpected value type for field %q: %T\n", key, actual)
	}
}

func failedTypeExpectation(t *testing.T, key, expectedType string, actualValue interface{}) {
	t.Helper()
	t.Fatalf("expected %s for field %q, got %T\n", expectedType, key, actualValue)
}

func failedValueExpectation(t *testing.T, key string, expectedValue, actualValue interface{}) {
	t.Helper()
	t.Fatalf("expected value %q for field %q, got %q\n", expectedValue, key, actualValue)
}

func failOnError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func requireBool(t *testing.T, key string, v interface{}) bool {
	t.Helper()
	b, ok := v.(bool)
	if !ok {
		failedTypeExpectation(t, key, "bool", v)
	}

	return b
}

func requireInt(t *testing.T, key string, v interface{}) int {
	t.Helper()
	i, ok := v.(int)
	if !ok {
		f, ok := v.(float64)
		if !ok || f != float64(int(f)) {
			failedTypeExpectation(t, key, "int", v)
		}

		i = int(f)
	}

	return i
}

func requireFloat(t *testing.T, key string, v interface{}) float64 {
	t.Helper()
	f, ok := v.(float64)
	if !ok {
		failedTypeExpectation(t, key, "float", v)
	}

	return f
}

func requireString(t *testing.T, key string, v interface{}) string {
	t.Helper()
	s, ok := v.(string)
	if !ok {
		failedTypeExpectation(t, key, "string", v)
	}

	return s
}

func requireJSONMap(t *testing.T, key string, v interface{}) JSONMap {
	t.Helper()
	m, ok := v.(JSONMap)
	if !ok {
		failedTypeExpectation(t, key, "map", v)
	}

	return m
}

func requireJSONArray(t *testing.T, key string, v interface{}) JSONArray {
	t.Helper()
	a, ok := v.(JSONArray)
	if !ok {
		failedTypeExpectation(t, key, "array", v)
	}

	return a
}
