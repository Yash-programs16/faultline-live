package api

import (
	"net/http"
	"time"

	"faultline-live/backend/internal/telemetry"
)

func TelemetryMiddleware(collector *telemetry.Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds() * 1000

			route := telemetry.NormalizeRoute(r.URL.Path)
			service := telemetry.ServiceFromRoute(route)

			success := rw.statusCode < 400

			collector.Record(telemetry.Observation{
				Timestamp:     start,
				RouteTemplate: route,
				Service:       service,
				StatusCode:    rw.statusCode,
				DurationMs:    duration,
				Success:       success,
			})
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}