package main

import (
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

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

	srv := server.New(cfg, log, redisClient)

	if err := srv.Run(); err != nil {
		log.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
