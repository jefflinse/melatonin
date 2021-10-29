package itest

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

// Object is a type alias for map[string]interface{}.
type Object map[string]interface{}

// Array is a type alias for []interface{}.
type Array []interface{}

func createRequest(method, path string,
	query url.Values,
	headers http.Header,
	body []byte,
	timeout time.Duration) (*http.Request, context.CancelFunc, error) {

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	req, err := http.NewRequestWithContext(ctx, method, path, reader)
	if err != nil {
		return nil, cancel, err
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	if headers != nil {
		req.Header = headers
	} else {
		req.Header = http.Header{}
	}

	return req, cancel, nil
}

func doRequest(c *http.Client, req *http.Request) (int, http.Header, []byte, error) {
	debug("%s %s", req.Method, req.URL.String())
	resp, err := c.Do(req)
	if err != nil {
		return -1, nil, nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, nil, err
	}

	debug("\n")

	return resp.StatusCode, resp.Header, body, nil
}

func handleRequest(h http.Handler, req *http.Request) (int, http.Header, []byte, error) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	resp := w.Result()
	b, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, b, err
}
