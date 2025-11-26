package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var HTTPRequestDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Duration of HTTP requests in seconds.",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "path", "status"})

var HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "Total number of HTTP requests.",
}, []string{"method", "path", "status"})

var InFlightRequests = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "http_in_flight_requests",
	Help: "Current number of in-flight HTTP requests.",
})

// Database Metrics
var DBConnectionsOpen = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "db_connections_open",
	Help: "Number of open database connections.",
}, []string{"db_name"})

var DBConnectionsInUse = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "db_connections_in_use",
	Help: "Number of in-use database connections.",
}, []string{"db_name"})

var DBConnectionsIdle = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "db_connections_idle",
	Help: "Number of idle database connections.",
}, []string{"db_name"})

var DBQueryDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "db_query_duration_seconds",
	Help:    "Duration of database queries in seconds.",
	Buckets: prometheus.DefBuckets,
}, []string{"query_type", "repository", "status"})

var DBQueryErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "db_query_errors_total",
	Help: "Total number of failed database queries.",
}, []string{"query_type", "repository"})

func RegisterMetrics() {
	prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.MustRegister(prometheus.NewGoCollector())
}
