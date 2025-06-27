package repositories

import (
	"context"

	"github.com/maxdikun/weatherapp/internal/models"
)

type SessionRepository interface {
	FindByToken(ctx context.Context, token string) (models.Session, error)
	Add(ctx context.Context, session models.Session) error
	Delete(ctx context.Context, token string) error
}
