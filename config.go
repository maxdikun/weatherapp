package main

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Postgres struct {
		Host     string `env:"HOST"`
		Port     int    `env:"PORT"`
		User     string `env:"USER"`
		Password string `env:"PASSWORD"`
		Db       string `env:"DB"`
	} `envPrefix:"POSTGRES_"`

	Redis struct {
		Addr     string `env:"ADDR"`
		Password string `env:"PASSWORD"`
	} `envPrefix:"REDIS_"`

	Domain struct {
		SessionDuration     time.Duration `env:"SESSION_DURATION"`
		AccessTokenDuration time.Duration `env:"ACCESS_TOKEN_DURATION"`
		AcessTokenSecret    string        `env:"ACCESS_TOKEN_SECRET"`
	} `envPrefix:"DOMAIN_"`

	HTTP struct {
		Port int `env:"PORT"`
	} `envPrefix:"HTTP_"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
