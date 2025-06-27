package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxdikun/weatherapp/internal/models"
)

type UserRepository interface {
	FindById(ctx context.Context, id uuid.UUID) (models.User, error)
	FindByLogin(ctx context.Context, login string) (models.User, error)
	Add(ctx context.Context, user models.User) error
}
