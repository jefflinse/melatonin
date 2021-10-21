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

	case JSONMap:
		actualMap, ok := actual.(JSONMap)
		if !ok {
			failedTypeExpectation(t, key, "map", actual)
		}

		for mapKey, mapVal := range expectedValue {
			actualMapVal := actualMap[mapKey]
			assertTypeAndValue(t, fmt.Sprintf("%s.%s", key, mapKey), mapVal, actualMapVal)
		}

	case JSONArray:
		actualSlice, ok := actual.(JSONArray)
		if !ok {
			failedTypeExpectation(t, key, "array", actual)
		}

		if len(actualSlice) != len(expectedValue) {
			t.Fatalf("expected %d values for field %q, got %d\n", len(expectedValue), key, len(actualSlice))
		}

		for i, arrayItem := range expectedValue {
			actualItem := actualSlice[i]
			assertTypeAndValue(t, fmt.Sprintf("%s[%d]", key, i), arrayItem, actualItem)
		}

	case string:
		actualStr, ok := actual.(string)
		if !ok {
			failedTypeExpectation(t, key, "string", actual)
		}

		if expectedValue == "ANY_DATETIME" {
			if _, err := time.Parse(time.RFC3339, actualStr); err != nil {
				t.Fatalf("expected valid datetime value for field %q, got %q\n", key, actualStr)
			}
		}

		if actualStr != expectedValue {
			failedValueExpectation(t, key, expectedValue, actualStr)
		}

	case int:
		actualInt, ok := actual.(int)
		if !ok {
			actualFloat, ok := actual.(float64)
			if !ok {
				failedTypeExpectation(t, key, "int", actual)
			}

			actualInt = int(actualFloat)
			if actualInt != expectedValue {
				failedValueExpectation(t, key, expectedValue, actualInt)
			}
		}

		if actualInt != expectedValue {
			failedValueExpectation(t, key, expectedValue, actualInt)
		}

	case float64:
		actualFloat, ok := actual.(float64)
		if !ok {
			failedTypeExpectation(t, key, "float", actual)
		}

		if actualFloat != expectedValue {
			failedValueExpectation(t, key, expectedValue, actualFloat)
		}

	case bool:
		actualBool, ok := actual.(bool)
		if !ok {
			failedTypeExpectation(t, key, "bool", actual)
		}

		if actualBool != expectedValue {
			failedValueExpectation(t, key, expectedValue, actualBool)
		}

	default:
		t.Fatalf("unexpected value type for field %q: %T\n", key, actual)
	}
}

func failedTypeExpectation(t *testing.T, key, expectedType string, actualValue interface{}) {
	t.Helper()
	t.Fatalf("expected %T for field %q, got %T\n", expectedType, key, actualValue)
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
