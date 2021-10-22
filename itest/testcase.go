package itest

import (
	"errors"
	"fmt"
	"log"
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
	Name              string
	Setup             func()
	Method            string
	URI               string
	RequestBody       Stringable
	WantStatus        int
	WantBody          Stringable
	ContinueOnFailure bool
}

// Validate ensures that the test case is valid can can be run.
func (tc *TestCase) validate(index int) error {
	if tc.Name == "" {
		tc.Name = fmt.Sprint("Test", index)
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

func assertTypeAndValue(key string, expected, actual interface{}) {
	switch expectedValue := expected.(type) {

	case JSONMap, map[string]interface{}:
		expectedMap, ok := expectedValue.(JSONMap)
		if !ok {
			expectedMap = JSONMap(expectedValue.(map[string]interface{}))
		}

		value := requireJSONMap(key, actual)
		for wantKey, wantVal := range expectedMap {
			assertTypeAndValue(fmt.Sprintf("%s.%s", key, wantKey), wantVal, value[wantKey])
		}

	case JSONArray, []interface{}:
		expectedArray, ok := expectedValue.(JSONArray)
		if !ok {
			expectedArray = JSONArray(expectedValue.([]interface{}))
		}

		value := requireJSONArray(key, actual)
		if len(value) != len(expectedArray) {
			log.Fatalf("expected %d values for field %q, got %d\n", len(expectedArray), key, len(value))
		}

		for i, wantItem := range expectedArray {
			assertTypeAndValue(fmt.Sprintf("%s[%d]", key, i), wantItem, value[i])
		}

	case string:
		value := requireString(key, actual)
		if expectedValue == "ANY_DATETIME" {
			if _, err := time.Parse(time.RFC3339, value); err != nil {
				log.Fatalf("expected valid datetime value for field %q, got %q\n", key, value)
			}
		} else if value != expectedValue {
			failedValueExpectation(key, expectedValue, value)
		}

	case int:
		value := requireInt(key, actual)
		if value != expectedValue {
			failedValueExpectation(key, expectedValue, value)
		}

	case float64:
		value := requireFloat(key, actual)
		if value != expectedValue {
			failedValueExpectation(key, expectedValue, value)
		}

	case bool:
		value := requireBool(key, actual)
		if value != expectedValue {
			failedValueExpectation(key, expectedValue, value)
		}

	case func(interface{}) bool:
		if !expectedValue(actual) {
			log.Fatalf("field %q did not satisfy predicate, got %q\n", key, actual)
		}

	default:
		log.Fatalf("unexpected value type for field %q: %T\n", key, actual)
	}
}

func failedTypeExpectation(key, expectedType string, actualValue interface{}) {
	log.Fatalf("expected %s for field %q, got %T\n", expectedType, key, actualValue)
}

func failedValueExpectation(key string, expectedValue, actualValue interface{}) {
	log.Fatalf("expected value %q for field %q, got %q\n", expectedValue, key, actualValue)
}

func failOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func requireBool(key string, v interface{}) bool {
	b, ok := v.(bool)
	if !ok {
		failedTypeExpectation(key, "bool", v)
	}

	return b
}

func requireInt(key string, v interface{}) int {
	i, ok := v.(int)
	if !ok {
		f, ok := v.(float64)
		if !ok || f != float64(int(f)) {
			failedTypeExpectation(key, "int", v)
		}

		i = int(f)
	}

	return i
}

func requireFloat(key string, v interface{}) float64 {
	f, ok := v.(float64)
	if !ok {
		failedTypeExpectation(key, "float", v)
	}

	return f
}

func requireString(key string, v interface{}) string {
	s, ok := v.(string)
	if !ok {
		failedTypeExpectation(key, "string", v)
	}

	return s
}

func requireJSONMap(key string, v interface{}) JSONMap {
	m, ok := v.(JSONMap)
	if !ok {
		m, ok = v.(map[string]interface{})
		if !ok {
			failedTypeExpectation(key, "map", v)
		}

		m = JSONMap(m)
	}

	return m
}

func requireJSONArray(key string, v interface{}) JSONArray {
	a, ok := v.(JSONArray)
	if !ok {
		a, ok = v.([]interface{})
		if !ok {
			failedTypeExpectation(key, "array", v)
		}

		a = JSONArray(a)
	}

	return a
}
