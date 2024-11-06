package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/lameaux/bro/internal/server/grpcserver"
	"github.com/lameaux/bro/internal/server/prom"
	"github.com/lameaux/bro/internal/server/restserver"
	"github.com/lameaux/bro/internal/shared/banner"
	"github.com/lameaux/bro/internal/shared/signals"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	appName = "brod"

	defaultPortGrpc = 8080
	defaultPortRest = 9090

	defaultMetricsPrefix = "bro_"
)

var (
	Version   string //nolint:gochecknoglobals
	BuildHash string //nolint:gochecknoglobals
	BuildDate string //nolint:gochecknoglobals
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	logJSON := flag.Bool("logJson", false, "log as json")
	skipBanner := flag.Bool("skipBanner", false, "skip banner")
	grpcPort := flag.Int("grpcPort", defaultPortGrpc, "port for grpc server")
	restPort := flag.Int("restPort", defaultPortRest, "port for rest server")
	metricsPrefix := flag.String("metricsPrefix", defaultMetricsPrefix, "prefix for metrics")

	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if !*logJSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if !*skipBanner {
		fmt.Print(banner.Banner) //nolint:forbidigo
	}

	log.Info().Str("version", Version).
		Str("buildHash", BuildHash).
		Str("buildDate", BuildDate).
		Int("GOMAXPROCS", runtime.GOMAXPROCS(-1)).
		Msg(appName)

	promMetrics := prom.NewMetrics(*metricsPrefix)

	grpcServer, err := grpcserver.StartGrpcServer(*grpcPort, promMetrics)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start grpc server")

		return
	}

	metricsServer := restserver.StartServer(*restPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals.Handle(true, func() {
		restserver.StopServer(ctx, metricsServer)
		grpcserver.StopGrpcServer(grpcServer)

		cancel()
	})
}
