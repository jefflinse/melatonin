package itest

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func (r *TestRunner) createRequest(method, uri string,
	headers http.Header,
	body []byte,
	timeout time.Duration) (*http.Request, context.CancelFunc, error) {

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	req, err := http.NewRequestWithContext(ctx, method, uri, reader)
	if err != nil {
		return nil, cancel, err
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
