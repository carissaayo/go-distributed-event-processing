package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/carissaayo/go-event-distributed/internal/logger"
	"github.com/carissaayo/go-event-distributed/internal/metrics"
	"go.uber.org/zap"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		logger.Log.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", ww.status),
			zap.Duration("duration", duration),
		)

		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", ww.status),
		).Observe(duration.Seconds())
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
