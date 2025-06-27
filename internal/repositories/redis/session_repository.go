package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/maxdikun/weatherapp/internal/models"
	"github.com/maxdikun/weatherapp/internal/repositories"
)

type SessionRepository struct {
	client *redis.Client
}

var _ repositories.SessionRepository = (*SessionRepository)(nil)

// Add implements repositories.SessionRepository.
func (s *SessionRepository) Add(ctx context.Context, session models.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("redis.SessionRepository.Add: %w", err)
	}

	pipe := s.client.Pipeline()

	pipe.SetNX(ctx, fmt.Sprintf("sessions:%s", session.Id.String()), data, time.Until(session.ExpiresAt))
	pipe.SetNX(ctx, fmt.Sprintf("session_tokens:%s", session.Token), session.Id.String(), time.Until(session.ExpiresAt))

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis.SessionRepository.Add: %w", err)
	}

	return nil
}

// Delete implements repositories.SessionRepository.
func (s *SessionRepository) Delete(ctx context.Context, token string) error {
	sessionId, err := s.client.Get(ctx, fmt.Sprintf("session_tokens:%s", token)).Result()
	if err != nil {
		if err == redis.Nil {
			return &repositories.NotFoundError{
				Object: "session",
				Field:  "token",
			}
		}
		return fmt.Errorf("redis.SessionRepository.Delete: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("session_tokens:%s", token))
	pipe.Del(ctx, fmt.Sprintf("sessions:%s", sessionId))
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis.SessionRepository.Delete: %w", err)
	}
	return nil
}

// FindByToken implements repositories.SessionRepository.
func (s *SessionRepository) FindByToken(ctx context.Context, token string) (models.Session, error) {
	sessionId, err := s.client.Get(ctx, fmt.Sprintf("session_tokens:%s", token)).Result()
	if err != nil {
		if err == redis.Nil {
			return models.Session{}, &repositories.NotFoundError{
				Object: "session",
				Field:  "token",
			}
		}
		return models.Session{}, fmt.Errorf("redis.SessionRepository.FindByToken: %w", err)
	}

	data, err := s.client.Get(ctx, fmt.Sprintf("sessions:%s", sessionId)).Result()
	if err != nil {
		if err == redis.Nil {
			return models.Session{}, &repositories.NotFoundError{
				Object: "session",
				Field:  "token",
			}
		}
		return models.Session{}, fmt.Errorf("redis.SessionRepository.FindByToken: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return models.Session{}, fmt.Errorf("redis.SessionRepository.FindByToken: %w", err)
	}

	return session, nil
}

// Update implements repositories.SessionRepository.
func (s *SessionRepository) Update(ctx context.Context, session models.Session) error {
	return s.Add(ctx, session)
}

func NewSessionRepository(client *redis.Client) *SessionRepository {
	return &SessionRepository{client: client}
}
