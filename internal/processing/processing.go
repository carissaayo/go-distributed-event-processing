package processing

import (
	"context"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/event"
	"github.com/carissaayo/go-event-distributed/internal/logger"
	"github.com/carissaayo/go-event-distributed/internal/metrics"
	"github.com/carissaayo/go-event-distributed/internal/storage"
	"go.uber.org/zap"
)

type Processor interface {
	Process(ctx context.Context, evt *event.Event) error
}

type LogProcessor struct{}

func (p *LogProcessor) Process(ctx context.Context, evt *event.Event) error {
	logger.Log.Info("processed event",
		zap.String("event_id", evt.ID),
		zap.String("type", evt.Type),
	)
	evt.Processed = true
	return nil
}

type BatchProcessor struct {
	batcher *storage.Batcher
}

func NewBatchProcessor(batcher *storage.Batcher) *BatchProcessor {
	return &BatchProcessor{batcher: batcher}
}

func (p *BatchProcessor) Process(ctx context.Context, evt *event.Event) error {
	start := time.Now()

	evt.Processed = true
	p.batcher.Add(evt)

	metrics.EventsProcessed.Inc()
	metrics.ProcessingDuration.Observe(time.Since(start).Seconds())

	return nil
}
