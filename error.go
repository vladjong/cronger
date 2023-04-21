package cronger

import (
	"fmt"
)

type incorrectClientError struct {
	name   string
	typeOf string
}

func newIncorrectClientError(name, typeOf string) *incorrectClientError {
	return &incorrectClientError{
		name:   name,
		typeOf: typeOf,
	}
}

func (e *incorrectClientError) Error() string {
	return fmt.Sprintf("incorrect client=%s, type=%s", e.name, e.typeOf)
}

func (e *incorrectClientError) Unwrap() error {
	return fmt.Errorf("incorrect client=%s, type=%s", e.name, e.typeOf)
}
