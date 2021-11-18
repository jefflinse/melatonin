package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/jefflinse/melatonin/mt"
)

// This example shows how to use melatonin to test an actual service endpoint.
func EndpointExample() {
	startExampleServer()

	// A custom HTTP request can be created to provide complete control over
	// all aspects of a particular test case, if needed.
	customReq, _ := http.NewRequest("GET", "http://localhost:8080/foo", nil)

	myAPI := mt.NewURLContext("http://localhost:8080").WithHTTPClient(http.DefaultClient)
	mt.NewTestRunner().RunTests([]mt.TestCase{

		myAPI.GET("/foo", "Fetch foo and ensure it takes less than one second").
			WithTimeout(1 * time.Second). // specify a timeout for the test case
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		myAPI.GET("/foo", "Fetch foo and ensure the returned JSON contains the right values").
			WithHeader("Accept", "application/json").
			ExpectStatus(201).
			ExpectBody(mt.Object{
				"a_string":       "Hello, world!",
				"a_number":       43,
				"another_number": 3.15,
				"a_bool":         false,
			}),

		myAPI.GET("/bar?first=foo&second=bar", "Fetch bar specifying a query string directly").
			ExpectStatus(404),

		myAPI.GET("/bar", "Fetch bar specifying query parameters all at once").
			WithQueryParams(url.Values{
				"first":  []string{"foo"},
				"second": []string{"bar"},
			}).
			ExpectStatus(404),

		myAPI.GET("/bar", "Fetch bar specifying query parameters individually").
			WithQueryParam("first", "foo").
			WithQueryParam("second", "bar").
			ExpectStatus(404),

		myAPI.POST("/foo", "Create a new foo").
			WithHeader("Accept", "application/json"). // add a single header
			WithBody(map[string]interface{}{          // specify the body using Go types
				"key": "value",
			}).
			ExpectStatus(201).
			ExpectHeader("My-Custom-Header", "foobar"),

		myAPI.POST("/foo", "Ensure auth credentials are accepted").
			WithHeaders(http.Header{ // set all headers at once
				"Accept": []string{"application/json"},
				"Auth":   []string{"Bearer 12345"},
			}).
			WithBody(`{"key":"value"}`). // specify body as a string
			ExpectStatus(201),

		myAPI.DELETE("/foo", "Delete a foo").
			ExpectStatus(204),

		// use any custom *http.Request for a test
		myAPI.DO(customReq, "Fetch foo using a custom HTTP request").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		// load expectations from a golden file
		myAPI.GET("/foo", "Fetch foo and match expectations from a golden file").
			WithHeader("Accept", "application/json").
			ExpectGolden("golden/expect-headers-and-json-body.golden"),
	})
}
