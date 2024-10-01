package prom

import (
	"github.com/prometheus/client_golang/prometheus"
)

//nolint:gochecknoglobals
var labelsTotal = []string{
	"group_id",

	"scenario",
	"method",
	"url",
	"code",

	"failed",
	"timeout",
	"success",
}

//nolint:gochecknoglobals
var labelsDuration = []string{
	"group_id",

	"scenario",
	"method",
	"url",
	"code",

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
			labelsTotal,
		),
		httpRequestDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    prefix + "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			labelsDuration,
		),
	}

	prometheus.MustRegister(
		requestMetrics.httpRequestsTotal,
		requestMetrics.httpRequestDurationSeconds,
	)

	return requestMetrics
}

func (m *Metrics) CountRequest(labels map[string]string, latency float64) {
	m.httpRequestsTotal.With(makePromLabels(labels, labelsTotal)).Inc()
	m.httpRequestDurationSeconds.With(makePromLabels(labels, labelsDuration)).Observe(latency)
}

func makePromLabels(labels prometheus.Labels, keys []string) prometheus.Labels {
	result := prometheus.Labels{}

	for _, key := range keys {
		result[key] = labels[key]
	}

	return result
}
