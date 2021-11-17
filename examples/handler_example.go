package main

import (
	"net/http"

	"github.com/jefflinse/melatonin/mt"
)

// This example shows how to use melatonin to directly test an HTTP handler locally.
func HandlerExample() {

	// Anything satifying the http.Handler interface can be tested as a handler.
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello, world!"))
	})

	// Use NewHandlerContext() to test a net/http.Handler.
	// No actual network calls are made, making this suitable for unit tests.
	myAPI := mt.NewHandlerContext(mux)
	mt.RunTests([]mt.TestCase{

		myAPI.GET("/foo", "Fetch foo by testing a local handler").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		myAPI.GET("/bar", "This should be a 404").
			ExpectStatus(200),
	})
}
