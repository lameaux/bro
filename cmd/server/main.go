package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/lameaux/bro/internal/server/grpc_server"
	"github.com/lameaux/bro/internal/server/metrics"
	"github.com/lameaux/bro/internal/shared/banner"
	"github.com/lameaux/bro/internal/shared/signals"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
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

	grpcServer, err := grpc_server.StartGrpcServer(*port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start grpc server")
		return
	}

	var metricsServer = metrics.StartServer(*metricsPort)

	signals.Handle(true, func() {
		metrics.StopServer(ctx, metricsServer)
		grpc_server.StopGrpcServer(grpcServer)

		cancel()
	})
}
