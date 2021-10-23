package itest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert.Equal(t, "foo", String("foo").String())
}

func TestInt(t *testing.T) {
	assert.Equal(t, "1", Int(1).String())
}

func TestFloat(t *testing.T) {
	assert.Equal(t, "1.100000", Float(1.1).String())
}

func TestBool(t *testing.T) {
	assert.Equal(t, "true", Bool(true).String())
}

func TestJSONObject(t *testing.T) {
	assert.Equal(t, `{"array":["foo","bar"],"bool":true,"float":1.1,"int":1,"obj":{"key":"foo"},"string":"foo"}`,
		JSONObject(map[string]interface{}{
			"array":  []interface{}{"foo", "bar"},
			"bool":   true,
			"float":  1.100000,
			"int":    1,
			"obj":    map[string]interface{}{"key": "foo"},
			"string": "foo",
		}).String())
	assert.Panics(t, func() {
		str := JSONObject(map[string]interface{}{"invalid": make(chan int)}).String()
		t.Log(str)
	})
}

func TestJSONArray(t *testing.T) {
	assert.Equal(t, `["foo","bar"]`,
		JSONArray([]interface{}{"foo", "bar"}).String())
	assert.Panics(t, func() {
		str := JSONArray([]interface{}{make(chan int)}).String()
		t.Log(str)
	})
}
