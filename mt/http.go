package mt

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"time"

	mtjson "github.com/jefflinse/melatonin/json"
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

func toBytes(body any) ([]byte, error) {
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

func toInterface(body []byte) any {
	if len(body) > 0 {
		var bodyMap map[string]any
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			return bodyMap
		}

		var bodyArray []any
		if err := json.Unmarshal(body, &bodyArray); err == nil {
			return bodyArray
		}

		return string(body)
	}

	return nil
}

type parameters map[string]any

// Apply maps the values to a target path.
func (p parameters) applyTo(path string) (string, error) {
	resolved, err := mtjson.ResolveDeferred(map[string]any(p))
	if err != nil {
		return "", err
	}

	result := path
	for k, v := range resolved.(map[string]any) {
		str, err := paramString(v)
		if err != nil {
			return "", err
		}

		result = strings.ReplaceAll(result, ":"+k, str)
		if err != nil {
			return "", err
		}
	}

	return result, nil
}

func paramString(v any) (string, error) {
	str := ""
	switch value := v.(type) {
	case bool:
		str = fmt.Sprintf("%t", value)
	case float32, float64:
		str = fmt.Sprintf("%g", value)
	case int, int32, int64, uint, uint32, uint64:
		str = fmt.Sprintf("%d", value)
	case string:
		str = value
	case []string:
		str = strings.Join(value, ",")
	default:
		return "", fmt.Errorf("unsupported parameter type: %T", value)
	}

	return str, nil
}

func (p parameters) asRawQuery() (string, error) {
	resolved, err := mtjson.ResolveDeferred(map[string]any(p))
	if err != nil {
		return "", err
	}

	params := url.Values{}
	for k, v := range resolved.(map[string]any) {
		str, err := paramString(v)
		if err != nil {
			return "", err
		}
		params.Add(k, str)
	}

	return params.Encode(), nil
}
