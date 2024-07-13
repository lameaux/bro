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
	HttpRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
	)
	HttpRequestsFailedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_failed_total",
			Help: "Number of failed HTTP requests",
		},
	)
	HttpRequestsTimedOutTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_timed_out_total",
			Help: "Number of timed out HTTP requests",
		},
	)
)

func init() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestsFailedTotal)
	prometheus.MustRegister(HttpRequestsTimedOutTotal)
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
