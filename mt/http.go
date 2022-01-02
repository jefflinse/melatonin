package mt

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

var (
	defaultRequestTimeout = 10 * time.Second
)

func init() {
	envTimeoutStr := os.Getenv("MELATONIN_DEFAULT_TEST_TIMEOUT")
	if envTimeoutStr != "" {
		if timeout, err := time.ParseDuration(envTimeoutStr); err == nil {
			defaultRequestTimeout = timeout
		} else {
			fmt.Printf("invalid MELATONIN_DEFAULT_TEST_TIMEOUT value %q in environment, using default of %s\n",
				envTimeoutStr, defaultRequestTimeout)
		}
	}
}

// DELETE is a shortcut for DefaultContext().NewTestCase(http.MethodDelete, path).
func DELETE(url string, description ...string) *HTTPTestCase {
	return DefaultContext().DELETE(url, description...)
}

// HEAD is a shortcut for NewTestCase(http.MethodHead, path, description...).
func HEAD(url string, description ...string) *HTTPTestCase {
	return DefaultContext().HEAD(url, description...)
}

// GET is a shortcut for NewTestCase(http.MethodGet, path, description...).
func GET(url string, description ...string) *HTTPTestCase {
	return DefaultContext().GET(url, description...)
}

// OPTIONS is a shortcut for NewTestCase(http.MethodOptions, path, description...).
func OPTIONS(url string, description ...string) *HTTPTestCase {
	return DefaultContext().OPTIONS(url, description...)
}

// PATCH is a shortcut for NewTestCase(http.MethodPatch, path, description...).
func PATCH(url string, description ...string) *HTTPTestCase {
	return DefaultContext().PATCH(url, description...)
}

// POST is a shortcut for NewTestCase(http.MethodPost, path, description...).
func POST(url string, description ...string) *HTTPTestCase {
	return DefaultContext().POST(url, description...)
}

// PUT is a shortcut for NewTestCase(http.MethodPut, path, description...).
func PUT(url string, description ...string) *HTTPTestCase {
	return DefaultContext().PUT(url, description...)
}

// DO creates a test case from a custom HTTP request.
func DO(request *http.Request, description ...string) *HTTPTestCase {
	tc := DefaultContext().newHTTPTestCase(request.Method, request.URL.Path, description...)
	tc.request = request
	return tc
}

func createRequest(method, path string) (*http.Request, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	req, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return nil, cancel, err
	}

	return req, cancel, nil
}

func doRequest(c *http.Client, req *http.Request) (int, http.Header, []byte, error) {
	resp, err := c.Do(req)
	if err != nil {
		return -1, nil, nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, nil, err
	}

	return resp.StatusCode, resp.Header, body, nil
}

func handleRequest(h http.Handler, req *http.Request) (int, http.Header, []byte, error) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	resp := w.Result()
	b, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, b, err
}

func toBytes(body interface{}) ([]byte, error) {
	var b []byte
	if body != nil {
		var err error
		switch v := body.(type) {
		case []byte:
			b = v
		case string:
			b = []byte(v)
		case func() []byte:
			b = v()
		case func() ([]byte, error):
			b, err = v()
		default:
			b, err = json.Marshal(body)
		}

		if err != nil {
			return nil, fmt.Errorf("request body: %w", err)
		}
	}

	return b, nil
}

func toInterface(body []byte) interface{} {
	if len(body) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			return bodyMap
		}

		var bodyArray []interface{}
		if err := json.Unmarshal(body, &bodyArray); err == nil {
			return bodyArray
		}

		return string(body)
	}

	return nil
}

type pathParameters map[string]interface{}

// Apply maps the path parameters to a request path.
//
//
func (p pathParameters) Apply(path string) (string, error) {
	var err error
	for k, v := range p {
		path, err = p.applyPathParam(path, k, v)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

func (p pathParameters) applyPathParam(path, k string, v interface{}) (string, error) {
	expanded := ""
	switch value := v.(type) {
	case string:
		expanded = value
	case *string:
		if value == nil {
			return "", fmt.Errorf("path parameter %q: cannot be nil", k)
		}
		return p.applyPathParam(path, k, *value)
	case int:
		return p.applyPathParam(path, k, int64(value))
	case *int:
		return p.applyPathParam(path, k, int64(*value))
	case *int64:
		if value == nil {
			return "", fmt.Errorf("path parameter %q: cannot be nil", k)
		}
		return p.applyPathParam(path, k, *value)
	case int64:
		expanded = fmt.Sprintf("%d", value)
	case *float64:
		if value == nil {
			return "", fmt.Errorf("path parameter %q: cannot be nil", k)
		}
		return p.applyPathParam(path, k, *value)
	case float64:
		expanded = fmt.Sprintf("%g", value)
	}

	return strings.ReplaceAll(path, ":"+k, expanded), nil
}
