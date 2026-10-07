package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/config"
	"github.com/ilya/codestream/internal/handler"
	"github.com/ilya/codestream/internal/middleware"
	"github.com/ilya/codestream/internal/repository"
	"github.com/ilya/codestream/internal/service"
)

// Server holds the HTTP server and its dependencies.
type Server struct {
	router *gin.Engine
	config *config.Config
	logger *slog.Logger
	redis  *redis.Client
}

// New creates a new HTTP server with routes and middleware.
func New(cfg *config.Config, logger *slog.Logger, redisClient *redis.Client, db *gorm.DB) *Server {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	s := &Server{
		router: router,
		config: cfg,
		logger: logger,
		redis:  redisClient,
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authHandler := handler.NewAuthHandler(authService)
	authMW := middleware.AuthMiddleware(cfg.JWTSecret)

	s.setupMiddleware()
	s.setupRoutes(authHandler, authMW)

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(s.loggingMiddleware())
}

func (s *Server) setupRoutes(authHandler *handler.AuthHandler, authMW gin.HandlerFunc) {
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/ready", s.handleReady)

	authGroup := s.router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
	}

	s.router.GET("/me", authMW, s.handleMe)
}

// Engine exposes the gin engine for testing.
func (s *Server) Engine() *gin.Engine {
	return s.router
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.HTTPPort)
	s.logger.Info("starting http server", slog.String("addr", addr))
	return s.router.Run(addr)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if s.redis != nil {
		if err := s.redis.Ping(ctx).Err(); err != nil {
			s.logger.Warn("redis readiness check failed", slog.String("error", err.Error()))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "redis unavailable"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		s.logger.Info(
			"request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
		)
	}
}
