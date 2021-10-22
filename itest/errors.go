package itest

import "fmt"

func WrongTypeError(key string, expected, actual interface{}) error {
	return fmt.Errorf(`%s: expected "%T", got '%T"`, key, expected, actual)
}

func WrongValueError(key string, expected, actual interface{}) error {
	return fmt.Errorf(`%s: expected "%v", got "%v"`, key, expected, actual)
}
