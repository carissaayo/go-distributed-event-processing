package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streamforge_events_received_total",
		Help: "Total number of events received by the API",
	})

	EventsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streamforge_events_processed_total",
		Help: "Total number of events successfully processed",
	})

	EventsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streamforge_events_failed_total",
		Help: "Total number of events that failed processing",
	})

	EventsDropped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "streamforge_events_dropped_total",
		Help: "Total number of events dropped due to full buffer",
	})

	ProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "streamforge_processing_duration_seconds",
		Help:    "Time spent processing each event",
		Buckets: prometheus.DefBuckets,
	})

	BatchWriteDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "streamforge_batch_write_duration_seconds",
		Help:    "Time spent writing batches to MongoDB",
		Buckets: prometheus.DefBuckets,
	})

	BatchSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "streamforge_batch_size",
		Help:    "Number of events per batch write",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
	})

	DLQSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "streamforge_dlq_size",
		Help: "Current number of events in the dead letter queue",
	})

	ChannelBufferUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "streamforge_channel_buffer_usage",
		Help: "Current number of events waiting in the channel buffer",
	})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "streamforge_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})
)
