package bind

import (
	"encoding/json"
	"fmt"

	"github.com/jefflinse/melatonin/expect"
)

// Bool creates a predicate requiring a value to be a bool,
// binding the value to a target variable.
func Bool(target *bool) expect.Predicate {
	if target == nil {
		return expect.Bool()
	}

	return expect.Bool().Then(func(actual any) error {
		*target = actual.(bool)
		return nil
	})
}

// Int creates a predicate requiring a value to be an integer,
// binding the value to a target variable.
func Int(target *int64) expect.Predicate {
	if target == nil {
		return expect.Int()
	}

	return expect.Int().Then(func(actual any) error {
		if v, ok := actual.(int64); ok {
			*target = v
		} else if v, ok := actual.(float64); ok {
			*target = int64(v)
		} else {
			return fmt.Errorf("expected to bind int64, found %T", actual)
		}

		return nil
	})
}

// Float creates a predicate requiring a value to be a floating point number,
// binding the value to a target variable.
func Float(target *float64) expect.Predicate {
	if target == nil {
		return expect.Float()
	}

	return expect.Float().Then(func(actual any) error {
		*target = actual.(float64)
		return nil
	})
}

// Map creates a predicate requiring a value to be a map,
// binding the value to a target variable.
func Map(target *map[string]any) expect.Predicate {
	if target == nil {
		return expect.Map()
	}

	return expect.Map().Then(func(actual any) error {
		*target = actual.(map[string]any)
		return nil
	})
}

// Slice creates a predicate requiring a value to be a slice,
// binding the value to a target variable.
func Slice(target *[]any) expect.Predicate {
	if target == nil {
		return expect.Slice()
	}

	return expect.Slice().Then(func(actual any) error {
		*target = actual.([]any)
		return nil
	})
}

// String creates a predicate requiring a value to be a string,
// binding the value to a target variable.
func String(target *string) expect.Predicate {
	if target == nil {
		return expect.String()
	}

	return expect.String().Then(func(actual any) error {
		*target = actual.(string)
		return nil
	})
}

// Struct creates a predicate requiring a value to be a struct,
// binding the value to a target variable by unmarshaling the
// JSON representation of the value into the target variable.
func Struct(target any) expect.Predicate {
	return func(actual any) error {
		b, err := json.Marshal(actual)
		if err != nil {
			return fmt.Errorf("failed to bind %T to %T: %w", actual, target, err)
		}

		if err := json.Unmarshal(b, target); err != nil {
			return fmt.Errorf("failed to bind %T to %T: %w", actual, target, err)
		}

		return nil
	}
}
