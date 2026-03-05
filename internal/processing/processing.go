package processing

import (
	"context"
	"fmt"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/event"
	"github.com/carissaayo/go-event-distributed/internal/storage"
)

type Processor interface {
	Process(ctx context.Context, evt *event.Event) error
}

type LogProcessor struct{}

func (p *LogProcessor) Process(ctx context.Context, evt *event.Event) error {
	fmt.Printf("[%s] Processed event: type=%s id=%s\n",
		time.Now().Format("15:04:05"), evt.Type, evt.ID)
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
	evt.Processed = true
	p.batcher.Add(evt)
	return nil
}
