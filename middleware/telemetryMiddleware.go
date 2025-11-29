package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/metadata"
)

func TelemetryMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tracer := otel.Tracer(serviceName)
			ctx, span := tracer.Start(ctx, r.URL.Path)
			defer span.End()

			ctx = metadata.AppendToOutgoingContext(ctx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
