package main

import "net/http"

func main() {
	HandlerExample()
	EndpointExample()
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
