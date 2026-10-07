package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/config"
)

func TestConfigDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_SECRET", "secret")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "dev", cfg.AppEnv)
	assert.Equal(t, "8080", cfg.HTTPPort)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfigMissingRequired(t *testing.T) {
	// Clear potentially set env vars.
	keys := []string{"DATABASE_URL", "REDIS_URL", "JWT_SECRET", "APP_ENV", "HTTP_PORT", "LOG_LEVEL"}
	for _, key := range keys {
		os.Unsetenv(key)
	}

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}
