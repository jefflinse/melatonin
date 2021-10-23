package itest

import (
	"encoding/json"
	"fmt"
)

type Stringable interface {
	String() string
}

type String string

func (s String) String() string {
	return string(s)
}

type Int int64

func (i Int) String() string {
	return fmt.Sprintf("%d", i)
}

type Float float64

func (f Float) String() string {
	return fmt.Sprintf("%f", f)
}

type Bool bool

func (b Bool) String() string {
	return fmt.Sprintf("%t", b)
}

// JSONObject is a shorter name for map[string]interface{} that satisfies the Stringable interface.
type JSONObject map[string]interface{}

func (m JSONObject) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		panic("failed to marshal JSONObject to string: " + err.Error())
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
