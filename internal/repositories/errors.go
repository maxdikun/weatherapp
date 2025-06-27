package repositories

import "fmt"

type AlreadyExistsError struct {
	Object string
	Field  string
}

var _ error = (*AlreadyExistsError)(nil)

// Error implements error.
func (err *AlreadyExistsError) Error() string {
	return fmt.Sprintf("object '%s' already exists with provided value for field '%s'", err.Object, err.Field)
}

type NotFoundError struct {
	Object string
	Field  string
}

var _ error = (*NotFoundError)(nil)

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("object '%s' not found by given value for the field '%s'", err.Object, err.Field)
}
