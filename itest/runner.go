package itest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type TestRunner struct {
	// BaseURL is the base URL for the API, including the port.
	//
	// Examples:
	//   http://localhost:8080
	//   https://api.example.com
	BaseURL string

	// ContinueOnFailure indicates whether the test runner should continue
	// executing further tests after a failure. Defaults to false.
	ContinueOnFailure bool

	// HTTPClient is the HTTP client to use for requests.
	// If left unset, http.DefaultClient will be used.
	HTTPClient *http.Client
}

func (r TestRunner) RunTests(tests []TestCase) {
	for i, test := range tests {
		if err := test.validate(i); err != nil {
			log.Fatalf("test case %q is invalid: %s", test.Name, err)
		}

		if test.Setup != nil {
			test.Setup()
		}

		status, body, err := r.doRequest(test.Method, r.BaseURL+test.URI, test.RequestBody)
		if err != nil {
			log.Fatalf("unexpeceted error while running test %q: %s", test.Name, err)
		}

		if status != test.WantStatus {
			log.Fatalf("expected status %d, got %d\n", test.WantStatus, status)
		}

		assertTypeAndValue("", test.WantBody, body)
	}
}

func RunTests(baseURL string, tests []TestCase) {
	runner := TestRunner{
		BaseURL:           baseURL,
		ContinueOnFailure: false,
		HTTPClient:        http.DefaultClient,
	}
	runner.RunTests(tests)
}

func (r TestRunner) doRequest(method, uri string, body Stringable) (int, JSONMap, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader([]byte(body.String()))
	}

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return -1, nil, err
	}

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return -1, nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, err
	}

	var bodyMap JSONMap
	if len(b) > 0 {
		if err := json.Unmarshal(b, &bodyMap); err != nil {
			return -1, nil, err
		}
	}

	return resp.StatusCode, bodyMap, nil
}
