package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/config"
	"github.com/ilya/codestream/internal/logger"
	"github.com/ilya/codestream/internal/server"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{AppEnv: "dev", HTTPPort: "8080", JWTAccessTTL: 15 * time.Minute, JWTRefreshTTL: 7 * 24 * time.Hour}
	log := logger.New("dev", "info")

	srv := server.New(cfg, log, nil)
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.NoError(t, err)

	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}
