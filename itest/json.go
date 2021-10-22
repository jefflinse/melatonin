package itest

import (
	"encoding/json"
)

const EmptyBody = ""

type Stringable interface {
	String() string
}

// JSONMap is a shorter name for map[string]interface{} that satisfies the Stringable interface.
type JSONMap map[string]interface{}

func (m JSONMap) String() (string, error) {
	b, err := json.Marshal(m)
	return string(b), err
}

// JSONArray is a shorter name for []interface{} that satisfies the Stringable interface.
type JSONArray []interface{}

func (a JSONArray) String() (string, error) {
	b, err := json.Marshal(a)
	return string(b), err
}
