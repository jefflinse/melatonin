package main

import (
	"net/http"

	"github.com/jefflinse/go-itest/itest"
)

func main() {
	startExampleServer()
	itest.RunTests("http://localhost:8080", []*itest.TestCase{

		itest.GET("/foo").
			WithSetup(func() error {
				return nil
			}).
			ExpectStatus(200).
			ExpectBody(itest.JSONMap{"response": "Hello, world!"}),
	})
}

func startExampleServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"response":"Hello, world!"}`))
	})
	go http.ListenAndServe("localhost:8080", mux)
}
