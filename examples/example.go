package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/jefflinse/go-itest/itest"
)

func main() {
	startExampleServer()

	// A custom HTTP request provides complete control over
	// all aspects of the request when needed.
	customReq, _ := http.NewRequest("GET", "http://localhost:8080/foo", nil)

	// Create a test runner to configure how the tests are run.
	runner := itest.NewEndpointTester("http://localhost:8080").
		WithHTTPClient(http.DefaultClient).
		WithRequestTimeout(time.Second * 5).
		WithContinueOnFailure(true)

	runner.RunTests([]*itest.TestCase{

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
			ExpectStatus(201).
			ExpectHeader("My-Custom-Header", "foobar"),

		itest.POST("/foo").
			Describe("Ensure auth credentials are accepted").
			WithHeaders(http.Header{ // set all headers at once
				"Accept": []string{"application/json"},
				"Auth":   []string{"Bearer 12345"},
			}).
			WithBody(`{"key":"value"}`). // specify body as a string
			ExpectStatus(201),

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

	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello, world!"))
	})

	utRunner := itest.NewHandlerTester(mux).WithContinueOnFailure(true)
	utRunner.RunTests([]*itest.TestCase{

		itest.GET("/foo", "Fetch foo by testing a local handler").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		itest.GET("/bar", "This should be a 404").
			ExpectStatus(404),
	})
}

// Simple static webserver for example purposes.
func startExampleServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Some-Header", "foo")
		w.Header().Add("Some-Header", "bar")

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			return
		} else if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		body := "Hello, world!"
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			body = `{"a_string":"Hello, world!","a_number":42,"another_number":3.14,"a_bool":true}`
		}

		w.Write([]byte(body))
	})

	go http.ListenAndServe("localhost:8080", mux)
}
