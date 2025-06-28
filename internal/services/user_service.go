package services

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/maxdikun/weatherapp/internal/models"
	"github.com/maxdikun/weatherapp/internal/repositories"
)

type TokenPair struct {
	Access           string
	AccessExpiresAt  time.Time
	Refresh          string
	RefreshExpiresAt time.Time
}

type UserService struct {
	logger *slog.Logger

	userStorage    repositories.UserRepository
	sessionStorage repositories.SessionRepository

	sessionDuration     time.Duration
	accessTokenDuration time.Duration
	tokenSecret         []byte
}

func (svc *UserService) Register(ctx context.Context, login string, password string) (TokenPair, error) {
	err := errors.Join(svc.validateLogin(login), svc.validatePassword(password))
	if err != nil {
		return TokenPair{}, err
	}

	user, err := svc.createUser(ctx, login, password)
	if err != nil {
		return TokenPair{}, err
	}

	session, err := svc.createSession(ctx, user)
	if err != nil {
		return TokenPair{}, err
	}

	expiresAt := time.Now().Add(svc.accessTokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":  expiresAt,
		"user": session.User,
	})
	tokenString, err := token.SignedString(svc.tokenSecret)
	if err != nil {
		return TokenPair{}, ErrInternal
	}

	return TokenPair{
		Access:           tokenString,
		AccessExpiresAt:  expiresAt,
		Refresh:          session.Token,
		RefreshExpiresAt: session.ExpiresAt,
	}, nil
}

func (svc *UserService) validateLogin(login string) error {
	if len(login) < 3 {
		return &ValidationError{Field: "login", Message: "should be at least 3 characters long"}
	}
	return nil
}

func (svc *UserService) validatePassword(password string) error {
	if len(password) < 6 {
		return &ValidationError{Field: "password", Message: "should be at least 6 characters long"}
	}
	return nil
}

func (svc *UserService) createUser(ctx context.Context, login string, password string) (models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, ErrInternal
	}

	user := models.User{
		Id:       uuid.New(),
		Login:    login,
		Password: string(hashedPassword),
	}

	if err := svc.userStorage.Add(ctx, user); err != nil {
		// TODO: Handle the other errors
		return models.User{}, ErrInternal
	}

	return user, nil
}

func (svc *UserService) createSession(ctx context.Context, user models.User) (models.Session, error) {
	session := models.Session{
		Id:          uuid.New(),
		User:        user.Id,
		Token:       randomString(32),
		CreatedAt:   time.Now(),
		RefreshedAt: time.Now(),
		ExpiresAt:   time.Now().Add(svc.sessionDuration),
	}

	if err := svc.sessionStorage.Add(ctx, session); err != nil {
		// TODO: Handle the other errors
		return models.Session{}, ErrInternal
	}

	return session, nil
}

func NewUserService(
	logger *slog.Logger,
	userStorage repositories.UserRepository,
	sessionStorage repositories.SessionRepository,
	sessionDuration time.Duration,
	accessTokenDuration time.Duration,
	tokenSecret []byte,
) *UserService {
	return &UserService{
		logger:              logger,
		userStorage:         userStorage,
		sessionStorage:      sessionStorage,
		sessionDuration:     sessionDuration,
		accessTokenDuration: accessTokenDuration,
		tokenSecret:         tokenSecret,
	}
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
