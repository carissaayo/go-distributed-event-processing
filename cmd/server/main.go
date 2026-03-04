package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/config"
	"github.com/carissaayo/go-event-distributed/internal/server"
)

func main() {
	cfg := config.Load()
	srv := server.New(cfg)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown with 10s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server stopped gracefully")
}
