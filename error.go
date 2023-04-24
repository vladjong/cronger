package cronger

import (
	"fmt"
	"reflect"
)

type incorrectClientError struct {
	name   string
	typeOf reflect.Type
}

func newIncorrectClientError(name string, typeOf reflect.Type) *incorrectClientError {
	return &incorrectClientError{
		name:   name,
		typeOf: typeOf,
	}
}

func (e *incorrectClientError) Error() string {
	return fmt.Sprintf("incorrect client=%s, type=%v", e.name, e.typeOf)
}

func (e *incorrectClientError) Unwrap() error {
	return fmt.Errorf("incorrect client=%s, type=%v", e.name, e.typeOf)
}
