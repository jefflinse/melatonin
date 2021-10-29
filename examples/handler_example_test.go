package main

import (
	"net/http"
	"testing"

	"github.com/jefflinse/go-itest/itest"
)

// This example shows how to use itest to directly test an HTTP handler locally.
func TestHandler(t *testing.T) {

	// Anything satifying the http.Handler interface can be tested as a handler.
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello, world!"))
	})

	// Use NetHandlerTester() to test a net/http.Handler.
	// No actual network calls are made, making this suitable for unit tests.
	runner := itest.NewHandlerTester(mux).WithContinueOnFailure(true)
	runner.RunTestsT(t, []*itest.TestCase{

		itest.GET("/foo", "Fetch foo by testing a local handler").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		itest.GET("/bar", "This should be a 404").
			ExpectStatus(200),
	})
}
