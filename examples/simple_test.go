package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jefflinse/melatonin/mt"
)

// This example shows how to use melatonin with all the quick, default settings.
func SimpleExampleTest(t *testing.T) {
	server := startSimpleExampleServer()
	defer server.Close()
	myURL := mt.NewURLContext(server.URL)

	mux := createSimpleServeMux()
	myHandler := mt.NewHandlerContext(mux)

	mt.RunTestsT(t, []mt.TestCase{

		myURL.GET("/foo", "Fetch /foo from a URL").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),

		myHandler.GET("/foo", "Fetch /foo from a handler").
			ExpectStatus(200).
			ExpectBody("Hello, world!"),
	})
}

func createSimpleServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello, world!"))
	})

	return mux
}

// Simple static webserver for example purposes.
func startSimpleExampleServer() *httptest.Server {
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

	return httptest.NewServer(mux)
}
