package itest

import (
	"encoding/json"
)

type Stringable interface {
	String() string
}

type JSONMap map[string]interface{}

func (m JSONMap) String() string {
	b, err := json.Marshal(m)
	failOnError(err)
	return string(b)
}

type JSONArray []interface{}

func (a JSONArray) String() string {
	b, err := json.Marshal(a)
	failOnError(err)
	return string(b)
}
