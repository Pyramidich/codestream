package main

import (
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/config"
	"github.com/ilya/codestream/internal/logger"
	"github.com/ilya/codestream/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Exit(1)
	}

	log := logger.New(cfg.AppEnv, cfg.LogLevel)
	log.Info("starting codestream", slog.String("env", cfg.AppEnv))

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Error("failed to parse redis url", slog.String("error", err.Error()))
		os.Exit(1)
	}
	redisClient := redis.NewClient(redisOpts)

	db, err := initDB(cfg, log)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := server.New(cfg, log, redisClient, db)

	if err := srv.Run(); err != nil {
		log.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func initDB(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.NewGORMLogger(log),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
