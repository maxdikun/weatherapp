package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maxdikun/weatherapp/internal/models"
	"github.com/maxdikun/weatherapp/internal/repositories"
	"github.com/maxdikun/weatherapp/internal/repositories/postgres/gen"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

var _ repositories.UserRepository = (*UserRepository)(nil)

// Add implements repositories.UserRepository.
func (u *UserRepository) Add(ctx context.Context, user models.User) error {
	queries := gen.New(u.pool)

	_, err := queries.InsertUser(ctx, gen.InsertUserParams{
		ID:       user.Id,
		Login:    user.Login,
		Password: user.Password,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return &repositories.AlreadyExistsError{
					Object: "user",
					Field:  "login",
				}
			}
		}
		return fmt.Errorf("postgres.userRepository.Add: %w", err)
	}

	return nil
}

// FindById implements repositories.UserRepository.
func (u *UserRepository) FindById(ctx context.Context, id uuid.UUID) (models.User, error) {
	queries := gen.New(u.pool)

	result, err := queries.SelectUserById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, &repositories.NotFoundError{
				Object: "user",
				Field:  "id",
			}
		}

		return models.User{}, fmt.Errorf("postgres.UserRepository.FindById: %w", err)
	}

	return models.User{
		Id:       result.ID,
		Login:    result.Login,
		Password: result.Password,
	}, nil
}

// FindByLogin implements repositories.UserRepository.
func (u *UserRepository) FindByLogin(ctx context.Context, login string) (models.User, error) {
	queries := gen.New(u.pool)

	result, err := queries.SelectUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, &repositories.NotFoundError{
				Object: "user",
				Field:  "login"}
		}

		return models.User{}, fmt.Errorf("postgres.UserRepository.FindByLogin: %w", err)
	}

	return models.User{
		Id:       result.ID,
		Login:    result.Login,
		Password: result.Password,
	}, nil
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}
