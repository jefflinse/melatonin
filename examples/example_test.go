package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/jefflinse/go-itest/itest"
)

func TestAPI(t *testing.T) {
	startExampleServer()

	// A custom HTTP request provides complete control over
	// all aspects of the request when needed.
	customReq, _ := http.NewRequest("GET", "http://localhost:8080/foo", nil)

	// Create a test runner to configure how the tests are run.
	runner := itest.NewTestRunner("http://localhost:8080").
		WithHTTPClient(http.DefaultClient).
		WithRequestTimeout(time.Second * 5).
		WithContinueOnFailure(true)

	runner.RunTestsT(t, []*itest.TestCase{

		itest.GET("/foo").
			WithTimeout(1 * time.Second). // specify a timeout for the test case
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		itest.GET("/foo").
			WithHeader("Accept", "application/json").
			ExpectStatus(200).
			ExpectBody(itest.Object{
				"a_string":       "Hello, world!",
				"a_number":       42,
				"another_number": 3.14,
				"a_bool":         true,
			}),

		itest.GET("/bar?query=foo&other=bar").
			ExpectStatus(404),

		itest.POST("/foo").
			WithHeader("Accept", "application/json"). // add a single header
			WithBody(map[string]interface{}{          // specify the body using Go types
				"key": "value",
			}).
			ExpectStatus(201),

		itest.POST("/foo").
			WithHeaders(http.Header{ // set all headers at once
				"Accept": []string{"application/json"},
				"Auth":   []string{"Bearer 12345"},
			}).
			WithBody(`{"key":"value"}`). // specify body as a string
			ExpectStatus(201),

		itest.DELETE("/foo").
			ExpectStatus(204),

		// use any custom *http.Request for a test
		itest.DO(customReq).
			ExpectStatus(200).
			ExpectBody("Hello, world!"),
	})
}
