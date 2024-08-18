package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/lameaux/bro/internal/app"
	"github.com/lameaux/bro/internal/config"
	"github.com/lameaux/bro/internal/metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	logo = `
 █████                       
░░███                        
 ░███████  ████████   ██████ 
 ░███░░███░░███░░███ ███░░███
 ░███ ░███ ░███ ░░░ ░███ ░███
 ░███ ░███ ░███     ░███ ░███
 ████████  █████    ░░██████ 
░░░░░░░░  ░░░░░      ░░░░░░  

`
	appName    = "bro"
	appVersion = "v0.0.1"
)

var GitHash string

func main() {
	var debug = flag.Bool("debug", false, "enable debug mode")
	var silent = flag.Bool("silent", false, "enable silent mode")
	var skipBanner = flag.Bool("skipBanner", false, "skip banner")
	var skipResults = flag.Bool("skipResults", false, "skip results")
	var metricsPort = flag.String("metricsPort", "", "port for metrics server")

	flag.Parse()

	if *silent {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if !*silent && !*skipBanner {
		fmt.Print(logo)
	}

	log.Info().Str("version", appVersion).Str("build", GitHash).Msg(appName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var metricsServer *http.Server
	if *metricsPort != "" {
		metricsServer := metrics.StartServer(*metricsPort)
		defer metrics.StopServer(ctx, metricsServer)
	}

	handleSignals(func() {
		if *metricsPort != "" {
			metrics.StopServer(ctx, metricsServer)
		}

		cancel()
	})

	conf := loadConfig(flag.Args())

	app.Run(ctx, conf, !*skipResults)
}

func loadConfig(args []string) *config.Config {
	if len(args) == 0 {
		log.Fatal().Msgf("config location is missing. Example: %s [flags] <config.yaml>", appName)
	}

	configFile := args[0]

	c, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading config from file")
	}

	log.Info().Str("configName", c.Name).Str("configFile", configFile).Msgf("config loaded")

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
