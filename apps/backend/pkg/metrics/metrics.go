package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metric collectors.
type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	AIRequestsTotal     *prometheus.CounterVec
	AITokensUsed        *prometheus.CounterVec
	QueueJobsTotal      *prometheus.CounterVec
	ActiveConnections   prometheus.Gauge
}

// New registers and returns all application metrics.
func New() *Metrics {
	m := &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "path", "status_code"}),

		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),

		AIRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ai_requests_total",
			Help: "Total number of AI provider requests.",
		}, []string{"provider", "operation", "status"}),

		AITokensUsed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ai_tokens_used_total",
			Help: "Total AI tokens consumed.",
		}, []string{"provider", "model"}),

		QueueJobsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "queue_jobs_total",
			Help: "Total number of queue jobs processed.",
		}, []string{"job_type", "status"}),

		ActiveConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Current number of active HTTP connections.",
		}),
	}

	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.AIRequestsTotal,
		m.AITokensUsed,
		m.QueueJobsTotal,
		m.ActiveConnections,
	)

	return m
}

// Handler returns the Prometheus HTTP handler.
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// RecordHTTP records an HTTP request metric.
func (m *Metrics) RecordHTTP(method, path, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordAI records an AI provider request metric.
func (m *Metrics) RecordAI(provider, operation, status, model string, tokens float64) {
	m.AIRequestsTotal.WithLabelValues(provider, operation, status).Inc()
	if tokens > 0 {
		m.AITokensUsed.WithLabelValues(provider, model).Add(tokens)
	}
}

// RecordQueueJob records a queue job metric.
func (m *Metrics) RecordQueueJob(jobType, status string) {
	m.QueueJobsTotal.WithLabelValues(jobType, status).Inc()
}
