package services

import (
	"context"
	"errors"
	"log/slog"
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

func (svc *UserService) Register(ctx context.Context, login string, password string) (TokenPair, error) {
	if err := errors.Join(svc.validateLogin(login), svc.validatePassword(password)); err != nil {
		return TokenPair{}, err
	}

	user, err := svc.createUser(ctx, login, password)
	if err != nil {
		return TokenPair{}, err
	}

	return svc.generateSessionTokens(ctx, user)
}

func (svc *UserService) Login(ctx context.Context, login string, password string) (TokenPair, error) {
	if err := errors.Join(svc.validateLogin(login), svc.validatePassword(password)); err != nil {
		return TokenPair{}, err
	}

	user, err := svc.findUser(ctx, login)
	if err != nil {
		return TokenPair{}, err
	}

	return svc.generateSessionTokens(ctx, user)
}

func (svc *UserService) RefreshSession(ctx context.Context, oldToken string) (TokenPair, error) {
	session, err := svc.sessionStorage.FindByToken(ctx, oldToken)
	if err != nil {
		var notFound *repositories.NotFoundError
		if errors.As(err, &notFound) {
			return TokenPair{}, ErrInvalidToken
		}

		return TokenPair{}, ErrInternal
	}

	session.Token = randomString(32)
	session.RefreshedAt = time.Now()
	session.ExpiresAt = session.RefreshedAt.Add(svc.sessionDuration)

	if err := svc.sessionStorage.Update(ctx, session); err != nil {
		return TokenPair{}, ErrInternal
	}

	panic("unimplemented")
}

func (svc *UserService) findUser(ctx context.Context, login string) (models.User, error) {
	user, err := svc.userStorage.FindByLogin(ctx, login)
	if err != nil {
		var notFound *repositories.NotFoundError
		if errors.As(err, &notFound) {
			return models.User{}, ErrInvalidCredentials
		}
		return models.User{}, ErrInvalidCredentials
	}
	return user, nil
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
		var alreadyExists *repositories.AlreadyExistsError
		if errors.As(err, &alreadyExists) {
			return models.User{}, ErrUserAlreadyExists
		}
		return models.User{}, ErrInternal
	}

	return user, nil
}

func (svc *UserService) generateSessionTokens(ctx context.Context, user models.User) (TokenPair, error) {
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
		return TokenPair{}, ErrInternal
	}

	expiresAt := time.Now().Add(svc.accessTokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": expiresAt,
		"sub": session.User.String(),
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
