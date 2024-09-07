package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/lameaux/bro/internal/banner"
	"github.com/lameaux/bro/internal/metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

const (
	appName    = "brod"
	appVersion = "v0.0.1"
)

var GitHash string

func main() {
	var debug = flag.Bool("debug", false, "enable debug mode")
	var logJson = flag.Bool("logJson", false, "log as json")
	var skipBanner = flag.Bool("skipBanner", false, "skip banner")
	var port = flag.Int("port", 8080, "port for grpc server")
	var metricsPort = flag.Int("metricsPort", 9090, "port for metrics server")

	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if !*logJson {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if !*skipBanner {
		fmt.Print(banner.Banner)
	}

	log.Info().Str("version", appVersion).Str("build", GitHash).Msg(appName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcServer, err := startGrpcServer(*port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start grpc server")
		return
	}

	var metricsServer = metrics.StartServer(*metricsPort)

	handleSignals(func() {
		metrics.StopServer(ctx, metricsServer)
		stopGrpcServer(grpcServer)

		cancel()
	})
}

func handleSignals(shutdownFn func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("received signal")
	shutdownFn()
}
