package json

import (
	"fmt"
	"strings"
)

// Object is a type alias for map[string]interface{}.
type Object map[string]interface{}

// Array is a type alias for []interface{}.
type Array []interface{}

// A DeferredValueError is an error type that is returned when a deferred value
// cannot be resolved.
type DeferredValueError struct {
	Label string
	Err   error
}

func (dve DeferredValueError) Error() string {
	return fmt.Sprintf("%s: %s", dve.Label, dve.Err.Error())
}

// WithPrefix prefixes the error's label with an additional string.
func (dve DeferredValueError) WithPrefix(prefix string) DeferredValueError {
	return DeferredValueError{
		Label: prefix + dve.Label,
		Err:   dve.Err,
	}
}

// ResolveDeferred resolves a concrete value from a number of different input types,
// such as pointers, functions, and maps and slices potentially containing more
// deferred values.
func ResolveDeferred(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case func() interface{}:
		return value(), nil
	case func() (interface{}, error):
		return value()
	case map[string]interface{}:
		mapVal, err := getDeferredMapValue(value)
		if err != nil {
			return nil, err
		}
		return mapVal, nil
	case []interface{}:
		mapVal, err := getDeferredSliceValue(value)
		if err != nil {
			return nil, err
		}
		return mapVal, nil
	case *bool:
		return *value, nil
	case *float32:
		return *value, nil
	case *float64:
		return *value, nil
	case *int:
		return *value, nil
	case *int32:
		return *value, nil
	case *int64:
		return *value, nil
	case *string:
		return *value, nil

	default:
		return value, nil
	}
}

func getDeferredMapValue(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		value, err := ResolveDeferred(v)
		if err != nil {
			if dve, ok := err.(DeferredValueError); ok {
				prefix := k

				// hacky
				if !strings.HasPrefix(dve.Label, "[") {
					prefix += "."
				}

				return nil, dve.WithPrefix(prefix)
			}
			return nil, DeferredValueError{k, err}
		}

		result[k] = value
	}

	return result, nil
}

func getDeferredSliceValue(s []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(s))
	for i, v := range s {
		value, err := ResolveDeferred(v)
		if err != nil {
			if dve, ok := err.(DeferredValueError); ok {
				return nil, dve.WithPrefix(fmt.Sprintf("[%d]", i))
			}
			return nil, DeferredValueError{fmt.Sprintf("[%d]", i), err}
		}

		result[i] = value
	}

	return result, nil
}
