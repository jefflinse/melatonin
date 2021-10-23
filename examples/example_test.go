package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/jefflinse/go-itest/itest"
)

func TestAPI(t *testing.T) {
	startExampleServer()
	// itest.Verbose = true

	customReq, _ := http.NewRequest("GET", "http://localhost:8080/foo", nil)

	itest.RunTestsT(t, "http://localhost:8080", []*itest.TestCase{

		itest.GET("/foo").
			WithHeader("Accept", "application/json").
			WithTimeout(1 * time.Second).
			ExpectStatus(200).
			ExpectBody(itest.JSONObject{"response": "Hello, world!"}),

		itest.GET("/bar").
			ExpectStatus(404),

		// Specify a custom *http.Request for a test.
		// Caller is responsible for constructing the request and ensuring it is valid.
		itest.DO(customReq).
			ExpectStatus(200).
			ExpectBody(itest.JSONObject{"response": "Hello, world!"}),
	})
}
