package cronger

import (
	"fmt"
)

type IncorrectClientError struct {
	name   string
	typeOf string
}

func NewIncorrectClientError(name, typeOf string) *IncorrectClientError {
	return &IncorrectClientError{
		name:   name,
		typeOf: typeOf,
	}
}

func (e *IncorrectClientError) Error() string {
	return fmt.Sprintf("incorrect client=%s, type=%s", e.name, e.typeOf)
}

func (e *IncorrectClientError) Unwrap() error {
	return fmt.Errorf("incorrect client=%s, type=%s", e.name, e.typeOf)
}
