package services

import (
	"errors"
	"fmt"
)

var (
	ErrInternal = errors.New("internal service error")
)

type ValidationError struct {
	Field   string
	Message string
}

var _ error = (*ValidationError)(nil)

// Error implements error.
func (err *ValidationError) Error() string {
	return fmt.Sprintf("field '%s' is invalid: %s", err.Field, err.Message)
}
