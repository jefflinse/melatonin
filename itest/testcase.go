package itest

import (
	"errors"
	"fmt"
	"log"
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
	// Name is the name of the test case. If not provided, defaults to
	// METHOD URI (body length), e.g.:
	//
	//   POST /users (210)
	Name string

	// Setup is an optional function that is run before the test is run.
	// It can be used to perform any prerequisites actions for the test,
	// such as adding or removing objects in a database.
	Setup func()

	// Method is the HTTP method to use for the request. Default is "GET".
	Method string

	// URI is the relative URI to use for the request.
	URI string

	// RequestBody is the content to send in the body of the HTTP request.
	RequestBody Stringable

	// WantStatus is the expected HTTP status code of the response. Default is 200.
	WantStatus int

	// WantBody is the expected HTTP response body content.
	WantBody Stringable

	// ContinueOnFailure indicates whether the test should continue to the next
	// test case if the current test fails. Default is false.
	ContinueOnFailure bool
}

// Validate ensures that the test case is valid can can be run.
func (tc *TestCase) Validate() error {
	if tc.Method == "" {
		return errors.New("missing Method")
	} else if tc.URI == "" {
		return errors.New("missing URI")
	} else if tc.URI[0] != '/' {
		return errors.New("URI must begin with '/'")
	} else if tc.WantStatus == 0 {
		return errors.New("missing WantStatus")
	} else if tc.WantBody == nil {
		return errors.New("missing WantBody")
	}

	if tc.Name == "" {
		tc.Name = fmt.Sprintf("%s %s (%d)", tc.Method, tc.URI, len(tc.RequestBody.String()))
	}

	return nil
}

func assertTypeAndValue(key string, expected, actual interface{}) error {
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
			log.Fatalf("field %q did not satisfy predicate, got %q\n", key, actual)
		}

	default:
		log.Fatalf("unexpected value type for field %q: %T\n", key, actual)
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
		if err := assertTypeAndValue(fmt.Sprintf("%s.%s", key, k), v, m[k]); err != nil {
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
		if err := assertTypeAndValue(fmt.Sprintf("%s[%d]", key, i), v, a[i]); err != nil {
			return err
		}
	}

	return nil
}
