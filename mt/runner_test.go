package mt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jefflinse/melatonin/mt"
	"github.com/stretchr/testify/assert"
)

func TestNewTestRunner(t *testing.T) {
	r := mt.NewURLTester("http://example.com")
	assert.NotNil(t, r)
	assert.False(t, r.ContinueOnFailure)
}

func TestRunnerConvenienceSetters(t *testing.T) {
	r := mt.NewURLTester("http://example.com")
	r.WithContinueOnFailure(true)
	assert.True(t, r.ContinueOnFailure)
	r.WithHTTPClient(http.DefaultClient)
	r.WithRequestTimeout(1)
	assert.Equal(t, time.Duration(1), r.TestTimeout)

	// setting these again overrides the previous values
	r.WithContinueOnFailure(false)
	assert.False(t, r.ContinueOnFailure)
	r.WithHTTPClient(nil)
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
		runner      *mt.TestRunner
		tests       []*mt.HTTPTestCase
		wantResults []*mt.TestCaseResult
		wantError   bool
	}{
		{
			name:      "invalid test runner",
			runner:    &mt.TestRunner{},
			wantError: true,
		},
		{
			name:      "invalid tests",
			runner:    mt.NewURLTester(mockServer.URL),
			tests:     []*mt.HTTPTestCase{mt.GET("")},
			wantError: true,
		},
		{
			name:        "nil HTTP client, use default",
			server:      mockServer,
			runner:      mt.NewURLTester(mockServer.URL),
			tests:       []*mt.HTTPTestCase{mt.GET("/path")},
			wantResults: []*mt.TestCaseResult{{TestCase: mt.GET("/path"), Status: http.StatusOK}},
		},
		{
			name:   "all tests pass",
			server: mockServer,
			runner: mt.NewURLTester(mockServer.URL).WithContinueOnFailure(true),
			tests: []*mt.HTTPTestCase{
				mt.GET("/path").ExpectStatus(http.StatusOK),
				mt.GET("/path").ExpectStatus(http.StatusOK),
				mt.GET("/path").ExpectStatus(http.StatusOK),
			},
			wantResults: []*mt.TestCaseResult{
				{Status: http.StatusOK},
				{Status: http.StatusOK},
				{Status: http.StatusOK},
			},
		},
		{
			name:   "test failure",
			server: mockServer,
			runner: mt.NewURLTester(mockServer.URL).WithContinueOnFailure(true),
			tests: []*mt.HTTPTestCase{
				mt.GET("/path").ExpectStatus(http.StatusOK),
				mt.GET("/path").ExpectStatus(http.StatusNotFound),
				mt.GET("/path").ExpectStatus(http.StatusOK),
			},
			wantResults: []*mt.TestCaseResult{
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
		runner      *mt.TestRunner
		expectError bool
	}{
		{
			name:        "valid",
			runner:      mt.NewURLTester("http://example.com"),
			expectError: false,
		},
		{
			name:        "invalid, empty base URL",
			runner:      mt.NewURLTester(""),
			expectError: true,
		},
		{
			name:        "invalid, base URL contains trailing slash",
			runner:      mt.NewURLTester("http://example.com/"),
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
