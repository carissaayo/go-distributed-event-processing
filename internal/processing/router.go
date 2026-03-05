package processing

import (
	"context"

	"github.com/carissaayo/go-event-distributed/internal/event"
)

type Router struct {
	routes           map[string]Processor
	defaultProcessor Processor
}

func NewRouter(defaultProcessor Processor) *Router {
	return &Router{
		routes:           make(map[string]Processor),
		defaultProcessor: defaultProcessor,
	}
}

func (r *Router) Register(eventType string, p Processor) {
	r.routes[eventType] = p
}

func (r *Router) Route(ctx context.Context, evt *event.Event) error {
	p, ok := r.routes[evt.Type]
	if !ok {
		p = r.defaultProcessor
	}
	return p.Process(ctx, evt)
}
