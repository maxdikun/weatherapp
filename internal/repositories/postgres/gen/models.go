// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package gen

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Login    string
	Password string
}
