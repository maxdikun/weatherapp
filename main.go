package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/maxdikun/weatherapp/internal/handlers"
	"github.com/maxdikun/weatherapp/internal/repositories/postgres"
	redisRepo "github.com/maxdikun/weatherapp/internal/repositories/redis"
	"github.com/maxdikun/weatherapp/internal/services"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := LoadConfig()
	if err != nil {
		logger.Error("failed to parse configuration", "err", err)
		return
	}

	logger.Info("Config is loaded", "config", cfg)

	postgresUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Db,
	)

	postgresPool, err := pgxpool.New(context.Background(), postgresUrl)
	if err != nil {
		logger.Error("Failed to connect to the database", "err", err)
		return
	}
	defer postgresPool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})
	if res := redisClient.Ping(context.Background()); res.Err() != nil {
		logger.Error("Failed to connect the redis", "err", err)
		return
	}
	defer redisClient.Close()

	userRepository := postgres.NewUserRepository(postgresPool)
	sessionRepository := redisRepo.NewSessionRepository(redisClient)

	userService := services.NewUserService(
		logger,
		userRepository,
		sessionRepository,
		cfg.Domain.SessionDuration,
		cfg.Domain.AccessTokenDuration,
		[]byte(cfg.Domain.AcessTokenSecret),
	)

	m := handlers.SetupHandlers(userService)

	server := &http.Server{
		Handler: m,
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("ListenAndServe failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown of HTTP Server failed", "err", err)
		os.Exit(1)
	}

	logger.Info("Server is gracefully stopped")
}
