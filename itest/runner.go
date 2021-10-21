package itest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

type TestRunner struct {
	BaseURL    string
	HTTPClient *http.Client
	T          *testing.T
}

func (r TestRunner) RunTests(tests []TestCase) {
	r.T.Helper()
	for _, test := range tests {
		if err := test.validate(); err != nil {
			r.T.Fatalf("Test case %q is invalid: %s", test.Name, err)
		}

		r.T.Run(test.Name, func(t *testing.T) {
			if test.Setup != nil {
				test.Setup(t)
			}

			status, body := doRequest(t, test.Method, r.BaseURL+test.URI, test.RequestBody)
			if status != test.WantStatus {
				t.Fatalf("expected status %d, got %d\n", test.WantStatus, status)
			}

			assertTypeAndValue(t, "<root>", body, test.WantBody)
		})
	}
}

func RunTests(t *testing.T, baseURL string, tests []TestCase) {
	runner := TestRunner{
		BaseURL:    baseURL,
		HTTPClient: http.DefaultClient,
		T:          t,
	}
	runner.RunTests(tests)
}

func doRequest(t *testing.T, method, uri string, body Stringable) (int, JSONMap) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader([]byte(body.String(t)))
	}
	req, err := http.NewRequest(method, uri, reader)
	failOnError(t, err)
	resp, err := http.DefaultClient.Do(req)
	failOnError(t, err)
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	failOnError(t, err)

	var bodyMap JSONMap
	if len(b) > 0 {
		err = json.Unmarshal(b, &bodyMap)
		failOnError(t, err)
	}

	return resp.StatusCode, bodyMap
}
