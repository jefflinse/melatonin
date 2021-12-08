package expect

import (
	"fmt"
	"log"
)

type BoundValueCollection map[string]interface{}

func (c BoundValueCollection) BindBytes(name string) CustomPredicateForKey {
	if v, ok := c[name]; !ok {
		v = []byte{}
		c[name] = &v
	}

	return Bind(c[name])
}

func (c BoundValueCollection) GetBytes(name string) []byte {
	if v, ok := c[name]; ok {
		if b, ok := v.(*[]byte); ok {
			return *b
		}

		return nil
	}

	return nil
}

func (c BoundValueCollection) BindInt(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		v = new(int64)
		c[name] = v
	}

	n, ok := v.(*int64)
	if !ok {
		return func(string, interface{}) error {
			return fmt.Errorf("can't bind int value to %q (previously bound as %T)", name, v)
		}
	}

	return Bind(n)
}

func (c BoundValueCollection) GetInt(name string) int64 {
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

func (c BoundValueCollection) BindFloat(name string) CustomPredicateForKey {
	v, ok := c[name]
	if !ok {
		v = new(float64)
		c[name] = v
	}

	f, ok := v.(*float64)
	if !ok {
		return func(string, interface{}) error {
			return fmt.Errorf("can't bind float value to %q (previously bound as %T)", name, v)
		}
	}

	return Bind(f)
}

func (c BoundValueCollection) GetFloat(name string) float64 {
	if v, ok := c[name]; ok {
		if f, ok := v.(*float64); ok {
			return *f
		}

		return 0.0
	}

	return 0.0
}

func (c BoundValueCollection) BindString(name string) CustomPredicateForKey {
	if _, ok := c[name]; !ok {
		s := ""
		c[name] = &s
	}

	return Bind(c[name])
}

func (c BoundValueCollection) GetString(name string) string {
	if v, ok := c[name]; ok {
		if s, ok := v.(*string); ok {
			return *s
		}

		return ""
	}

	return ""
}

func (c BoundValueCollection) BindJSONArray(name string) CustomPredicateForKey {
	if _, ok := c[name]; !ok {
		c[name] = &[]interface{}{}
	}

	return Bind(c[name])
}

func (c BoundValueCollection) GetJSONArray(name string) []interface{} {
	if v, ok := c[name]; ok {
		if a, ok := v.(*[]interface{}); ok {
			return *a
		}

		return nil
	}

	return nil
}

func (c BoundValueCollection) BindJSONObject(name string) CustomPredicateForKey {
	if _, ok := c[name]; !ok {
		c[name] = &map[string]interface{}{}
	}

	return Bind(c[name])
}

func (c BoundValueCollection) GetJSONObject(name string) map[string]interface{} {
	if v, ok := c[name]; ok {
		if o, ok := v.(*map[string]interface{}); ok {
			return *o
		}

		return nil
	}

	return nil
}
