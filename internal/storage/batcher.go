package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/event"
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
	fmt.Println("Batcher started")
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

	if err := b.store.InsertMany(ctx, events); err != nil {
		fmt.Printf("Batch write failed (%d events): %v\n", len(events), err)
		for _, evt := range events {
			b.dlq.Add(evt, err)
		}
		return
	}
	fmt.Printf("Batch wrote %d events\n", len(events))
}

func (b *Batcher) Shutdown() {
	fmt.Println("Batcher shutting down...")
	<-b.done
	fmt.Println("Batcher stopped")
}
