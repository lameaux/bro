package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"scenario", "method", "url"},
	)
	HttpRequestsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_failed_total",
			Help: "Number of failed HTTP requests",
		},
		[]string{"scenario", "method", "url", "reason"},
	)
	HttpResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Number of HTTP responses",
		},
		[]string{"scenario", "method", "url", "code", "success"},
	)
	HttpRequestDurationSec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"scenario", "method", "url", "code", "success"},
	)
)

func init() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestsFailedTotal)
	prometheus.MustRegister(HttpResponsesTotal)
	prometheus.MustRegister(HttpRequestDurationSec)
}

func StartServer(port string) *http.Server {
	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr: ":" + port,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).
				Str("port", port).
				Msg("failed to start metrics server")
		}
	}()

	log.Debug().Str("port", port).Msg("metrics server started")

	return server
}

func StopServer(ctx context.Context, server *http.Server) {
	timedOutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(timedOutCtx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown metric server")
	}

	log.Debug().Msg("metrics server stopped")
}
