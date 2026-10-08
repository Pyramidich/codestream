package testutil

import (
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/config"
	"github.com/ilya/codestream/internal/logger"
	"github.com/ilya/codestream/internal/server"
)

// NewTestServer creates a real HTTP server for integration tests.
func NewTestServer(t *testing.T, db *gorm.DB) *httptest.Server {
	t.Helper()

	cfg := &config.Config{
		AppEnv:       "dev",
		HTTPPort:     "0",
		DatabaseURL:  "",
		JWTSecret:    "test-secret",
		JWTAccessTTL: 15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		LogLevel:     "info",
	}

	log := logger.New("dev", "info")
	srv := server.New(cfg, log, db)

	return httptest.NewServer(srv.Handler())
}
