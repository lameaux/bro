package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/lameaux/bro/internal/server/grpcserver"
	"github.com/lameaux/bro/internal/server/restserver"
	"github.com/lameaux/bro/internal/shared/banner"
	"github.com/lameaux/bro/internal/shared/signals"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	appName    = "brod"
	appVersion = "v0.0.1"

	defaultPortGrpc = 8080
	defaultPortRest = 9090
)

var GitHash string //nolint:gochecknoglobals

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	logJSON := flag.Bool("logJson", false, "log as json")
	skipBanner := flag.Bool("skipBanner", false, "skip banner")
	grpcPort := flag.Int("grpcPort", defaultPortGrpc, "port for grpc server")
	restPort := flag.Int("restPort", defaultPortRest, "port for rest server")

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

	log.Info().Str("version", appVersion).Str("build", GitHash).Msg(appName)

	grpcServer, err := grpcserver.StartGrpcServer(*grpcPort)
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
