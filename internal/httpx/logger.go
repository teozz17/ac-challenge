package httpx

import (
	"log/slog"
	"net/http"
)

type statusAwareResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusAwareResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Logger() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			saw := &statusAwareResponseWriter{ResponseWriter: w}

			defer func() {
				// Skip logging for metrics endpoint to avoid noise
				// Clear console and all data is on Grafana
				if r.URL.Path == "/metrics" {
					return
				}

				if saw.status/100 == 5 {
					slog.ErrorContext(r.Context(), "HTTP request failed", "http_method", r.Method, "http_path", r.URL.Path, "http_status", saw.status)
				} else {
					slog.InfoContext(r.Context(), "HTTP request complete", "http_method", r.Method, "http_path", r.URL.Path, "http_status", saw.status)
				}
			}()

			handler.ServeHTTP(saw, r)
		})
	}
}
