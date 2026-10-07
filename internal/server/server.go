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

	"github.com/ilya/codestream/internal/authz"
	"github.com/ilya/codestream/internal/config"
	"github.com/ilya/codestream/internal/handler"
	"github.com/ilya/codestream/internal/middleware"
	"github.com/ilya/codestream/internal/repository"
	"github.com/ilya/codestream/internal/service"
	"github.com/ilya/codestream/internal/ws"
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

	// Repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	projectMemberRepo := repository.NewProjectMemberRepository(db)
	fileRepo := repository.NewFileRepository(db)

	// Authorization
	authzInstance := authz.NewAuthorization(projectMemberRepo)

	// Services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	changeHistRepo := repository.NewChangeHistoryRepository(db)
	changeHistService := service.NewChangeHistoryService(changeHistRepo, logger)
	projectService := service.NewProjectService(projectRepo, projectMemberRepo, authzInstance, changeHistService)
	memberService := service.NewProjectMemberService(projectMemberRepo, userRepo, authzInstance, changeHistService)
	fileService := service.NewFileService(fileRepo, authzInstance, changeHistService)

	// WebSocket
	hub := ws.NewHub()
	documentVersionRepo := repository.NewDocumentVersionRepository(db)
	documentStateManager := service.NewDocumentStateManager(hub, fileRepo, documentVersionRepo, logger)
	documentStateManager.StartSnapshotWorker(context.Background(), 30*time.Second)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	memberHandler := handler.NewProjectMemberHandler(memberService)
	fileHandler := handler.NewFileHandler(fileService)
	wsHandler := handler.NewWSHandler(hub, fileService, documentStateManager, cfg.JWTSecret, logger)

	authMW := middleware.AuthMiddleware(cfg.JWTSecret)

	s.setupMiddleware()
	s.setupRoutes(authHandler, projectHandler, memberHandler, fileHandler, wsHandler, authMW)

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(s.loggingMiddleware())
}

func (s *Server) setupRoutes(
	authHandler *handler.AuthHandler,
	projectHandler *handler.ProjectHandler,
	memberHandler *handler.ProjectMemberHandler,
	fileHandler *handler.FileHandler,
	wsHandler *handler.WSHandler,
	authMW gin.HandlerFunc,
) {
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/ready", s.handleReady)

	authGroup := s.router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
	}

	s.router.GET("/ws", wsHandler.Handle)

	authorized := s.router.Group("/")
	authorized.Use(authMW)
	{
		authorized.GET("/me", s.handleMe)

		// Projects
		authorized.POST("/projects", projectHandler.Create)
		authorized.GET("/projects", projectHandler.List)
		authorized.GET("/projects/:id", projectHandler.Get)
		authorized.PATCH("/projects/:id", projectHandler.Update)
		authorized.DELETE("/projects/:id", projectHandler.Delete)

		// Project members
		authorized.POST("/projects/:id/members", memberHandler.Add)
		authorized.GET("/projects/:id/members", memberHandler.List)
		authorized.DELETE("/projects/:id/members/:user_id", memberHandler.Remove)

		// Files
		authorized.POST("/projects/:id/files", fileHandler.Create)
		authorized.GET("/projects/:id/files", fileHandler.List)
		authorized.GET("/files/:id", fileHandler.Get)
		authorized.PATCH("/files/:id", fileHandler.Update)
		authorized.DELETE("/files/:id", fileHandler.Delete)
	}
}

// Engine exposes the gin engine for testing.
func (s *Server) Engine() *gin.Engine {
	return s.router
}

// Handler returns the http.Handler for testing with httptest.
func (s *Server) Handler() http.Handler {
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
