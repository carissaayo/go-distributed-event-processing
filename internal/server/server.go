package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carissaayo/go-event-distributed/internal/api"
	"github.com/carissaayo/go-event-distributed/internal/config"
	"github.com/carissaayo/go-event-distributed/internal/logger"
	"github.com/carissaayo/go-event-distributed/internal/processing"
	"github.com/carissaayo/go-event-distributed/internal/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	store      *storage.MongoDBStore
	batcher    *storage.Batcher
	workerPool *processing.WorkerPool
	cfg        *config.Config
}

func New(cfg *config.Config) (*Server, error) {
	ctx := context.Background()

	store, err := storage.NewMongoDBStore(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongodb store: %w", err)
	}

	dlq := storage.NewDLQ()

	batcher := storage.NewBatcher(
		store,
		dlq,
		cfg.Processing.BatchSize,
		cfg.Processing.FlushInterval,
	)
	batcher.Start(ctx)

	batchProcessor := processing.NewBatchProcessor(batcher)
	router := processing.NewRouter(batchProcessor)

	workerPool := processing.NewWorkerPool(
		cfg.Processing.WorkerCount,
		cfg.Processing.BufferSize,
		router,
	)
	workerPool.Start(ctx)

	handlers := api.NewHandlers(workerPool, store)

	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(api.RequestLogger)

	r.Get("/health", handlers.HealthCheck)
	r.Get("/ready", handlers.ReadyCheck)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/events", handlers.CreateEvent)
		r.Get("/events/{id}", handlers.GetEvent)
	})

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      r,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		store:      store,
		batcher:    batcher,
		workerPool: workerPool,
		cfg:        cfg,
	}, nil
}

func (s *Server) Start() error {
	logger.Log.Info("server starting", zap.Int("port", s.cfg.Server.Port))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Log.Info("server shutting down...")
	s.workerPool.Shutdown()
	s.batcher.Shutdown()
	s.store.Close(ctx)
	return s.httpServer.Shutdown(ctx)
}
