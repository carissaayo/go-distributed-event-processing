# StreamForge

> A distributed event processing pipeline built with Go

StreamForge receives events via HTTP, queues them in a buffered channel, and processes them concurrently with a pool of worker goroutines. Processed events are batch-written to MongoDB. The system includes structured logging, Prometheus metrics, health checks, and graceful shutdown.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![MongoDB](https://img.shields.io/badge/MongoDB-7.0+-47A248?style=flat&logo=mongodb&logoColor=white)](https://mongodb.com)

---

## Architecture

```
Client
  │
  ▼
HTTP API (Chi Router)
  │
  ▼
Event Validation & Enrichment
  │
  ▼
Buffered Channel (10,000 capacity)
  │
  ▼
Event Router (routes by event type)
  │
  ├──── Worker 0 ──┐
  ├──── Worker 1 ──┤
  ├──── Worker 2 ──┤
  ├──── ...        ├──→ Batch Writer (flushes every 100 events or 1s)
  ├──── Worker 6 ──┤              │
  └──── Worker 7 ──┘              ▼
                              MongoDB
                                │
                          (on failure)
                                ▼
                        Dead Letter Queue
```

**How it works:**

1. HTTP handler validates and enriches the incoming event
2. Event is submitted to a buffered channel (non-blocking; returns 503 if full)
3. One of 8 worker goroutines picks up the event
4. The router directs it to the appropriate processor based on event type
5. The processor adds the event to a batch buffer
6. The batcher flushes to MongoDB when the buffer hits 100 events or every 1 second
7. Failed writes go to an in-memory dead letter queue

---

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness check (always 200) |
| `GET` | `/ready` | Readiness check (pings MongoDB) |
| `GET` | `/metrics` | Prometheus metrics |
| `POST` | `/api/v1/events` | Create an event |
| `GET` | `/api/v1/events/{id}` | Retrieve an event by ID |

### Create Event

```
POST /api/v1/events
Content-Type: application/json

{
  "type": "user.signup",
  "data": {
    "user_id": "123",
    "email": "user@example.com"
  }
}

→ 201 Created
{
  "event_id": "evt_abc123def4",
  "status": "accepted"
}
```

### Get Event

```
GET /api/v1/events/evt_abc123def4

→ 200 OK
{
  "event_id": "evt_abc123def4",
  "type": "user.signup",
  "data": {
    "user_id": "123",
    "email": "user@example.com",
    "_enriched_at": "2026-03-04T14:30:00Z"
  },
  "timestamp": "2026-03-04T14:30:00Z",
  "processed": true
}
```

---

## Prometheus Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `streamforge_events_received_total` | Counter | Events received by the API |
| `streamforge_events_processed_total` | Counter | Events successfully processed |
| `streamforge_events_failed_total` | Counter | Events that failed processing |
| `streamforge_events_dropped_total` | Counter | Events dropped (buffer full) |
| `streamforge_processing_duration_seconds` | Histogram | Per-event processing time |
| `streamforge_batch_write_duration_seconds` | Histogram | MongoDB bulk write time |
| `streamforge_batch_size` | Histogram | Events per batch |
| `streamforge_channel_buffer_usage` | Gauge | Events waiting in the channel |
| `streamforge_dlq_size` | Gauge | Events in the dead letter queue |
| `streamforge_http_request_duration_seconds` | Histogram | HTTP request duration by method/path/status |

---

## Project Structure

```
go-event-distributed/
├── cmd/
│   └── server/
│       └── main.go                # Entry point, signal handling, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go              # Environment-based configuration
│   ├── server/
│   │   └── server.go              # HTTP server, routing, component wiring
│   ├── api/
│   │   ├── handlers.go            # HTTP handlers (create, get, health, ready)
│   │   └── middleware.go          # Request logging, metrics middleware
│   ├── event/
│   │   ├── event.go               # Event struct and request/response types
│   │   ├── validator.go           # Input validation
│   │   └── enricher.go            # Timestamp and metadata enrichment
│   ├── processing/
│   │   ├── processing.go          # Processor interface, LogProcessor, BatchProcessor
│   │   ├── worker.go              # Worker pool (goroutines + channel + WaitGroup)
│   │   └── router.go              # Routes events to processors by type
│   ├── storage/
│   │   ├── mongodb.go             # MongoDB client (connect, insert, find, ping)
│   │   ├── batcher.go             # Batch writer (buffer, flush loop, mutex)
│   │   └── dlq.go                 # Dead letter queue for failed writes
│   ├── metrics/
│   │   └── metrics.go             # Prometheus counters, histograms, gauges
│   └── logger/
│       └── logger.go              # Structured JSON logging (Zap)
├── .env                           # Environment variables
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## Configuration

All configuration is via environment variables (loaded from `.env` with godotenv):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `ENV` | `development` | Environment name |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DATABASE` | `events` | MongoDB database name |
| `WORKER_COUNT` | `8` | Number of worker goroutines |
| `BUFFER_SIZE` | `10000` | Event channel buffer capacity |
| `BATCH_SIZE` | `100` | Events per batch write |
| `FLUSH_INTERVAL` | `1s` | Max time between flushes |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json or console) |

---

## Running Locally

**Prerequisites:** Go 1.21+, MongoDB running locally

```bash
git clone https://github.com/carissaayo/go-event-distributed.git
cd go-event-distributed
```

Create a `.env` file:

```bash
PORT=8080
MONGO_URI=mongodb://localhost:27017
WORKER_COUNT=8
BUFFER_SIZE=10000
BATCH_SIZE=100
FLUSH_INTERVAL=1s
LOG_LEVEL=info
LOG_FORMAT=json
```

Run:

```bash
go mod tidy
go run cmd/server/main.go
```

---

## Key Go Patterns Used

| Pattern | Where | Purpose |
|---------|-------|---------|
| Worker pool | `processing/worker.go` | Fixed number of goroutines reading from a shared channel |
| Buffered channel | `processing/worker.go` | Decouples HTTP handlers from event processing |
| `select` statement | `worker.go`, `batcher.go` | Multiplexes channel operations (events, timers, shutdown) |
| Non-blocking send | `worker.go` Submit() | Back-pressure: returns false instead of blocking when buffer is full |
| `sync.Mutex` | `batcher.go`, `dlq.go` | Protects shared state accessed by multiple goroutines |
| `sync.WaitGroup` | `worker.go` | Waits for all workers to finish during shutdown |
| `context.WithCancel` | `server.go` | Propagates shutdown signal to workers and batcher |
| Interface-based dispatch | `processing.go`, `router.go` | `Processor` interface allows swappable processing strategies |
| Graceful shutdown | `main.go`, `server.go` | Signal capture, worker drain, final flush, DB disconnect |
| Batch + flush timer | `batcher.go` | Amortizes MongoDB writes; flushes on size threshold or time interval |

---

## Tech Stack

- **Go** -- HTTP server, concurrency (goroutines, channels, select)
- **Chi** -- Lightweight HTTP router and middleware
- **MongoDB** -- Event persistence (bulk writes via official Go driver)
- **Prometheus** -- Metrics collection and exposition
- **Zap** -- Structured JSON logging
- **godotenv** -- Environment variable loading from `.env`

---

## License

MIT
