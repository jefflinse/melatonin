package main

import (
	"net/http"
	"time"

	"github.com/jefflinse/go-itest/itest"
)

func main() {
	startExampleServer()

	// A custom HTTP request provides complete control over
	// all aspects of the request when needed.
	customReq, _ := http.NewRequest("GET", "http://localhost:8080/foo", nil)

	// Create a test runner to configure how the tests are run.
	runner := itest.NewTestRunner("http://localhost:8080").
		WithHTTPClient(http.DefaultClient).
		WithRequestTimeout(time.Second * 5).
		WithContinueOnFailure(true)

	runner.RunTests([]*itest.TestCase{

		itest.GET("/foo").
			WithTimeout(1 * time.Second). // specify a timeout for the test case
			ExpectStatus(200).
			ExpectBody(itest.String("Hello, world!")),

		itest.GET("/bar?query=foo&other=bar").
			ExpectStatus(404),

		itest.POST("/foo").
			WithHeader("Accept", "application/json"). // add a single header
			WithBody(itest.JSONObject{                // specify the body using Go types
				"key": "value",
			}).
			ExpectStatus(201),

		itest.POST("/foo").
			WithHeaders(http.Header{ // set all headers at once
				"Accept": []string{"application/json"},
				"Auth":   []string{"Bearer 12345"},
			}).
			WithBody(itest.String(`{"key":"value"}`)). // specify body as a string
			ExpectStatus(201),

		itest.DELETE("/foo").
			ExpectStatus(204),

		// use any custom *http.Request for a test
		itest.DO(customReq).
			ExpectStatus(200).
			ExpectBody(itest.String("Hello, world!")),
	})
}

// Simple static webserver for example purposes.
func startExampleServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
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
			body = `{"message":"Hello, World!"}`
		}

		w.Write([]byte(body))
	})

	go http.ListenAndServe("localhost:8080", mux)
}
