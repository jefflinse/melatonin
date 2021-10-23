package itest

import (
	"encoding/json"
	"fmt"
)

const EmptyBody = ""

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
