package restserver

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	//nolint:gochecknoglobals
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"scenario", "method", "url"},
	)

	//nolint:gochecknoglobals
	httpRequestsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_failed_total",
			Help: "Number of failed HTTP requests",
		},
		[]string{"scenario", "method", "url", "reason"},
	)

	//nolint:gochecknoglobals
	httpResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Number of HTTP responses",
		},
		[]string{"scenario", "method", "url", "code", "success"},
	)

	//nolint:gochecknoglobals
	httpRequestDurationSec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"scenario", "method", "url", "code", "success"},
	)
)

func init() { //nolint:gochecknoinits
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestsFailedTotal)
	prometheus.MustRegister(httpResponsesTotal)
	prometheus.MustRegister(httpRequestDurationSec)
}

func CountFailedRequest(labels prometheus.Labels) {
	httpRequestsFailedTotal.With(labels).Inc()
}
