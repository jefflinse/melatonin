package main

import "net/http"

func startExampleServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"response":"Hello, world!"}`))
	})
	go http.ListenAndServe("localhost:8080", mux)
}
