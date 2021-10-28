package main

import (
	"net/http"
	"net/url"
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
			Describe("Fetch foo and ensure it takes less than one second").
			WithTimeout(1 * time.Second). // specify a timeout for the test case
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		itest.GET("/foo", "The test description can be specified here instead").
			ExpectBody("Hello, world!"),

		itest.GET("/foo").
			Describe("Fetch foo and ensure the returned JSON contains the right values").
			WithHeader("Accept", "application/json").
			ExpectStatus(201).
			ExpectBody(itest.Object{
				"a_string":       "Hello, world!",
				"a_number":       43,
				"another_number": 3.15,
				"a_bool":         false,
			}),

		itest.GET("/bar?first=foo&second=bar").
			Describe("Fetch bar specifying a query string directly").
			ExpectStatus(404),

		itest.GET("/bar").
			Describe("Fetch bar specifying query parameters all at once").
			WithQueryParams(url.Values{
				"first":  []string{"foo"},
				"second": []string{"bar"},
			}).
			ExpectStatus(404),

		itest.GET("/bar").
			Describe("Fetch bar specifying query parameters individually").
			WithQueryParam("first", "foo").
			WithQueryParam("second", "bar").
			ExpectStatus(404),

		itest.POST("/foo").
			Describe("Create a new foo").
			WithHeader("Accept", "application/json"). // add a single header
			WithBody(map[string]interface{}{          // specify the body using Go types
				"key": "value",
			}).
			ExpectStatus(201),

		itest.POST("/foo").
			Describe("Ensure auth credentials are accepted").
			WithHeaders(http.Header{ // set all headers at once
				"Accept": []string{"application/json"},
				"Auth":   []string{"Bearer 12345"},
			}).
			WithBody(`{"key":"value"}`). // specify body as a string
			ExpectStatus(201).
			ExpectHeader("My-Custom-Header", "foobar"),

		itest.DELETE("/foo").
			Describe("Delete a foo").
			ExpectStatus(204),

		// use any custom *http.Request for a test
		itest.DO(customReq).
			Describe("Fetch foo using a custom HTTP request").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		// load expectations from a golden file
		itest.GET("/foo").
			Describe("Fetch foo and match expectations from a golden file").
			WithHeader("Accept", "application/json").
			ExpectGolden("golden/expect-headers-and-json-body.golden"),
	})
}
