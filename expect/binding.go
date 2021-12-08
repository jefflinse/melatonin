package expect

import (
	"fmt"
	"log"
)

type Values map[string]interface{}

func (c Values) BindBytes(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		c[name] = new([]byte)
	}

	if b, ok := c[name].(*[]byte); ok {
		return Bind(b)
	}

	return func(string, interface{}) error {
		return fmt.Errorf("can't bind byte slice value to %q (previously bound as %T)", name, v)
	}
}

func (c Values) GetBytes(name string) []byte {
	v, ok := c[name]
	if !ok {
		return nil
	}

	if b, ok := v.(*[]byte); ok {
		return *b
	}

	return nil
}

func (c Values) BindInt(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		c[name] = new(int64)
	}

	if n, ok := c[name].(*int64); ok {
		return Bind(n)
	}

	return func(string, interface{}) error {
		return fmt.Errorf("can't bind int value to %q (previously bound as %T)", name, v)
	}
}

func (c Values) GetInt(name string) int64 {
	v, ok := c[name]
	if !ok {
		return 0
	}

	if n, ok := v.(*int64); ok {
		return *n
	}

	if f, ok := v.(*float64); ok {
		if n, ok := floatToInt(*f); ok {
			log.Printf("coerced float %f to int %d", *f, n)
			return n
		}
	}

	return 0
}

func (c Values) BindFloat(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		c[name] = new(float64)
	}

	if n, ok := c[name].(*float64); ok {
		return Bind(n)
	}

	return func(string, interface{}) error {
		return fmt.Errorf("can't bind float value to %q (previously bound as %T)", name, v)
	}
}

func (c Values) GetFloat(name string) float64 {
	v, ok := c[name]
	if !ok {
		return 0
	}

	if f, ok := v.(*float64); ok {
		return *f
	}

	if n, ok := v.(*int64); ok {
		return float64(*n)
	}

	return 0
}

func (c Values) BindString(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		c[name] = new(string)
	}

	if s, ok := c[name].(*string); ok {
		return Bind(s)
	}

	return func(string, interface{}) error {
		return fmt.Errorf("can't bind string value to %q (previously bound as %T)", name, v)
	}
}

func (c Values) GetString(name string) string {
	v, ok := c[name]
	if !ok {
		return ""
	}

	if s, ok := v.(*string); ok {
		return *s
	}

	if b, ok := v.(*[]byte); ok {
		return string(*b)
	}

	return ""
}

func (c Values) BindJSONArray(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		c[name] = new([]interface{})
	}

	if s, ok := c[name].(*[]interface{}); ok {
		return Bind(s)
	}

	return func(string, interface{}) error {
		return fmt.Errorf("can't bind JSON array to %q (previously bound as %T)", name, v)
	}
}

func (c Values) GetJSONArray(name string) []interface{} {
	v, ok := c[name]
	if !ok {
		return nil
	}

	if s, ok := v.(*[]interface{}); ok {
		return *s
	}

	return nil
}

func (c Values) BindJSONObject(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		c[name] = new(map[string]interface{})
	}

	if s, ok := c[name].(*map[string]interface{}); ok {
		return Bind(s)
	}

	return func(string, interface{}) error {
		return fmt.Errorf("can't bind JSON object to %q (previously bound as %T)", name, v)
	}
}

func (c Values) GetJSONObject(name string) map[string]interface{} {
	v, ok := c[name]
	if !ok {
		return nil
	}

	if o, ok := v.(*map[string]interface{}); ok {
		return *o
	}

	return nil
}
