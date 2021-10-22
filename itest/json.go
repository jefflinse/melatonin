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

func (m JSONMap) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		panic("failed to marshal JSONMap to string: " + err.Error())
	}
	return string(b)
}

// JSONArray is a shorter name for []interface{} that satisfies the Stringable interface.
type JSONArray []interface{}

func (a JSONArray) String() string {
	b, err := json.Marshal(a)
	if err != nil {
		panic("failed to marshal JSONArray to string: " + err.Error())
	}
	return string(b)
}
