package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Lameaux/bro/internal/app"
	"github.com/Lameaux/bro/internal/config"
	"github.com/Lameaux/bro/internal/metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

const (
	logo = `
 _               
 | |              
 | |__  _ __ ___  
 | '_ \| '__/ _ \ 
 | |_) | | | (_) |
 |_.__/|_|  \___/ 
`
	appName    = "bro"
	appVersion = "v0.0.1"
)

var GitHash string

func main() {
	var debug = flag.Bool("debug", false, "enable debug mode")
	var showBanner = flag.Bool("banner", true, "show banner")
	var showStats = flag.Bool("stats", true, "show stats")
	var configFile = flag.String("config", "", "path to yaml file")
	var metricsPort = flag.String("metricsPort", "9090", "port for metrics server")

	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if *showBanner {
		fmt.Print(logo)
	}

	log.Info().Str("version", appVersion).Str("build", GitHash).Msg(appName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := metrics.StartServer(*metricsPort)
	defer metrics.StopServer(ctx, server)

	handleSignals(func() {
		metrics.StopServer(ctx, server)
		cancel()
	})

	conf := loadConfig(*configFile)

	app.Run(ctx, conf, *showStats)
}

func loadConfig(configFile string) *config.Config {
	if configFile == "" {
		log.Fatal().Msgf("--config parameter is missing. Example: %s --config myconfig.yaml", appName)
	}

	c, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading config from file")
	}

	log.Debug().Str("configName", c.Name).Msgf("config loaded")

	return c
}

func handleSignals(shutdownFn func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msgf("received signal")
		shutdownFn()
	}()
}
