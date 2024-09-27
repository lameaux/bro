package restserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lameaux/bro/internal/shared/banner"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	readHeaderTimeout = 5 * time.Second
	stopTimeout       = 5 * time.Second
)

func StartServer(port int) *http.Server {
	http.Handle("/", IndexHandler())

	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).
				Int("port", port).
				Msg("failed to start metrics server")
		}
	}()

	log.Debug().Int("port", port).Msg("metrics server started")

	return server
}

func StopServer(ctx context.Context, server *http.Server) {
	timedOutCtx, cancel := context.WithTimeout(ctx, stopTimeout)
	defer cancel()

	if err := server.Shutdown(timedOutCtx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown metric server")
	}

	log.Debug().Msg("metrics server stopped")
}

func IndexHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)

		_, err := writer.Write([]byte(banner.Banner))
		if err != nil {
			log.Warn().Err(err).Msg("failed to write response")
		}
	})
}
