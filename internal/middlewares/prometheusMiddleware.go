package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"markly/internal/utils"
)

// responseWriterWrapper wraps http.ResponseWriter to capture the status code.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{w, http.StatusOK} // Default status to 200
}

func (lrw *responseWriterWrapper) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// PrometheusMiddleware is a middleware that records HTTP request metrics.
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the in-flight requests gauge
		utils.InFlightRequests.Inc()
		defer utils.InFlightRequests.Dec()

		// Start timer for request duration
		start := time.Now()
		
		// Wrap the response writer to capture the status code
		wrappedWriter := newResponseWriterWrapper(w)

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(wrappedWriter.statusCode)

		utils.HTTPRequestDurationSeconds.WithLabelValues(r.Method, r.URL.Path, statusCode).Observe(duration)
		utils.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode).Inc()
	})
}
