package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/config"
	"github.com/carissaayo/go-event-distributed/internal/logger"
	"github.com/carissaayo/go-event-distributed/internal/server"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	defer logger.Sync()

	srv, err := server.New(cfg)
	if err != nil {
		logger.Log.Fatal("failed to create server", zap.Error(err))
	}

	go func() {
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			logger.Log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("shutdown error", zap.Error(err))
	}
	logger.Log.Info("server stopped gracefully")
}
