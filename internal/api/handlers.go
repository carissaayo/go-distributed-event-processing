package api

import (
	"encoding/json"
	"net/http"

	"github.com/carissaayo/go-event-distributed/internal/event"
	"github.com/carissaayo/go-event-distributed/internal/processing"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	workerPool *processing.WorkerPool
	// store      *storage.MongoDBStore
}

func NewHandlers(
	wp *processing.WorkerPool,
	// store *storage.MongoDBStore
) *Handlers {
	return &Handlers{
		workerPool: wp,
		// store:      store,
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req event.CreateEventRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if err := event.Validate(&req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	evt := event.NewEvent(req.Type, req.Data)
	// evt = event.Enrich(evt)

	// metrics.EventsReceivedTotal.Inc()

	if !h.workerPool.Submit(evt) {
		http.Error(w, `{"error":"server busy, try again later"}`, http.StatusServiceUnavailable)
		return
	}

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

	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
