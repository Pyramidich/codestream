package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration from environment variables.
type Config struct {
	AppEnv       string        `envconfig:"APP_ENV" default:"dev"`
	HTTPPort     string        `envconfig:"HTTP_PORT" default:"8080"`
	DatabaseURL  string        `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret    string        `envconfig:"JWT_SECRET" required:"true"`
	JWTAccessTTL time.Duration `envconfig:"JWT_ACCESS_TTL" default:"15m"`
	JWTRefreshTTL time.Duration `envconfig:"JWT_REFRESH_TTL" default:"168h"`
	LogLevel     string        `envconfig:"LOG_LEVEL" default:"info"`
	AllowOrigins string        `envconfig:"ALLOW_ORIGINS" default:"http://localhost:5173"`
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// IsDev returns true if the application is running in development mode.
func (c *Config) IsDev() bool {
	return c.AppEnv == "dev"
}

func (c *Config) validate() error {
	if c.AppEnv != "dev" && c.AppEnv != "prod" {
		return errors.New("APP_ENV must be either 'dev' or 'prod'")
	}

	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}

	if c.HTTPPort == "" {
		c.HTTPPort = "8080"
	}

	if c.JWTAccessTTL <= 0 {
		c.JWTAccessTTL = 15 * time.Minute
	}

	if c.JWTRefreshTTL <= 0 {
		c.JWTRefreshTTL = 168 * time.Hour
	}

	return nil
}

// Getenv is a small helper to make config loading testable.
var Getenv = os.Getenv
