package api

import (
	"encoding/json"
	"net/http"

	"github.com/carissaayo/go-event-distributed/internal/event"
	"github.com/carissaayo/go-event-distributed/internal/logger"
	"github.com/carissaayo/go-event-distributed/internal/metrics"
	"github.com/carissaayo/go-event-distributed/internal/processing"
	"github.com/carissaayo/go-event-distributed/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handlers struct {
	workerPool *processing.WorkerPool
	store      *storage.MongoDBStore
}

func NewHandlers(
	wp *processing.WorkerPool,
	store *storage.MongoDBStore,
) *Handlers {
	return &Handlers{
		workerPool: wp,
		store:      store,
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		logger.Log.Error("readiness check failed", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req event.CreateEventRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Warn("invalid json received", zap.Error(err))
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if err := event.Validate(&req); err != nil {
		logger.Log.Warn("event validation failed", zap.Error(err))
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	evt := event.NewEvent(req.Type, req.Data)
	evt = event.Enrich(evt)

	metrics.EventsReceived.Inc()

	if !h.workerPool.Submit(evt) {
		metrics.EventsDropped.Inc()
		logger.Log.Warn("event dropped, buffer full", zap.String("event_id", evt.ID))
		http.Error(w, `{"error":"server busy, try again later"}`, http.StatusServiceUnavailable)
		return
	}

	logger.Log.Info("event accepted",
		zap.String("event_id", evt.ID),
		zap.String("type", evt.Type),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(event.CreateEventResponse{
		EventID: evt.ID,
		Status:  "accepted",
	})
}

func (h *Handlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, `{"error":"event id required"}`, http.StatusBadRequest)
		return
	}

	evt, err := h.store.FindByID(r.Context(), id)
	if err != nil {
		logger.Log.Error("failed to find event", zap.String("event_id", id), zap.Error(err))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	if evt == nil {
		http.Error(w, `{"error":"event not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(evt)
}
