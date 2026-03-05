package processing

import (
	"context"
	"fmt"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/event"
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
