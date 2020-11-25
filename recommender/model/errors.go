package model

import (
	"fmt"
)

// KeyError is returned when the mapping key was not found.
type KeyError struct {
	key interface{}
}

// NewKeyError returns a new KeyError.
func NewKeyError(key interface{}) KeyError {
	return KeyError{key}
}

func (e KeyError) Error() string {
	return fmt.Sprintf("KeyError: %s", e.key)
}
