package processing

import (
	"context"
	"sync"

	"github.com/carissaayo/go-event-distributed/internal/event"
	"github.com/carissaayo/go-event-distributed/internal/logger"
	"github.com/carissaayo/go-event-distributed/internal/metrics"
	"go.uber.org/zap"
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
	logger.Log.Info("worker pool started", zap.Int("workers", wp.workerCount))
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	logger.Log.Info("worker started", zap.Int("worker_id", id))

	for {
		select {
		case evt, ok := <-wp.eventCh:
			if !ok {
				logger.Log.Info("worker stopping: channel closed", zap.Int("worker_id", id))
				return
			}
			metrics.ChannelBufferUsage.Set(float64(len(wp.eventCh)))
			if err := wp.router.Route(ctx, evt); err != nil {
				metrics.EventsFailed.Inc()
				logger.Log.Error("worker failed to process event",
					zap.Int("worker_id", id),
					zap.String("event_id", evt.ID),
					zap.Error(err),
				)
			}
		case <-ctx.Done():
			logger.Log.Info("worker stopping: context cancelled", zap.Int("worker_id", id))
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
	logger.Log.Info("shutting down worker pool...")
	close(wp.eventCh)
	wp.wg.Wait()
	logger.Log.Info("all workers stopped")
}
