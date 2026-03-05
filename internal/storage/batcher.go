package storage

import (
	"context"
	"sync"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/event"
	"github.com/carissaayo/go-event-distributed/internal/logger"
	"github.com/carissaayo/go-event-distributed/internal/metrics"
	"go.uber.org/zap"
)

type Batcher struct {
	store         *MongoDBStore
	dlq           *DLQ
	buffer        []*event.Event
	mu            sync.Mutex
	batchSize     int
	flushInterval time.Duration
	flushCh       chan struct{}
	done          chan struct{}
}

func NewBatcher(store *MongoDBStore, dlq *DLQ, batchSize int, flushInterval time.Duration) *Batcher {
	b := &Batcher{
		store:         store,
		dlq:           dlq,
		buffer:        make([]*event.Event, 0, batchSize),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		flushCh:       make(chan struct{}, 1),
		done:          make(chan struct{}),
	}
	return b
}

func (b *Batcher) Start(ctx context.Context) {
	go b.flushLoop(ctx)
	logger.Log.Info("batcher started",
		zap.Int("batch_size", b.batchSize),
		zap.Duration("flush_interval", b.flushInterval),
	)
}

func (b *Batcher) Add(evt *event.Event) {
	b.mu.Lock()
	b.buffer = append(b.buffer, evt)
	shouldFlush := len(b.buffer) >= b.batchSize
	b.mu.Unlock()

	if shouldFlush {
		select {
		case b.flushCh <- struct{}{}:
		default:
		}
	}
}

func (b *Batcher) flushLoop(ctx context.Context) {
	defer close(b.done)
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flush(ctx)
		case <-b.flushCh:
			b.flush(ctx)
		case <-ctx.Done():
			b.flush(context.Background())
			return
		}
	}
}

func (b *Batcher) flush(ctx context.Context) {
	b.mu.Lock()
	if len(b.buffer) == 0 {
		b.mu.Unlock()
		return
	}
	events := b.buffer
	b.buffer = make([]*event.Event, 0, b.batchSize)
	b.mu.Unlock()

	start := time.Now()
	metrics.BatchSize.Observe(float64(len(events)))

	if err := b.store.InsertMany(ctx, events); err != nil {
		logger.Log.Error("batch write failed",
			zap.Int("count", len(events)),
			zap.Error(err),
		)
		for _, evt := range events {
			b.dlq.Add(evt, err)
		}
		metrics.DLQSize.Set(float64(b.dlq.Len()))
		return
	}

	metrics.BatchWriteDuration.Observe(time.Since(start).Seconds())
	logger.Log.Info("batch wrote events", zap.Int("count", len(events)))
}

func (b *Batcher) Shutdown() {
	logger.Log.Info("batcher shutting down...")
	<-b.done
	logger.Log.Info("batcher stopped")
}
