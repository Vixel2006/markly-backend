package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMiddleware is a middleware that records Prometheus metrics for HTTP requests.
type PrometheusMiddleware struct {
	totalRequests   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
}

// NewPrometheusMiddleware creates and returns a new PrometheusMiddleware instance.
func NewPrometheusMiddleware() *PrometheusMiddleware {
	m := &PrometheusMiddleware{
		totalRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_request_duration_seconds",
				Help: "Duration of HTTP requests in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		responseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_response_size_bytes",
				Help: "Size of HTTP responses in bytes.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
	}
	return m
}

// Instrument is the HTTP middleware function.
func (m *PrometheusMiddleware) Instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer that captures the status code and response size
		lrw := &loggingResponseWriter{ResponseWriter: w}
		
		next.ServeHTTP(lrw, r)

		statusCode := strconv.Itoa(lrw.statusCode)
		path := r.URL.Path
		method := r.Method

		m.totalRequests.WithLabelValues(method, path, statusCode).Inc()
		m.requestDuration.WithLabelValues(method, path, statusCode).Observe(time.Since(start).Seconds())
		m.responseSize.WithLabelValues(method, path, statusCode).Observe(float64(lrw.responseSize))
	})
}

// loggingResponseWriter is a wrapper around http.ResponseWriter that captures the status code and response size.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	responseSize int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	if lrw.statusCode == 0 {
		lrw.statusCode = http.StatusOK // Default status code if WriteHeader is not called
	}
	n, err := lrw.ResponseWriter.Write(data)
	lrw.responseSize += n
	return n, err
}
