package itest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jefflinse/go-itest/itest"
	"github.com/stretchr/testify/assert"
)

func TestNewTestRunner(t *testing.T) {
	r := itest.NewTestRunner("http://example.com")
	assert.NotNil(t, r)
	assert.False(t, r.ContinueOnFailure)
	assert.Same(t, http.DefaultClient, r.HTTPClient)
}

func TestRunnerConvenienceSetters(t *testing.T) {
	r := itest.NewTestRunner("http://example.com")
	r.WithContinueOnFailure(true)
	assert.True(t, r.ContinueOnFailure)
	r.WithHTTPClient(http.DefaultClient)
	assert.Same(t, http.DefaultClient, r.HTTPClient)
	r.WithRequestTimeout(1)
	assert.Equal(t, time.Duration(1), r.RequestTimeout)

	// setting these again overrides the previous values
	r.WithContinueOnFailure(false)
	assert.False(t, r.ContinueOnFailure)
	r.WithHTTPClient(nil)
	assert.Nil(t, r.HTTPClient)
	r.WithRequestTimeout(2)
	assert.Equal(t, time.Duration(2), r.RequestTimeout)
}

func TestRunner_RunTests(t *testing.T) {
	for _, test := range []struct {
		name       string
		statusCode int
		body       itest.Stringable
		withT      bool
	}{
		{
			name:       "success",
			statusCode: 200,
			body:       itest.String("foo"),
			withT:      false,
		},
		{
			name:       "success, test context",
			statusCode: 200,
			body:       itest.String("foo"),
			withT:      true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := newMockServer(test.statusCode, []byte(test.body.String()))
			defer server.Close()
			r := itest.NewTestRunner(server.URL)
			tcs := []*itest.TestCase{
				itest.GET("/foo").ExpectStatus(test.statusCode).ExpectBody(test.body),
				itest.POST("/foo").ExpectStatus(400).ExpectBody(test.body),
			}
			results := r.RunTests(tcs)
			assert.Len(t, results, 2)
			assert.Empty(t, results[0].Errors)
			assert.Len(t, results[1].Errors, 1)
		})
	}
}

func TestRunner_RunTest(t *testing.T) {

}

func TestRunner_RunTestT(t *testing.T) {

}

func TestRunnerValidate(t *testing.T) {
	for _, test := range []struct {
		name        string
		runner      *itest.TestRunner
		expectError bool
	}{
		{
			name: "valid",
			runner: &itest.TestRunner{
				BaseURL: "http://example.com",
			},
			expectError: false,
		},
		{
			name: "invalid, empty base URL",
			runner: &itest.TestRunner{
				BaseURL: "",
			},
			expectError: true,
		},
		{
			name: "invalid, base URL contains trailing slash",
			runner: &itest.TestRunner{
				BaseURL: "http://example.com/",
			},
			expectError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.runner.Validate()
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newMockServer(statusCode int, body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if len(body) > 0 {
			w.Write(body)
		}
	}))
}
