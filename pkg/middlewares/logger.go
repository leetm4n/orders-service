package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func LoggerMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()

		span := trace.SpanFromContext(r.Context())
		sc := span.SpanContext()

		next.ServeHTTP(w, r)

		durationInMs := time.Since(timeStart).Milliseconds()
		duration := fmt.Sprintf("%dms", durationInMs)

		slog.Info("completed request uri", "uri", r.RequestURI, "method", r.Method, "duration", duration,
			"trace_id", sc.TraceID(), "span_id", sc.SpanID())
	})
}
