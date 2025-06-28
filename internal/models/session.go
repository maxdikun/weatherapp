package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Id          uuid.UUID
	User        uuid.UUID
	Token       string
	CreatedAt   time.Time
	RefreshedAt time.Time
	ExpiresAt   time.Time
}
