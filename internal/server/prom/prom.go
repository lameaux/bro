package prom

import (
	"github.com/prometheus/client_golang/prometheus"
)

//nolint:gochecknoglobals
var labels = []string{
	"instance_id",
	"group_id",

	"scenario",
	"method",
	"url",
	"code",

	"failed",
	"timeout",
	"success",
}

type Metrics struct {
	httpRequestsTotal          *prometheus.CounterVec
	httpRequestDurationSeconds *prometheus.HistogramVec
}

func NewMetrics(prefix string) *Metrics {
	requestMetrics := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prefix + "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			labels,
		),
		httpRequestDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    prefix + "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			labels,
		),
	}

	prometheus.MustRegister(
		requestMetrics.httpRequestsTotal,
		requestMetrics.httpRequestDurationSeconds,
	)

	return requestMetrics
}

func (m *Metrics) CountRequest(labels prometheus.Labels, latency float64) {
	m.httpRequestsTotal.With(labels).Inc()
	m.httpRequestDurationSeconds.With(labels).Observe(latency)
}
