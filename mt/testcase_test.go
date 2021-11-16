package mt_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/jefflinse/melatonin/mt"
	"github.com/stretchr/testify/assert"
)

func TestNewTestCase(t *testing.T) {
	tc := mt.NewTestCase("method", "path")
	assert.Equal(t, "method", tc.Method)
	assert.Equal(t, "path", tc.Path)
}

func TestDisplayName(t *testing.T) {
	for _, test := range []struct {
		name     string
		testCase *mt.HTTPTestCase
		expected string
	}{
		{
			name:     "no body",
			testCase: mt.NewTestCase("method", "path"),
			expected: "method path",
		},
		{
			name:     "empty body",
			testCase: mt.NewTestCase("method", "path").WithBody(""),
			expected: "method path",
		},
		{
			name:     "non-empty body",
			testCase: mt.NewTestCase("method", "path").WithBody("body"),
			expected: "method path",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.testCase.DisplayName())
		})
	}
}

func TestConvenienceCtors(t *testing.T) {
	for _, test := range []struct {
		testCase       *mt.HTTPTestCase
		expectedMethod string
	}{
		{
			testCase:       mt.DELETE("path"),
			expectedMethod: "DELETE",
		},
		{
			testCase:       mt.HEAD("path"),
			expectedMethod: "HEAD",
		},
		{
			testCase:       mt.GET("path"),
			expectedMethod: "GET",
		},
		{
			testCase:       mt.OPTIONS("path"),
			expectedMethod: "OPTIONS",
		},
		{
			testCase:       mt.PATCH("path"),
			expectedMethod: "PATCH",
		},
		{
			testCase:       mt.POST("path"),
			expectedMethod: "POST",
		},
		{
			testCase:       mt.PUT("path"),
			expectedMethod: "PUT",
		},
	} {
		t.Run(test.expectedMethod, func(t *testing.T) {
			assert.Equal(t, test.expectedMethod, test.testCase.Method)
		})
	}
}

func TestDO(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	tc := mt.DO(req)
	assert.Equal(t, "GET", tc.Method)
	assert.Equal(t, "", tc.Path)

	req, _ = http.NewRequest("GET", "http://example.com/foo", nil)
	tc = mt.DO(req)
	assert.Equal(t, "GET", tc.Method)
	assert.Equal(t, "/foo", tc.Path)
}

func TestTestCaseConvenienceSetters(t *testing.T) {
	tc := mt.NewTestCase("method", "path")
	tc.WithBody("body")
	assert.Equal(t, "body", tc.RequestBody)
	tc.WithHeaders(map[string][]string{http.CanonicalHeaderKey("header1"): {"value1"}})
	assert.Equal(t, []string{"value1"}, tc.RequestHeaders[http.CanonicalHeaderKey("header1")])
	tc.WithHeader("header2", "value2")
	assert.Equal(t, []string{"value2"}, tc.RequestHeaders[http.CanonicalHeaderKey("header2")])
	tc.WithTimeout(1 * time.Second)
	assert.Equal(t, 1*time.Second, tc.Timeout)

	// setting these again overrides the previous values
	tc.WithBody("body2")
	assert.Equal(t, "body2", tc.RequestBody)
	tc.WithHeaders(map[string][]string{http.CanonicalHeaderKey("header3"): {"value3"}})
	assert.Equal(t, []string{"value3"}, tc.RequestHeaders[http.CanonicalHeaderKey("header3")])
	tc.WithHeader("header4", "value4")
	assert.Equal(t, []string{"value4"}, tc.RequestHeaders[http.CanonicalHeaderKey("header4")])
	tc.WithTimeout(2 * time.Second)
	assert.Equal(t, 2*time.Second, tc.Timeout)

	// WithHeader creates a new map if one does not exist
	tc = mt.NewTestCase("method", "path")
	tc.WithHeader("header1", "value1")
	assert.Equal(t, []string{"value1"}, tc.RequestHeaders[http.CanonicalHeaderKey("header1")])
}

func TestBeforeAfter(t *testing.T) {
	tc := mt.NewTestCase("method", "path")
	assert.Nil(t, tc.BeforeFunc)
	assert.Nil(t, tc.AfterFunc)

	tc.Before(func() error {
		return nil
	})
	assert.NotNil(t, tc.BeforeFunc)
	assert.NoError(t, tc.BeforeFunc())

	tc.After(func() error {
		return nil
	})
	assert.NotNil(t, tc.AfterFunc)
	assert.NoError(t, tc.AfterFunc())

	tc.Before(func() error {
		return assert.AnError
	})
	assert.Error(t, tc.BeforeFunc())

	tc.After(func() error {
		return assert.AnError
	})
	assert.Error(t, tc.AfterFunc())
}

func TestExpectations(t *testing.T) {
	tc := mt.NewTestCase("method", "path")
	tc.ExpectStatus(200)
	assert.Equal(t, 200, tc.WantStatus)
	tc.ExpectHeaders(map[string][]string{http.CanonicalHeaderKey("header1"): {"value1"}})
	assert.Equal(t, []string{"value1"}, tc.WantHeaders[http.CanonicalHeaderKey("header1")])
	tc.ExpectHeader("header2", "value2")
	assert.Equal(t, []string{"value2"}, tc.WantHeaders[http.CanonicalHeaderKey("header2")])
	tc.ExpectBody("body1")
	assert.Equal(t, "body1", tc.WantBody)

	// setting these again overrides the previous values
	tc.ExpectStatus(201)
	assert.Equal(t, 201, tc.WantStatus)
	tc.ExpectHeaders(map[string][]string{http.CanonicalHeaderKey("header3"): {"value3"}})
	assert.Equal(t, []string{"value3"}, tc.WantHeaders[http.CanonicalHeaderKey("header3")])
	tc.ExpectHeader("header4", "value4")
	assert.Equal(t, []string{"value4"}, tc.WantHeaders[http.CanonicalHeaderKey("header4")])
	tc.ExpectBody("body2")
	assert.Equal(t, "body2", tc.WantBody)

	// ExpectHeader creates a new map if one does not exist
	tc = mt.NewTestCase("method", "path")
	tc.ExpectHeader("header1", "value1")
	assert.Equal(t, []string{"value1"}, tc.WantHeaders[http.CanonicalHeaderKey("header1")])
}

func TestValidate(t *testing.T) {
	for _, test := range []struct {
		name        string
		testCase    *mt.HTTPTestCase
		expectError bool
	}{
		{
			name:        "valid",
			testCase:    mt.NewTestCase("method", "/path"),
			expectError: false,
		},
		{
			name:        "invalid, empty method",
			testCase:    mt.NewTestCase("", "/path"),
			expectError: true,
		},
		{
			name:        "invalid, empty path",
			testCase:    mt.NewTestCase("method", ""),
			expectError: true,
		},
		{
			name:        "invalid, path missing leading slash",
			testCase:    mt.NewTestCase("method", "path"),
			expectError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.expectError {
				assert.Error(t, test.testCase.Validate())
			} else {
				assert.NoError(t, test.testCase.Validate())
			}
		})
	}
}
