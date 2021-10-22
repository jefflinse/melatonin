package itest

import (
	"errors"
	"fmt"
	"net/http"
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
	// METHOD Path (body length), e.g.:
	//
	//   POST /users (210)
	Name string

	// Setup is an optional function that is run before the test is run.
	// It can be used to perform any prerequisites actions for the test,
	// such as adding or removing objects in a database.
	Setup func() error

	// Method is the HTTP method to use for the request. Default is "GET".
	Method string

	// Path is the relative Path to use for the request.
	Path string

	// RequestHeaders is a map of HTTP headers to use for the request.
	RequestHeaders http.Header

	// RequestBody is the content to send in the body of the HTTP request.
	RequestBody Stringable

	// WantStatus is the expected HTTP status code of the response. Default is 200.
	WantStatus int

	// WantHeaders is a map of HTTP headers that are expected to be present in
	// the HTTP response.
	WantHeaders http.Header

	// WantBody is the expected HTTP response body content.
	WantBody Stringable

	// ContinueOnFailure indicates whether the test should continue to the next
	// test case if the current test fails. Default is false.
	ContinueOnFailure bool
}

func GET(path string) *TestCase {
	return &TestCase{Method: "GET", Path: path}
}

func POST(path string) *TestCase {
	return &TestCase{Method: "POST", Path: path}
}

func PUT(path string) *TestCase {
	return &TestCase{Method: "PUT", Path: path}
}

func PATCH(path string) *TestCase {
	return &TestCase{Method: "PATCH", Path: path}
}

func DELETE(path string) *TestCase {
	return &TestCase{Method: "DELETE", Path: path}
}

func (tc *TestCase) WithSetup(setup func() error) *TestCase {
	if tc.Setup != nil {
		fatal("test case %q specifies more than one setup function", underlineText(tc.Name))
	}

	tc.Setup = setup
	return tc
}

func (tc *TestCase) WithHeaders(headers http.Header) *TestCase {
	if tc.RequestHeaders != nil {
		fatal("test case %q specifies request headers more than once", tc.Name)
	}

	tc.RequestHeaders = headers
	return tc
}

func (tc *TestCase) WithHeader(key, value string) *TestCase {
	if tc.RequestHeaders == nil {
		tc.RequestHeaders = http.Header{}
	}

	tc.RequestHeaders.Set(key, value)
	return tc
}

func (tc *TestCase) WithBody(body Stringable) *TestCase {
	if tc.RequestBody != nil {
		fatal("test case %q specifies more than one request body", tc.Name)
	}

	tc.RequestBody = body
	return tc
}

func (tc *TestCase) ExpectStatus(status int) *TestCase {
	if tc.WantStatus > 0 {
		fatal("test case %q specifies more than one expected status", tc.Name)
	}

	tc.WantStatus = status
	return tc
}

func (tc *TestCase) ExpectHeaders(headers http.Header) *TestCase {
	if tc.WantHeaders != nil && len(tc.WantHeaders) > 0 {
		fatal("test case %q overrides previously defined expected headers", tc.Name)
	}

	tc.WantHeaders = headers
	return tc
}

func (tc *TestCase) ExpectHeader(key, value string) *TestCase {
	if tc.WantHeaders == nil {
		tc.WantHeaders = http.Header{}
	}

	tc.WantHeaders.Set(key, value)
	return tc
}

func (tc *TestCase) ExpectBody(body Stringable) *TestCase {
	if tc.WantBody != nil {
		fatal("test case %q specifies more than one expected body", tc.Name)
	}

	tc.WantBody = body
	return tc
}

// Validate ensures that the test case is valid can can be run.
func (tc *TestCase) Validate() error {
	if tc.Method == "" {
		return errors.New("missing Method")
	} else if tc.Path == "" {
		return errors.New("missing Path")
	} else if tc.Path[0] != '/' {
		return errors.New("path must begin with '/'")
	} else if tc.WantStatus == 0 {
		return errors.New("missing WantStatus")
	} else if tc.WantBody == nil {
		return errors.New("missing WantBody")
	}

	if tc.Name == "" {
		tc.Name = fmt.Sprintf("%s %s", tc.Method, tc.Path)
		if tc.RequestBody != nil {
			tc.Name += fmt.Sprintf(" (%d)", len(tc.RequestBody.String()))
		}
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
