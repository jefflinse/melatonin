package json_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jefflinse/melatonin/json"
	"github.com/stretchr/testify/assert"
)

func TestGetDeferredValue(t *testing.T) {
	for _, test := range []struct {
		name    string
		have    any
		want    any
		wantErr string
	}{
		{
			"resolve a deferred bool", boolPtr(true), true, "",
		},
		{
			"resolve a deferred float32", float32Ptr(3.14), float32(3.14), "",
		},
		{
			"resolve a deferred float64", float64Ptr(3.14), float64(3.14), "",
		},
		{
			"resolve a deferred int", intPtr(42), int(42), "",
		},
		{
			"resolve a deferred int32", int32Ptr(42), int32(42), "",
		},
		{
			"resolve a deferred int64", int64Ptr(42), int64(42), "",
		},
		{
			"resolve a deferred string", strPtr("foo"), "foo", "",
		},
		{
			"resolve a deferred map",
			map[string]any{"foo": strPtr("bar")},
			map[string]any{"foo": "bar"},
			"",
		},
		{
			"resolve a deferred map with error",
			map[string]any{"foo": func() (any, error) { return nil, fmt.Errorf("error") }},
			nil,
			"foo: error",
		},
		{
			"resolve a deferred slice",
			[]any{strPtr("foo")},
			[]any{"foo"},
			"",
		},
		{
			"resolve a deferred slice with error",
			[]any{func() (any, error) { return nil, fmt.Errorf("error") }},
			nil,
			"[0]: error",
		},
		{
			"resolve a deferred slice with slice with error",
			[]any{[]any{func() (any, error) { return nil, fmt.Errorf("error") }}},
			nil,
			"[0][0]: error",
		},
		{
			"resolve a deferred map with slice with error",
			map[string]any{"foo": []any{func() (any, error) { return nil, fmt.Errorf("error") }}},
			nil,
			"foo[0]: error",
		},

		{
			"resolve a deferred map with map with error",
			map[string]any{"foo": map[string]any{"bar": func() (any, error) { return nil, fmt.Errorf("error") }}},
			nil,
			"foo.bar: error",
		},

		{
			"resolve a deferred map with map with slice with error",
			map[string]any{"foo": map[string]any{"bar": []any{func() (any, error) { return nil, fmt.Errorf("error") }}}},
			nil,
			"foo.bar[0]: error",
		},

		{
			"resolve a deferred function",
			func() any { return "foo" },
			"foo",
			"",
		},
		{
			"resolve a deferred function with no error",
			func() (any, error) { return "foo", nil },
			"foo",
			"",
		},
		{
			"resolve a deferred function with error",
			func() (any, error) { return nil, fmt.Errorf("foo") },
			nil,
			"foo",
		},
		{
			"resolve an unknown type by passing the value through directy",
			struct{}{},
			struct{}{},
			"",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := json.ResolveDeferred(test.have)
			if test.wantErr != "" {
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("got %#v, want %#v", got, test.want)
				}
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func float32Ptr(f float32) *float32 {
	return &f
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
