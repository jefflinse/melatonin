package main

import (
	"net/http"
	"testing"

	"github.com/jefflinse/melatonin/mt"
)

// This example shows how to use melatonin to directly test an HTTP handler locally.
func TestHandler(t *testing.T) {

	// Anything satifying the http.Handler interface can be tested as a handler.
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello, world!"))
	})

	// Use NetHandlerTester() to test a net/http.Handler.
	// No actual network calls are made, making this suitable for unit tests.
	myAPI := mt.NewHandlerContext(mux)
	mt.RunTestsT(t, []mt.TestCase{

		myAPI.GET("/foo", "Fetch foo by testing a local handler").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		myAPI.GET("/bar", "This should be a 404").
			ExpectStatus(200),
	})
}
