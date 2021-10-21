package itest

import (
	"encoding/json"
	"testing"
)

type Stringable interface {
	String(t *testing.T) string
}

type JSONMap map[string]interface{}

func (m JSONMap) String(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(m)
	failOnError(t, err)
	return string(b)
}

type JSONArray []interface{}

func (a JSONArray) String(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(a)
	failOnError(t, err)
	return string(b)
}
