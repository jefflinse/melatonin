package main

import (
	"net/http"

	"github.com/jefflinse/go-itest/itest"
)

func main() {
	startExampleServer()
	itest.RunTests("http://localhost:8080", []itest.TestCase{
		{
			Name: "foo",
			Setup: func() error {
				return nil
			},
			Method:     "GET",
			URI:        "/foo",
			WantStatus: 200,
			WantBody:   itest.JSONMap{"response": "Hello, world!"},
		},
	})
}

func startExampleServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"response":"Hello, world!"}`))
	})
	go http.ListenAndServe("localhost:8080", mux)
}
