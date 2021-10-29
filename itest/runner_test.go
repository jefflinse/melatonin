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
	r := itest.NewEndpointTester("http://example.com")
	assert.NotNil(t, r)
	assert.False(t, r.ContinueOnFailure)
}

func TestRunnerConvenienceSetters(t *testing.T) {
	r := itest.NewEndpointTester("http://example.com")
	r.WithContinueOnFailure(true)
	assert.True(t, r.ContinueOnFailure)
	r.WithHTTPClient(http.DefaultClient)
	assert.Same(t, http.DefaultClient, r.HTTPClient)
	r.WithRequestTimeout(1)
	assert.Equal(t, time.Duration(1), r.TestTimeout)

	// setting these again overrides the previous values
	r.WithContinueOnFailure(false)
	assert.False(t, r.ContinueOnFailure)
	r.WithHTTPClient(nil)
	assert.Nil(t, r.HTTPClient)
	r.WithRequestTimeout(2)
	assert.Equal(t, time.Duration(2), r.TestTimeout)
}

func TestRunner_RunTestsT(t *testing.T) {
	mockServer := newMockServer(http.StatusOK, nil)
	mockServer.Start()
	defer mockServer.Close()

	for _, test := range []struct {
		name        string
		server      *httptest.Server
		runner      *itest.TestRunner
		tests       []*itest.TestCase
		wantResults []*itest.TestCaseResult
		wantError   bool
	}{
		{
			name:      "invalid test runner",
			runner:    &itest.TestRunner{},
			wantError: true,
		},
		{
			name:      "invalid tests",
			runner:    itest.NewEndpointTester(mockServer.URL),
			tests:     []*itest.TestCase{itest.GET("")},
			wantError: true,
		},
		{
			name:        "nil HTTP client, use default",
			server:      mockServer,
			runner:      itest.NewEndpointTester(mockServer.URL),
			tests:       []*itest.TestCase{itest.GET("/path")},
			wantResults: []*itest.TestCaseResult{{TestCase: itest.GET("/path"), Status: http.StatusOK}},
		},
		{
			name:   "all tests pass",
			server: mockServer,
			runner: itest.NewEndpointTester(mockServer.URL).WithContinueOnFailure(true),
			tests: []*itest.TestCase{
				itest.GET("/path").ExpectStatus(http.StatusOK),
				itest.GET("/path").ExpectStatus(http.StatusOK),
				itest.GET("/path").ExpectStatus(http.StatusOK),
			},
			wantResults: []*itest.TestCaseResult{
				{Status: http.StatusOK},
				{Status: http.StatusOK},
				{Status: http.StatusOK},
			},
		},
		{
			name:   "test failure",
			server: mockServer,
			runner: itest.NewEndpointTester(mockServer.URL).WithContinueOnFailure(true),
			tests: []*itest.TestCase{
				itest.GET("/path").ExpectStatus(http.StatusOK),
				itest.GET("/path").ExpectStatus(http.StatusNotFound),
				itest.GET("/path").ExpectStatus(http.StatusOK),
			},
			wantResults: []*itest.TestCaseResult{
				{Status: http.StatusOK},
				{Status: http.StatusOK, Errors: []error{assert.AnError}},
				{Status: http.StatusOK},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			results, err := test.runner.RunTests(test.tests)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, results, len(test.wantResults))

			for i := 0; i < len(test.wantResults); i++ {
				assert.Equal(t, test.wantResults[i].Status, results[i].Status)
			}
		})
	}
}

func TestRunnerValidate(t *testing.T) {
	for _, test := range []struct {
		name        string
		runner      *itest.TestRunner
		expectError bool
	}{
		{
			name:        "valid",
			runner:      itest.NewEndpointTester("http://example.com"),
			expectError: false,
		},
		{
			name:        "invalid, empty base URL",
			runner:      itest.NewEndpointTester(""),
			expectError: true,
		},
		{
			name:        "invalid, base URL contains trailing slash",
			runner:      itest.NewEndpointTester("http://example.com/"),
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

func newMockServer(statusCode int, respBody []byte) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if len(respBody) > 0 {
			w.Write(respBody)
		}
	})
	return httptest.NewUnstartedServer(mux)
}
