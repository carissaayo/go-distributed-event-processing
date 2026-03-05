package processing

import (
	"context"
	"fmt"
	"sync"

	"github.com/carissaayo/go-event-distributed/internal/event"
)

type WorkerPool struct {
	eventCh     chan *event.Event
	router      *Router
	workerCount int
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount, bufferSize int, router *Router) *WorkerPool {
	return &WorkerPool{
		eventCh:     make(chan *event.Event, bufferSize),
		router:      router,
		workerCount: workerCount,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
	fmt.Printf("Started %d workers\n", wp.workerCount)
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	fmt.Printf("Worker %d started\n", id)

	for {
		select {
		case evt, ok := <-wp.eventCh:
			if !ok {
				fmt.Printf("Worker %d stopping: channel closed\n", id)
				return
			}
			if err := wp.router.Route(ctx, evt); err != nil {
				fmt.Printf("Worker %d error processing event %s: %v\n", id, evt.ID, err)
			}
		case <-ctx.Done():
			fmt.Printf("Worker %d stopping: context cancelled\n", id)
			return
		}
	}
}

func (wp *WorkerPool) Submit(evt *event.Event) bool {
	select {
	case wp.eventCh <- evt:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) Shutdown() {
	fmt.Println("Shutting down worker pool...")
	close(wp.eventCh)
	wp.wg.Wait()
	fmt.Println("All workers stopped")
}
