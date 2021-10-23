package main

import (
	"net/http"

	"github.com/jefflinse/go-itest/itest"
)

func main() {
	startExampleServer()
	// itest.Verbose = true

	itest.RunTests("http://localhost:8080", []*itest.TestCase{

		itest.GET("/foo").
			WithHeader("Accept", "application/json").
			ExpectStatus(200).
			ExpectBody(itest.JSONObject{"response": "Hello, world!"}),

		itest.GET("/bar").
			ExpectStatus(404),

		// Specify a custom *http.Request for a test
		itest.DO(http.NewRequest("GET", "http://localhost:8080/foo", nil)).
			ExpectStatus(200).
			ExpectBody(itest.JSONObject{"response": "Hello, world!"}),
	})
}

func startExampleServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"response":"Hello, world!"}`))
	})
	go http.ListenAndServe("localhost:8080", mux)
}
