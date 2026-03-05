package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carissaayo/go-event-distributed/internal/api"
	"github.com/carissaayo/go-event-distributed/internal/config"
	"github.com/carissaayo/go-event-distributed/internal/processing"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	httpServer *http.Server
	// store      *storage.MongoDBStore
	// batcher    *storage.Batcher
	workerPool *processing.WorkerPool
	cfg        *config.Config
}

func New(cfg *config.Config) *Server {
	router := processing.NewRouter(&processing.LogProcessor{})
	workerPool := processing.NewWorkerPool(
		cfg.Processing.WorkerCount,
		cfg.Processing.BufferSize,
		router,
	)
	workerPool.Start(context.Background())
	handlers := api.NewHandlers(workerPool)
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	r.Use(api.RequestLogger)

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
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	fmt.Printf("Server starting on port %d\n", s.cfg.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("Server shutting down...")
	return s.httpServer.Shutdown(ctx)
}
