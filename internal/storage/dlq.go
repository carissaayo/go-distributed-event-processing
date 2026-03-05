package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/event"
)

type FailedEvent struct {
	Event    *event.Event
	Error    string
	FailedAt time.Time
	Retries  int
}

type DLQ struct {
	events []FailedEvent
	mu     sync.Mutex
}

func NewDLQ() *DLQ {
	return &DLQ{
		events: make([]FailedEvent, 0),
	}
}

func (d *DLQ) Add(evt *event.Event, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.events = append(d.events, FailedEvent{
		Event:    evt,
		Error:    err.Error(),
		FailedAt: time.Now(),
		Retries:  0,
	})
	fmt.Printf("Event %s added to DLQ: %v\n", evt.ID, err)
}

func (d *DLQ) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.events)
}

func (d *DLQ) Drain() []FailedEvent {
	d.mu.Lock()
	defer d.mu.Unlock()

	events := d.events
	d.events = make([]FailedEvent, 0)
	return events
}
